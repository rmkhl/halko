package router

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/rmkhl/halko/sensorunit/serial"
	"github.com/rmkhl/halko/types"
	"github.com/rmkhl/halko/types/log"
)

func writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	log.Debug("HTTP Response: %d", statusCode)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Error("Failed to encode JSON response: %v", err)
		_ = err
	}
}

func writeError(w http.ResponseWriter, statusCode int, message string) {
	log.Debug("HTTP Error Response: %d - %s", statusCode, message)
	writeJSON(w, statusCode, types.APIErrorResponse{Err: message})
}

func (api *API) getTemperatures(w http.ResponseWriter, r *http.Request) {
	log.Debug("Processing temperature request from %s", r.RemoteAddr)

	// Attempt to get temperatures, with retry if all readings are invalid
	var temperatures []serial.Temperature
	var err error
	const maxAttempts = 2

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		temperatures, err = api.sensorUnit.GetTemperatures()
		if err != nil {
			log.Error("Failed to get temperatures from sensor unit (attempt %d/%d): %v", attempt, maxAttempts, err)
			if attempt == maxAttempts {
				writeError(w, http.StatusInternalServerError, err.Error())
				return
			}
			log.Warning("Retrying temperature read in 500ms...")
			time.Sleep(500 * time.Millisecond)
			continue
		}
		log.Debug("Retrieved %d temperature readings from sensor unit (attempt %d/%d)", len(temperatures), attempt, maxAttempts)

		response := make(types.TemperatureResponse)
		var kilnPrimary float32
		var kilnSecondary float32

		for _, temp := range temperatures {
			switch temp.Name {
			case "KilnPrimary":
				kilnPrimary = temp.Value
			case "KilnSecondary":
				kilnSecondary = temp.Value
			case "Wood":
				response["material"] = temp.Value
			}
		}
		log.Debug("Temperature readings processed (attempt %d/%d): KilnPrimary=%.2f°C, KilnSecondary=%.2f°C, Material=%.2f°C",
			attempt, maxAttempts, kilnPrimary, kilnSecondary, response["material"])

		// Check if all readings are invalid
		allInvalid := (kilnPrimary == types.InvalidTemperatureReading &&
			kilnSecondary == types.InvalidTemperatureReading &&
			response["material"] == types.InvalidTemperatureReading)

		if allInvalid && attempt < maxAttempts {
			log.Warning("All temperature readings are invalid on attempt %d/%d, retrying in 500ms...", attempt, maxAttempts)
			time.Sleep(500 * time.Millisecond)
			continue
		}

		// Process temperature readings (either some are valid or this is the final attempt)
		switch {
		case kilnPrimary != types.InvalidTemperatureReading && kilnSecondary != types.InvalidTemperatureReading:
			api.updateKilnStatus(kilnSensorBothOK)
		case kilnPrimary != types.InvalidTemperatureReading:
			api.updateKilnStatus(kilnSensorPrimaryOnly)
		case kilnSecondary != types.InvalidTemperatureReading:
			api.updateKilnStatus(kilnSensorSecondaryOnly)
		default:
			api.updateKilnStatus(kilnSensorBothInvalid)
		}
		response["kiln"] = api.selectKilnTemperature(kilnPrimary, kilnSecondary)
		api.updateMaterialStatus(response["material"] != types.InvalidTemperatureReading)

		log.Debug("Temperature selection complete: kiln=%.1f°C, material=%.1f°C",
			response["kiln"], response["material"])

		log.Debug("Returning temperature data: kiln=%.1f°C, material=%.1f°C", response["kiln"], response["material"])
		writeJSON(w, http.StatusOK, types.APIResponse[types.TemperatureResponse]{
			Data: response,
		})
		return
	}
}

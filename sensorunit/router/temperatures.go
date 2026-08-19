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
		// A reading the device did not report must not read as 0 degrees.
		kilnPrimary := float32(types.InvalidTemperatureReading)
		kilnSecondary := float32(types.InvalidTemperatureReading)

		// Cold junction values ride along on the same read but stay out of
		// this response: they are diagnostics about the measurement, not
		// temperatures the system controls on. They are kept per chip rather
		// than folded into one value because whether they moved together is
		// what says a shift is the board and not the kiln.
		dies := make(types.TemperatureResponse, dieSensorCount)

		for _, temp := range temperatures {
			switch temp.Name {
			case "KilnPrimary":
				kilnPrimary = temp.Value
			case "KilnSecondary":
				kilnSecondary = temp.Value
			case "Wood":
				response["material"] = temp.Value
			case "KilnPrimaryDie":
				dies["kiln_primary_die"] = temp.Value
			case "KilnSecondaryDie":
				dies["kiln_secondary_die"] = temp.Value
			case "WoodDie":
				dies["material_die"] = temp.Value
			}
		}
		api.storeDieReadings(dies)

		log.Debug("Temperature readings processed (attempt %d/%d): KilnPrimary=%.2f°C, KilnSecondary=%.2f°C, Material=%.2f°C, dies %.2f/%.2f/%.2f°C",
			attempt, maxAttempts, kilnPrimary, kilnSecondary, response["material"],
			dies["kiln_primary_die"], dies["kiln_secondary_die"], dies["material_die"])

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

// dieSensorCount is how many cold junctions the unit reports, one per chip.
const dieSensorCount = 3

// getDieTemperatures serves the cold junction readings recorded by the last
// temperature read. It deliberately does not trigger a read of its own: the
// values move with the sensor board, not the kiln, so a poll-old value says
// the same thing as a fresh one and the serial link stays free for the
// readings the run depends on.
func (api *API) getDieTemperatures(w http.ResponseWriter, r *http.Request) {
	log.Debug("Processing die temperature request from %s", r.RemoteAddr)

	writeJSON(w, http.StatusOK, types.APIResponse[types.TemperatureResponse]{
		Data: api.dieReadings(),
	})
}

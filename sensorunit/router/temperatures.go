package router

import (
	"encoding/json"
	"net/http"

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
	temperatures, err := api.sensorUnit.GetTemperatures()
	if err != nil {
		log.Error("Failed to get temperatures from sensor unit: %v", err)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	log.Debug("Retrieved %d temperature readings from sensor unit", len(temperatures))

	response := make(types.TemperatureResponse)

	var ovenPrimary float32
	var ovenSecondary float32

	for _, temp := range temperatures {
		switch temp.Name {
		case "OvenPrimary":
			ovenPrimary = temp.Value
		case "OvenSecondary":
			ovenSecondary = temp.Value
		case "Wood":
			response["material"] = temp.Value
		}
	}
	log.Debug("Temperature readings processed: OvenPrimary=%.2f°C, OvenSecondary=%.2f°C, Material=%.2f°C",
		ovenPrimary, ovenSecondary, response["material"])
	// in case the primary or secondary temperature is not available we only use the other one
	// if both are available we use the higher one
	var selectedOvenTemp string
	switch {
	case ovenPrimary != types.InvalidTemperatureReading && ovenSecondary != types.InvalidTemperatureReading:
		if ovenPrimary > ovenSecondary {
			response["oven"] = ovenPrimary
			selectedOvenTemp = "primary (higher)"
		} else {
			response["oven"] = ovenSecondary
			selectedOvenTemp = "secondary (higher)"
		}
	case ovenPrimary != types.InvalidTemperatureReading:
		log.Warning("Secondary oven temperature reading is invalid, using primary only")
		response["oven"] = ovenPrimary
		selectedOvenTemp = "primary only"
	case ovenSecondary != types.InvalidTemperatureReading:
		log.Warning("Primary oven temperature reading is invalid, using secondary only")
		response["oven"] = ovenSecondary
		selectedOvenTemp = "secondary only"
	default:
		log.Warning("Both oven temperature readings are invalid")
		response["oven"] = types.InvalidTemperatureReading
		selectedOvenTemp = "invalid"
	}
	if response["material"] == types.InvalidTemperatureReading {
		log.Warning("Material temperature reading is invalid")
	}

	log.Debug("Temperature selection complete: oven=%.1f°C (%s), material=%.1f°C",
		response["oven"], selectedOvenTemp, response["material"])

	log.Debug("Returning temperature data: oven=%.1f°C, material=%.1f°C", response["oven"], response["material"])
	writeJSON(w, http.StatusOK, types.APIResponse[types.TemperatureResponse]{
		Data: response,
	})
}

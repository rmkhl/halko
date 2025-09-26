package router

import (
	"encoding/json"
	"net/http"

	"github.com/rmkhl/halko/types"
	"github.com/rmkhl/halko/types/log"
)

func writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	log.Trace("Writing JSON response with status code %d", statusCode)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Trace("Failed to encode JSON response: %v", err)
		_ = err
	} else {
		log.Trace("JSON response written successfully")
	}
}

func writeError(w http.ResponseWriter, statusCode int, message string) {
	log.Trace("Writing error response: status=%d, message=%q", statusCode, message)
	writeJSON(w, statusCode, types.APIErrorResponse{Err: message})
}

func (api *API) getTemperatures(w http.ResponseWriter, _ *http.Request) {
	log.Trace("Handling GET temperatures request")
	log.Trace("Getting temperatures from sensor unit")
	temperatures, err := api.sensorUnit.GetTemperatures()
	if err != nil {
		log.Trace("Failed to get temperatures: %v", err)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	log.Trace("Retrieved %d temperature readings", len(temperatures))

	response := make(types.TemperatureResponse)

	var ovenPrimary float32
	var ovenSecondary float32

	for _, temp := range temperatures {
		log.Info("Temperature %s: %.2f", temp.Name, temp.Value)
		log.Trace("Processing temperature reading: %s = %.2f", temp.Name, temp.Value)
		switch temp.Name {
		case "OvenPrimary":
			ovenPrimary = temp.Value
			log.Trace("Set oven primary temperature: %.2f", temp.Value)
		case "OvenSecondary":
			ovenSecondary = temp.Value
			log.Trace("Set oven secondary temperature: %.2f", temp.Value)
		case "Wood":
			response["material"] = temp.Value
			log.Trace("Set material temperature: %.2f", temp.Value)
		}
	}
	// in case the primary or secondary temperature is not available we only use the other one
	// if both are available we use the higher one
	log.Trace("Processing oven temperature logic: primary=%.2f, secondary=%.2f", ovenPrimary, ovenSecondary)
	switch {
	case ovenPrimary != types.InvalidTemperatureReading && ovenSecondary != types.InvalidTemperatureReading:
		if ovenPrimary > ovenSecondary {
			response["oven"] = ovenPrimary
			log.Trace("Using primary oven temperature (higher): %.2f", ovenPrimary)
		} else {
			response["oven"] = ovenSecondary
			log.Trace("Using secondary oven temperature (higher): %.2f", ovenSecondary)
		}
	case ovenPrimary != types.InvalidTemperatureReading:
		log.Info("Secondary oven temperature reading is invalid.")
		response["oven"] = ovenPrimary
		log.Trace("Using only primary oven temperature: %.2f", ovenPrimary)
	case ovenSecondary != types.InvalidTemperatureReading:
		log.Info("Primary oven temperature reading is invalid.")
		response["oven"] = ovenSecondary
		log.Trace("Using only secondary oven temperature: %.2f", ovenSecondary)
	default:
		log.Info("Oven temperature reading is invalid.")
		response["oven"] = types.InvalidTemperatureReading
		log.Trace("Both oven temperatures invalid, setting to invalid reading")
	}
	if response["material"] == types.InvalidTemperatureReading {
		log.Info("Wood temperature reading is invalid.")
		log.Trace("Material temperature is invalid")
	} else {
		log.Trace("Material temperature is valid: %.2f", response["material"])
	}

	log.Trace("Sending temperature response with %d values", len(response))
	writeJSON(w, http.StatusOK, types.APIResponse[types.TemperatureResponse]{
		Data: response,
	})
}

package router

import (
	"encoding/json"
	"net/http"

	"github.com/rmkhl/halko/types"
	"github.com/rmkhl/halko/types/log"
)

func writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		_ = err
	}
}

func writeError(w http.ResponseWriter, statusCode int, message string) {
	writeJSON(w, statusCode, types.APIErrorResponse{Err: message})
}

func (api *API) getTemperatures(w http.ResponseWriter, _ *http.Request) {
	temperatures, err := api.sensorUnit.GetTemperatures()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := make(types.TemperatureResponse)

	var ovenPrimary float32
	var ovenSecondary float32

	for _, temp := range temperatures {
		log.Info("Temperature %s: %.2f", temp.Name, temp.Value)
		switch temp.Name {
		case "OvenPrimary":
			ovenPrimary = temp.Value
		case "OvenSecondary":
			ovenSecondary = temp.Value
		case "Wood":
			response["material"] = temp.Value
		}
	}
	// in case the primary or secondary temperature is not available we only use the other one
	// if both are available we use the higher one
	switch {
	case ovenPrimary != types.InvalidTemperatureReading && ovenSecondary != types.InvalidTemperatureReading:
		if ovenPrimary > ovenSecondary {
			response["oven"] = ovenPrimary
		} else {
			response["oven"] = ovenSecondary
		}
	case ovenPrimary != types.InvalidTemperatureReading:
		log.Info("Secondary oven temperature reading is invalid.")
		response["oven"] = ovenPrimary
	case ovenSecondary != types.InvalidTemperatureReading:
		log.Info("Primary oven temperature reading is invalid.")
		response["oven"] = ovenSecondary
	default:
		log.Info("Oven temperature reading is invalid.")
		response["oven"] = types.InvalidTemperatureReading
	}
	if response["material"] == types.InvalidTemperatureReading {
		log.Info("Wood temperature reading is invalid.")
	}

	writeJSON(w, http.StatusOK, types.APIResponse[types.TemperatureResponse]{
		Data: response,
	})
}

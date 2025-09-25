package router

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/rmkhl/halko/types"
)

// writeJSON writes a JSON response
func writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// writeError writes an error response
func writeError(w http.ResponseWriter, statusCode int, message string) {
	writeJSON(w, statusCode, types.APIErrorResponse{Err: message})
}

// getTemperatures handles GET requests to fetch temperature data
func (api *API) getTemperatures(w http.ResponseWriter, r *http.Request) {
	temperatures, err := api.sensorUnit.GetTemperatures()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := make(types.TemperatureResponse)

	// Map the sensor values to the expected keys
	var ovenPrimary float32
	var ovenSecondary float32

	for _, temp := range temperatures {
		log.Printf("Temperature %s: %.2f", temp.Name, temp.Value)
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
		log.Println("Secondary oven temperature reading is invalid.")
		response["oven"] = ovenPrimary
	case ovenSecondary != types.InvalidTemperatureReading:
		log.Println("Primary oven temperature reading is invalid.")
		response["oven"] = ovenSecondary
	default:
		log.Println("Oven temperature reading is invalid.")
		response["oven"] = types.InvalidTemperatureReading
	}
	if response["material"] == types.InvalidTemperatureReading {
		log.Println("Wood temperature reading is invalid.")
	}

	writeJSON(w, http.StatusOK, types.APIResponse[types.TemperatureResponse]{
		Data: response,
	})
}

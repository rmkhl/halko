package router

import (
	"encoding/json"
	"net/http"

	"github.com/rmkhl/halko/powerunit/power"
	"github.com/rmkhl/halko/powerunit/shelly"
	"github.com/rmkhl/halko/types"
)

// writeJSON writes a JSON response
func writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		// Log the error but don't change the response as headers are already sent
		_ = err
	}
}

// writeError writes an error response
func writeError(w http.ResponseWriter, statusCode int, message string) {
	writeJSON(w, statusCode, types.APIErrorResponse{Err: message})
}

func getAllPercentages(p *power.Controller, idMapping [shelly.NumberOfDevices]string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		percentages := p.GetAllPercentages()

		response := make(types.PowerStatusResponse)
		for id := range shelly.NumberOfDevices {
			response[idMapping[id]] = types.PowerResponse{Percent: percentages[id]}
		}

		writeJSON(w, http.StatusOK, types.APIResponse[types.PowerStatusResponse]{Data: response})
	}
}

func setAllPercentages(p *power.Controller, powerMapping map[string]int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var commands types.PowersCommand

		err := json.NewDecoder(r.Body).Decode(&commands)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		var percentages [shelly.NumberOfDevices]uint8

		for powerName, command := range commands {
			id, ok := powerMapping[powerName]
			if !ok {
				writeError(w, http.StatusBadRequest, "Unknown power '"+powerName+"'")
				return
			}
			percentages[id] = command.Percent
		}

		p.SetAllPercentages(percentages)

		writeJSON(w, http.StatusOK, types.APIResponse[types.PowerOperationResponse]{
			Data: types.PowerOperationResponse{Message: "completed"},
		})
	}
}

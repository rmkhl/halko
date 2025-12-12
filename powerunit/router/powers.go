package router

import (
	"encoding/json"
	"net/http"

	"github.com/rmkhl/halko/powerunit/power"
	"github.com/rmkhl/halko/powerunit/shelly"
	"github.com/rmkhl/halko/types"
	"github.com/rmkhl/halko/types/log"
)

func writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Error("Failed to encode JSON response: %v", err)
	}
}

func writeError(w http.ResponseWriter, statusCode int, message string) {
	log.Warning("API error response: status=%d, message=%s", statusCode, message)
	writeJSON(w, statusCode, types.APIErrorResponse{Err: message})
}

func getAllPercentages(p *power.Controller, idMapping [shelly.NumberOfDevices]string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Trace("GET /power request from %s", r.RemoteAddr)
		percentages := p.GetAllPercentages()

		response := make(types.PowerStatusResponse)
		for id := range shelly.NumberOfDevices {
			response[idMapping[id]] = types.PowerResponse{Percent: percentages[id]}
		}
		log.Debug("Returning power status: %v", response)

		writeJSON(w, http.StatusOK, types.APIResponse[types.PowerStatusResponse]{Data: response})
	}
}

func setAllPercentages(p *power.Controller, powerMapping map[string]int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Trace("POST /power request from %s", r.RemoteAddr)
		var commands types.PowersCommand

		err := json.NewDecoder(r.Body).Decode(&commands)
		if err != nil {
			log.Warning("Invalid JSON in power command request: %v", err)
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		log.Debug("Received power commands: %v", commands)

		var percentages [shelly.NumberOfDevices]uint8

		for powerName, command := range commands {
			id, ok := powerMapping[powerName]
			if !ok {
				log.Warning("Unknown power device requested: %s", powerName)
				writeError(w, http.StatusBadRequest, "Unknown power '"+powerName+"'")
				return
			}
			percentages[id] = command.Percent
			log.Trace("Mapped power %s (id=%d) to %d%%", powerName, id, command.Percent)
		}

		p.SetAllPercentages(percentages)
		log.Info("Power percentages updated successfully")

		writeJSON(w, http.StatusOK, types.APIResponse[types.PowerOperationResponse]{
			Data: types.PowerOperationResponse{Message: "completed"},
		})
	}
}

func getPercentage(p *power.Controller, powerMapping map[string]int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		powerName := r.PathValue("power")
		log.Trace("GET /power/%s request from %s", powerName, r.RemoteAddr)

		id, ok := powerMapping[powerName]
		if !ok {
			log.Warning("Unknown power device requested: %s", powerName)
			writeError(w, http.StatusNotFound, "Unknown power '"+powerName+"'")
			return
		}

		percentages := p.GetAllPercentages()
		log.Debug("Returning power status for %s: %d%%", powerName, percentages[id])

		writeJSON(w, http.StatusOK, types.APIResponse[types.PowerResponse]{
			Data: types.PowerResponse{Percent: percentages[id]},
		})
	}
}

func setPercentage(p *power.Controller, powerMapping map[string]int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		powerName := r.PathValue("power")
		log.Trace("POST /power/%s request from %s", powerName, r.RemoteAddr)

		id, ok := powerMapping[powerName]
		if !ok {
			log.Warning("Unknown power device requested: %s", powerName)
			writeError(w, http.StatusNotFound, "Unknown power '"+powerName+"'")
			return
		}

		var command types.PowerCommand
		err := json.NewDecoder(r.Body).Decode(&command)
		if err != nil {
			log.Warning("Invalid JSON in power command request: %v", err)
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		log.Debug("Received power command for %s: %d%%", powerName, command.Percent)

		percentages := p.GetAllPercentages()
		percentages[id] = command.Percent

		p.SetAllPercentages(percentages)
		log.Info("Power percentage for %s updated to %d%%", powerName, command.Percent)

		writeJSON(w, http.StatusOK, types.APIResponse[types.PowerResponse]{
			Data: types.PowerResponse(command),
		})
	}
}

package router

import (
	"encoding/json"
	"net/http"

	"github.com/rmkhl/halko/types"
	"github.com/rmkhl/halko/types/log"
)

func (r *Router) setDisplay(w http.ResponseWriter, req *http.Request) {
	log.Trace("Processing display update request from %s", req.RemoteAddr)
	var payload types.DisplayRequest
	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		log.Error("Failed to decode display request body: %v", err)
		writeError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	log.Info("Simulator received display update: %s", payload.Message)

	// Reset simulation when receiving "idle" status (only if transitioning from non-idle)
	if payload.Message == "idle" && r.Resetter != nil {
		r.Resetter.Mutex.Lock()
		shouldReset := r.Resetter.LastMessage != "idle"
		r.Resetter.LastMessage = payload.Message
		r.Resetter.Mutex.Unlock()

		if shouldReset {
			r.Resetter.Mutex.Lock()
			defer r.Resetter.Mutex.Unlock()

			log.Info("Resetting simulation to initial state (Oven: %.1f°C, Material: %.1f°C, Environment: %.1f°C)",
				r.Resetter.InitialOvenTemp, r.Resetter.InitialMaterialTemp, r.Resetter.EnvironmentTemp)

			// Reset element temperatures
			r.Resetter.Heater.SetTemperature(r.Resetter.InitialOvenTemp)
			r.Resetter.Wood.SetTemperature(r.Resetter.InitialMaterialTemp)

			// Turn off all power elements
			r.Resetter.Heater.TurnOn(false)
			r.Resetter.Fan.TurnOn(false)
			r.Resetter.Humidifier.TurnOn(false)

			// Reset physics state
			r.Resetter.PhysicsState.OvenTemp = r.Resetter.InitialOvenTemp
			r.Resetter.PhysicsState.MaterialTemp = r.Resetter.InitialMaterialTemp
			r.Resetter.PhysicsState.EnvironmentTemp = r.Resetter.EnvironmentTemp
			r.Resetter.PhysicsState.HeaterIsOn = false
			r.Resetter.PhysicsState.FanIsOn = false
			r.Resetter.PhysicsState.HumidifierIsOn = false

			log.Info("Simulation reset complete")
		}
	} else if r.Resetter != nil {
		// Update last message tracker for non-idle messages
		r.Resetter.Mutex.Lock()
		r.Resetter.LastMessage = payload.Message
		r.Resetter.Mutex.Unlock()
	}

	response := types.APIResponse[types.StatusResponse]{
		Data: types.StatusResponse{
			Status: types.SensorStatusOK,
		},
	}

	writeJSON(w, http.StatusOK, response)
}

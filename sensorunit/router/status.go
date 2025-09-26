package router

import (
	"encoding/json"
	"net/http"

	"github.com/rmkhl/halko/types"
	"github.com/rmkhl/halko/types/log"
)

// setStatus handles POST requests to update the status text on the LCD
// This function is now part of the API struct and called by SetupRoutes.
// No longer a standalone setupStatusRoutes function.
func (api *API) setStatus(w http.ResponseWriter, r *http.Request) {
	log.Debug("Processing status update request from %s", r.RemoteAddr)
	var payload types.StatusRequest
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		log.Error("Failed to decode status request body: %v", err)
		writeError(w, http.StatusBadRequest, "Invalid request format")
		return
	}
	log.Debug("Status update request received: %q", payload.Message)

	err = api.sensorUnit.SetStatusText(payload.Message)
	if err != nil {
		log.Error("Failed to set status text on sensor unit: %v", err)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	log.Info("LCD status updated: %q", payload.Message)
	writeJSON(w, http.StatusOK, types.APIResponse[types.StatusResponse]{
		Data: types.StatusResponse{
			Status: types.SensorStatusOK,
		},
	})
}

func (api *API) getStatus(w http.ResponseWriter, r *http.Request) {
	log.Debug("Processing status check request from %s", r.RemoteAddr)
	isConnected := api.sensorUnit.IsConnected()
	log.Debug("Sensor unit connection check result: %t", isConnected)

	status := types.SensorStatusConnected
	if !isConnected {
		status = types.SensorStatusDisconnected
	}
	log.Debug("Returning sensor status: %s", status)

	writeJSON(w, http.StatusOK, types.APIResponse[types.StatusResponse]{
		Data: types.StatusResponse{
			Status: status,
		},
	})
}

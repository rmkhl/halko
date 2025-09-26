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
	log.Trace("Handling POST status request")
	var payload types.StatusRequest
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		log.Trace("Failed to decode status request body: %v", err)
		writeError(w, http.StatusBadRequest, "Invalid request format")
		return
	}
	log.Trace("Status message received: %q", payload.Message)

	log.Trace("Setting status text on sensor unit")
	err = api.sensorUnit.SetStatusText(payload.Message)
	if err != nil {
		log.Trace("Failed to set status text: %v", err)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	log.Trace("Status text set successfully, returning response")
	writeJSON(w, http.StatusOK, types.APIResponse[types.StatusResponse]{
		Data: types.StatusResponse{
			Status: types.SensorStatusOK,
		},
	})
}

func (api *API) getStatus(w http.ResponseWriter, _ *http.Request) {
	log.Trace("Handling GET status request")
	log.Trace("Checking sensor unit connection status")
	isConnected := api.sensorUnit.IsConnected()
	log.Trace("Sensor unit connection status: %t", isConnected)

	status := types.SensorStatusConnected
	if !isConnected {
		status = types.SensorStatusDisconnected
	}
	log.Trace("Returning status: %s", status)

	writeJSON(w, http.StatusOK, types.APIResponse[types.StatusResponse]{
		Data: types.StatusResponse{
			Status: status,
		},
	})
}

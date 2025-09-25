package router

import (
	"encoding/json"
	"net/http"

	"github.com/rmkhl/halko/types"
)

// setStatus handles POST requests to update the status text on the LCD
// This function is now part of the API struct and called by SetupRoutes.
// No longer a standalone setupStatusRoutes function.
func (api *API) setStatus(w http.ResponseWriter, r *http.Request) {
	var payload types.StatusRequest
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	err = api.sensorUnit.SetStatusText(payload.Message)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, types.APIResponse[types.StatusResponse]{
		Data: types.StatusResponse{
			Status: types.SensorStatusOK,
		},
	})
}

// getStatus handles GET requests to check the connection status
func (api *API) getStatus(w http.ResponseWriter, _ *http.Request) {
	isConnected := api.sensorUnit.IsConnected()

	status := types.SensorStatusConnected
	if !isConnected {
		status = types.SensorStatusDisconnected
	}

	writeJSON(w, http.StatusOK, types.APIResponse[types.StatusResponse]{
		Data: types.StatusResponse{
			Status: status,
		},
	})
}

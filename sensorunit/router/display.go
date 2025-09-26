package router

import (
	"encoding/json"
	"net/http"

	"github.com/rmkhl/halko/types"
	"github.com/rmkhl/halko/types/log"
)

// setDisplay handles POST requests to update the display text on the LCD
func (api *API) setDisplay(w http.ResponseWriter, r *http.Request) {
	log.Debug("Processing display update request from %s", r.RemoteAddr)
	var payload types.DisplayRequest
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		log.Error("Failed to decode display request body: %v", err)
		writeError(w, http.StatusBadRequest, "Invalid request format")
		return
	}
	log.Debug("Display update request received: %q", payload.Message)

	err = api.sensorUnit.SetStatusText(payload.Message)
	if err != nil {
		log.Error("Failed to set display text on sensor unit: %v", err)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, types.APIResponse[types.StatusResponse]{
		Data: types.StatusResponse{
			Status: types.SensorStatusOK,
		},
	})
}

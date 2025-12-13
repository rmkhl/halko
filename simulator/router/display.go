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

	response := types.APIResponse[types.StatusResponse]{
		Data: types.StatusResponse{
			Status: types.SensorStatusOK,
		},
	}

	writeJSON(w, http.StatusOK, response)
}

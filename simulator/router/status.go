package router

import (
	"net/http"

	"github.com/rmkhl/halko/types"
	"github.com/rmkhl/halko/types/log"
)

func (r *Router) getStatus(w http.ResponseWriter, req *http.Request) {
	log.Debug("Processing status check request from %s", req.RemoteAddr)
	log.Debug("Returning sensor status: %s", types.SensorStatusConnected)

	response := types.APIResponse[types.StatusResponse]{
		Data: types.StatusResponse{
			Status: types.SensorStatusConnected,
		},
	}

	writeJSON(w, http.StatusOK, response)
}

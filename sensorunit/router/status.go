package router

import (
	"net/http"

	"github.com/rmkhl/halko/types"
	"github.com/rmkhl/halko/types/log"
)

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

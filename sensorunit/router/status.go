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

	status := types.ServiceStatusHealthy
	if !isConnected {
		status = types.ServiceStatusUnavailable
	}

	details := make(map[string]interface{})
	details["arduino_connected"] = isConnected

	response := types.ServiceStatusResponse{
		Status:  status,
		Service: "sensorunit",
		Details: details,
	}

	log.Debug("Returning sensor status: %s", status)

	writeJSON(w, http.StatusOK, types.APIResponse[types.ServiceStatusResponse]{
		Data: response,
	})
}

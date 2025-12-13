package router

import (
	"net/http"

	"github.com/rmkhl/halko/types"
	"github.com/rmkhl/halko/types/log"
)

func (r *Router) getStatus(w http.ResponseWriter, req *http.Request) {
	log.Trace("Processing status check request from %s", req.RemoteAddr)

	details := make(map[string]interface{})
	details["arduino_connected"] = true

	response := types.ServiceStatusResponse{
		Status:  types.ServiceStatusHealthy,
		Service: "sensorunit",
		Details: details,
	}

	log.Trace("Returning sensor status: %s", response.Status)

	writeJSON(w, http.StatusOK, types.APIResponse[types.ServiceStatusResponse]{
		Data: response,
	})
}

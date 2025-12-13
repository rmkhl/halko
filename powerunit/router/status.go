package router

import (
	"net/http"

	"github.com/rmkhl/halko/powerunit/power"
	"github.com/rmkhl/halko/types"
)

func getStatus(p *power.Controller) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		details := make(map[string]interface{})

		// Check if controller is initialized
		isHealthy := p != nil
		status := types.ServiceStatusHealthy
		if !isHealthy {
			status = types.ServiceStatusUnavailable
		}

		details["controller_initialized"] = isHealthy
		if isHealthy {
			details["is_idle"] = p.IsIdle()
		}

		response := types.ServiceStatusResponse{
			Status:  status,
			Service: "powerunit",
			Details: details,
		}

		writeJSON(w, http.StatusOK, types.APIResponse[types.ServiceStatusResponse]{Data: response})
	}
}

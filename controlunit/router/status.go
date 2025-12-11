package router

import (
	"net/http"

	"github.com/rmkhl/halko/controlunit/engine"
	"github.com/rmkhl/halko/types"
)

func getStatus(eng *engine.ControlEngine) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		details := make(map[string]interface{})

		currentStatus := eng.CurrentStatus()
		if currentStatus != nil {
			details["program_running"] = true
			details["current_step"] = currentStatus.CurrentStep
			details["started_at"] = currentStatus.StartedAt
		} else {
			details["program_running"] = false
		}

		response := types.ServiceStatusResponse{
			Status:  types.ServiceStatusHealthy,
			Service: "executor",
			Details: details,
		}

		writeJSON(w, http.StatusOK, types.APIResponse[types.ServiceStatusResponse]{Data: response})
	}
}

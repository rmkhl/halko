package router

import (
	"net/http"

	"github.com/rmkhl/halko/types"
)

func getStatus() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		response := types.ServiceStatusResponse{
			Status:  types.ServiceStatusHealthy,
			Service: "controlunit",
		}

		writeJSON(w, http.StatusOK, types.APIResponse[types.ServiceStatusResponse]{Data: response})
	}
}

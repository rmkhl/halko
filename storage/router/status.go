package router

import (
	"net/http"

	"github.com/rmkhl/halko/storage/storagefs"
	"github.com/rmkhl/halko/types"
)

func getStatus(storage *storagefs.ProgramStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		details := make(map[string]interface{})

		// Check if storage is accessible
		_, err := storage.ListStoredPrograms()
		status := types.ServiceStatusHealthy
		if err != nil {
			status = types.ServiceStatusDegraded
			details["error"] = err.Error()
			details["accessible"] = false
		} else {
			details["accessible"] = true
		}

		response := types.ServiceStatusResponse{
			Status:  status,
			Service: "storage",
			Details: details,
		}

		writeJSON(w, http.StatusOK, types.APIResponse[types.ServiceStatusResponse]{Data: response})
	}
}

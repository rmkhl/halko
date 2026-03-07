package router

import (
	"encoding/json"
	"net/http"

	"github.com/rmkhl/halko/dbusunit/dbus"
	"github.com/rmkhl/halko/types"
	"github.com/rmkhl/halko/types/log"
)

// getStatus returns a handler that provides service health status
func getStatus(manager *dbus.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Debug("Received request: GET /status")

		// Check if D-Bus connection is active
		connected := manager.IsConnected()

		status := types.ServiceStatusHealthy
		if !connected {
			status = types.ServiceStatusUnavailable
		}

		details := map[string]interface{}{
			"dbus_connected": connected,
		}

		response := types.ServiceStatusResponse{
			Status:  status,
			Service: "dbusunit",
			Details: details,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Error("Failed to encode status response: %v", err)
		}
	}
}

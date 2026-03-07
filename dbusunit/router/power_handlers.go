package router

import (
	"encoding/json"
	"net/http"

	"github.com/rmkhl/halko/dbusunit/dbus"
	"github.com/rmkhl/halko/types"
	"github.com/rmkhl/halko/types/log"
)

// PowerRequest represents the request body for power operations
type PowerRequest struct {
	Delay int `json:"delay"` // Delay in seconds (currently unused)
}

// shutdown returns a handler that initiates system shutdown
func shutdown(manager *dbus.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Debug("Received request: POST /power/shutdown")

		var req PowerRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			// If no body provided, use default delay of 0
			req.Delay = 0
		}

		if err := manager.Shutdown(req.Delay); err != nil {
			log.Error("Failed to shutdown system: %v", err)
			http.Error(w, `{"error": "failed to shutdown system"}`, http.StatusInternalServerError)
			return
		}

		response := types.APIResponse[map[string]string]{
			Data: map[string]string{
				"message": "System shutdown initiated",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Error("Failed to encode shutdown response: %v", err)
		}
	}
}

// reboot returns a handler that initiates system reboot
func reboot(manager *dbus.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Debug("Received request: POST /power/reboot")

		var req PowerRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			// If no body provided, use default delay of 0
			req.Delay = 0
		}

		if err := manager.Reboot(req.Delay); err != nil {
			log.Error("Failed to reboot system: %v", err)
			http.Error(w, `{"error": "failed to reboot system"}`, http.StatusInternalServerError)
			return
		}

		response := types.APIResponse[map[string]string]{
			Data: map[string]string{
				"message": "System reboot initiated",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Error("Failed to encode reboot response: %v", err)
		}
	}
}

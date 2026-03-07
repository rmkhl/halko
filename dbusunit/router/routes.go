package router

import (
	"net/http"

	"github.com/rmkhl/halko/dbusunit/dbus"
	"github.com/rmkhl/halko/types"
	"github.com/rmkhl/halko/types/log"
)

// corsMiddleware adds CORS headers to allow cross-origin requests from webapp
func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Allow requests from any origin (for development)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next(w, r)
	}
}

// SetupRoutes configures all HTTP routes for the dbusunit service
func SetupRoutes(mux *http.ServeMux, manager *dbus.Manager, endpoints *types.APIEndpoints) {
	// VPN endpoints
	mux.HandleFunc("GET "+endpoints.DBusUnit.VPN, corsMiddleware(listVPNs(manager)))
	mux.HandleFunc("GET "+endpoints.DBusUnit.VPN+"/{name}", corsMiddleware(getVPNStatus(manager)))
	mux.HandleFunc("POST "+endpoints.DBusUnit.VPN+"/{name}/start", corsMiddleware(startVPN(manager)))
	mux.HandleFunc("POST "+endpoints.DBusUnit.VPN+"/{name}/stop", corsMiddleware(stopVPN(manager)))

	// Power endpoints
	mux.HandleFunc("POST "+endpoints.DBusUnit.Power+"/shutdown", corsMiddleware(shutdown(manager)))
	mux.HandleFunc("POST "+endpoints.DBusUnit.Power+"/reboot", corsMiddleware(reboot(manager)))

	// Status endpoint
	mux.HandleFunc("GET "+endpoints.DBusUnit.Status, corsMiddleware(getStatus(manager)))

	log.Info("HTTP API initialized with 7 endpoints")
}

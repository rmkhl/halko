package router

import (
	"net/http"

	"github.com/rmkhl/halko/sensorunit/serial"
)

// API represents the REST API for the sensor unit
type API struct {
	sensorUnit *serial.SensorUnit
}

// NewAPI creates a new API instance
func NewAPI(sensorUnit *serial.SensorUnit) *API {
	return &API{
		sensorUnit: sensorUnit,
	}
}

// SetupRouter configures the HTTP router with the API routes
func SetupRouter(api *API) http.Handler {
	mux := http.NewServeMux()

	// Setup routes
	SetupRoutes(mux, api)

	// Add CORS middleware wrapper
	return addCORSHeaders(mux)
}

// addCORSHeaders wraps the mux with CORS headers
func addCORSHeaders(mux *http.ServeMux) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Forward to original mux
		mux.ServeHTTP(w, r)
	})
}

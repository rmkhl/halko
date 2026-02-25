package router

import (
	"net/http"

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

func SetupRoutes(mux *http.ServeMux, api *API, endpoints *types.APIEndpoints) {
	mux.HandleFunc("GET "+endpoints.SensorUnit.Temperatures, corsMiddleware(api.getTemperatures))
	mux.HandleFunc("GET "+endpoints.SensorUnit.Status, corsMiddleware(api.getStatus))
	mux.HandleFunc("POST "+endpoints.SensorUnit.Display, corsMiddleware(api.setDisplay))
	log.Info("HTTP API initialized with 3 endpoints: %s, %s, %s",
		endpoints.SensorUnit.Temperatures, endpoints.SensorUnit.Status, endpoints.SensorUnit.Display)
}

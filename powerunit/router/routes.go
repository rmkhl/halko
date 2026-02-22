package router

import (
	"net/http"

	"github.com/rmkhl/halko/powerunit/power"
	"github.com/rmkhl/halko/powerunit/shelly"
	"github.com/rmkhl/halko/types"
)

// corsMiddleware adds CORS headers to allow cross-origin requests from webapp
func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Allow requests from any origin (for development)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next(w, r)
	}
}

func setupRoutes(mux *http.ServeMux, p *power.Controller, powerMapping map[string]int, idMapping [shelly.NumberOfDevices]string, endpoints *types.APIEndpoints) {
	mux.HandleFunc("GET "+endpoints.PowerUnit.Power, corsMiddleware(getAllPercentages(p, idMapping)))
	mux.HandleFunc("POST "+endpoints.PowerUnit.Power, corsMiddleware(setAllPercentages(p, powerMapping)))
	mux.HandleFunc("GET "+endpoints.PowerUnit.Power+"/{power}", corsMiddleware(getPercentage(p, powerMapping)))
	mux.HandleFunc("POST "+endpoints.PowerUnit.Power+"/{power}", corsMiddleware(setPercentage(p, powerMapping)))
	mux.HandleFunc("PUT "+endpoints.PowerUnit.Power+"/{power}", corsMiddleware(setPercentage(p, powerMapping)))
	mux.HandleFunc("PATCH "+endpoints.PowerUnit.Power+"/{power}", corsMiddleware(setPercentage(p, powerMapping)))
	mux.HandleFunc("GET "+endpoints.PowerUnit.Status, corsMiddleware(getStatus(p)))
}

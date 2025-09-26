package router

import (
	"net/http"

	"github.com/rmkhl/halko/sensorunit/serial"
	"github.com/rmkhl/halko/types"
	"github.com/rmkhl/halko/types/log"
)

type API struct {
	sensorUnit *serial.SensorUnit
}

func NewAPI(sensorUnit *serial.SensorUnit) *API {
	log.Trace("Creating new API instance")
	return &API{
		sensorUnit: sensorUnit,
	}
}

func SetupRouter(api *API, endpoints *types.APIEndpoints) http.Handler {
	log.Trace("Setting up HTTP router")
	mux := http.NewServeMux()

	log.Trace("Setting up API routes")
	SetupRoutes(mux, api, endpoints)

	log.Trace("Adding CORS headers to router")
	return addCORSHeaders(mux)
}

func addCORSHeaders(mux *http.ServeMux) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Trace("Adding CORS headers for request %s %s", r.Method, r.URL.Path)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if r.Method == "OPTIONS" {
			log.Trace("Handling OPTIONS preflight request")
			w.WriteHeader(http.StatusNoContent)
			return
		}

		log.Trace("Forwarding request to main handler")
		mux.ServeHTTP(w, r)
	})
}

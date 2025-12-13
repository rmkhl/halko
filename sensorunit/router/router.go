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
	return &API{
		sensorUnit: sensorUnit,
	}
}

func SetupRouter(api *API, endpoints *types.APIEndpoints) http.Handler {
	mux := http.NewServeMux()
	SetupRoutes(mux, api, endpoints)
	return addCORSHeaders(mux)
}

func addCORSHeaders(mux *http.ServeMux) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Debug("HTTP Request: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if r.Method == "OPTIONS" {
			log.Debug("HTTP Response: OPTIONS %s -> 204 No Content", r.URL.Path)
			w.WriteHeader(http.StatusNoContent)
			return
		}
		mux.ServeHTTP(w, r)
	})
}

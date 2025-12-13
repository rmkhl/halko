package router

import (
	"net/http"

	"github.com/rmkhl/halko/powerunit/power"
	"github.com/rmkhl/halko/powerunit/shelly"
	"github.com/rmkhl/halko/types"
	"github.com/rmkhl/halko/types/log"
)

func New(p *power.Controller, powerMapping map[string]int, idMapping [shelly.NumberOfDevices]string, endpoints *types.APIEndpoints) http.Handler {
	log.Trace("Creating HTTP router")
	mux := http.NewServeMux()

	setupRoutes(mux, p, powerMapping, idMapping, endpoints)
	log.Debug("HTTP routes configured")

	handler := addCORSHeaders(mux)
	log.Debug("CORS headers configured")

	return handler
}

func addCORSHeaders(mux *http.ServeMux) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Length")
		w.Header().Set("Access-Control-Max-Age", "43200") // 12 hours

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		mux.ServeHTTP(w, r)
	})
}

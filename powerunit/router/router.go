package router

import (
	"net/http"

	"github.com/rmkhl/halko/powerunit/power"
	"github.com/rmkhl/halko/powerunit/shelly"
)

func New(p *power.Controller, powerMapping map[string]int, idMapping [shelly.NumberOfDevices]string) http.Handler {
	mux := http.NewServeMux()

	setupRoutes(mux, p, powerMapping, idMapping)

	// Add CORS middleware wrapper
	handler := addCORSHeaders(mux)

	return handler
}

// addCORSHeaders wraps the mux with CORS headers
func addCORSHeaders(mux *http.ServeMux) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:1234")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Length")
		w.Header().Set("Access-Control-Max-Age", "43200") // 12 hours

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Forward to original mux
		mux.ServeHTTP(w, r)
	})
}

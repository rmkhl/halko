package router

import (
	"net/http"
)

// SetupRoutes configures the HTTP router with the API routes for the sensor unit.
func SetupRoutes(mux *http.ServeMux, api *API) {
	mux.HandleFunc("GET /sensors/temperatures", api.getTemperatures)
	mux.HandleFunc("GET /sensors/status", api.getStatus)
	mux.HandleFunc("POST /sensors/status", api.setStatus)
}

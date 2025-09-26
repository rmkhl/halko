package router

import (
	"net/http"

	"github.com/rmkhl/halko/types"
	"github.com/rmkhl/halko/types/log"
)

func SetupRoutes(mux *http.ServeMux, api *API, endpoints *types.APIEndpoints) {
	log.Trace("Setting up HTTP routes")
	log.Trace("Registering GET %s for temperatures", endpoints.Temperatures)
	mux.HandleFunc("GET "+endpoints.Temperatures, api.getTemperatures)
	log.Trace("Registering GET %s for status", endpoints.Status)
	mux.HandleFunc("GET "+endpoints.Status, api.getStatus)
	log.Trace("Registering POST %s for status updates", endpoints.Status)
	mux.HandleFunc("POST "+endpoints.Status, api.setStatus)
	log.Trace("All routes registered successfully")
}

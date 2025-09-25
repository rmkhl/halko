package router

import (
	"net/http"

	"github.com/rmkhl/halko/types"
)

func SetupRoutes(mux *http.ServeMux, api *API, endpoints *types.APIEndpoints) {
	mux.HandleFunc("GET "+endpoints.Temperatures, api.getTemperatures)
	mux.HandleFunc("GET "+endpoints.Status, api.getStatus)
	mux.HandleFunc("POST "+endpoints.Status, api.setStatus)
}

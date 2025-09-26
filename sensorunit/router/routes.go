package router

import (
	"net/http"

	"github.com/rmkhl/halko/types"
	"github.com/rmkhl/halko/types/log"
)

func SetupRoutes(mux *http.ServeMux, api *API, endpoints *types.APIEndpoints) {
	mux.HandleFunc("GET "+endpoints.SensorUnit.Temperatures, api.getTemperatures)
	mux.HandleFunc("GET "+endpoints.SensorUnit.Status, api.getStatus)
	mux.HandleFunc("POST "+endpoints.SensorUnit.Display, api.setDisplay)
	log.Info("HTTP API initialized with 3 endpoints: %s, %s, %s",
		endpoints.SensorUnit.Temperatures, endpoints.SensorUnit.Status, endpoints.SensorUnit.Display)
}

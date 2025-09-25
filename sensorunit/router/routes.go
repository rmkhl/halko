package router

import (
	"net/http"
)

func SetupRoutes(mux *http.ServeMux, api *API) {
	mux.HandleFunc("GET /sensors/temperatures", api.getTemperatures)
	mux.HandleFunc("GET /sensors/status", api.getStatus)
	mux.HandleFunc("POST /sensors/status", api.setStatus)
}

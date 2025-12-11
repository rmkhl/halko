package router

import (
	"net/http"

	"github.com/rmkhl/halko/powerunit/power"
	"github.com/rmkhl/halko/powerunit/shelly"
	"github.com/rmkhl/halko/types"
)

func setupRoutes(mux *http.ServeMux, p *power.Controller, powerMapping map[string]int, idMapping [shelly.NumberOfDevices]string, endpoints *types.APIEndpoints) {
	mux.HandleFunc("GET "+endpoints.PowerUnit.Power, getAllPercentages(p, idMapping))
	mux.HandleFunc("POST "+endpoints.PowerUnit.Power, setAllPercentages(p, powerMapping))
	mux.HandleFunc("GET "+endpoints.PowerUnit.Status, getStatus(p))
}

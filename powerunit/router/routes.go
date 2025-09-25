package router

import (
	"net/http"

	"github.com/rmkhl/halko/powerunit/power"
	"github.com/rmkhl/halko/powerunit/shelly"
)

func setupRoutes(mux *http.ServeMux, p *power.Controller, powerMapping map[string]int, idMapping [shelly.NumberOfDevices]string) {
	mux.HandleFunc("GET /powers", getAllPercentages(p, idMapping))
	mux.HandleFunc("POST /powers", setAllPercentages(p, powerMapping))
}

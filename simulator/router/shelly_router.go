package router

import (
	"net/http"

	"github.com/rmkhl/halko/types/log"
)

// SetupShellyRoutes sets up routes for the Shelly emulation server
func SetupShellyRoutes(mux *http.ServeMux, shellyControls map[int8]interface{}) {
	log.Trace("Setting up Shelly emulation routes")
	mux.HandleFunc("GET /rpc/Switch.GetStatus", readSwitchStatus(shellyControls))
	mux.HandleFunc("GET /rpc/Switch.Set", setSwitchState(shellyControls))
	log.Info("Shelly API initialized with 2 endpoints: /rpc/Switch.GetStatus, /rpc/Switch.Set")
}

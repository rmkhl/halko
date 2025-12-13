package router

import (
	"net/http"

	"github.com/rmkhl/halko/simulator/engine"
	"github.com/rmkhl/halko/types"
	"github.com/rmkhl/halko/types/log"
)

// SetupSensorUnitRoutes sets up routes for the SensorUnit emulation server using configurable endpoints
func SetupSensorUnitRoutes(mux *http.ServeMux, temperatureSensors map[string]engine.TemperatureSensor, endpoints types.SensorUnitEndpoints) {
	log.Trace("Setting up SensorUnit emulation routes with configurable endpoints")
	router := &Router{}

	// Use configured endpoint paths
	tempEndpoint := "GET " + endpoints.Temperatures
	statusEndpoint := "GET " + endpoints.Status
	displayEndpoint := "POST " + endpoints.Display

	mux.HandleFunc(tempEndpoint, readAllTemperatureSensors(temperatureSensors))
	mux.HandleFunc(statusEndpoint, router.getStatus)
	mux.HandleFunc(displayEndpoint, router.setDisplay)

	log.Info("SensorUnit API initialized with configurable endpoints: %s, %s, %s",
		endpoints.Temperatures, endpoints.Status, endpoints.Display)
}

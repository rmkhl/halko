package router

import (
	"net/http"

	"github.com/rmkhl/halko/simulator/engine"
	"github.com/rmkhl/halko/types"
	"github.com/rmkhl/halko/types/log"
)

func readAllTemperatureSensors(sensors map[string]engine.TemperatureSensor) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Trace("Processing temperature request from %s", r.RemoteAddr)
		resp := make(types.TemperatureResponse)
		for name, sensor := range sensors {
			resp[name] = sensor.Temperature()
		}
		log.Trace("Retrieved %d temperature readings from simulator", len(resp))

		response := types.APIResponse[types.TemperatureResponse]{Data: resp}
		log.Trace("Returning temperature data: oven=%.1f°C, material=%.1f°C", resp["oven"], resp["material"])
		writeJSON(w, http.StatusOK, response)
	}
}

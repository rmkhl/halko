package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rmkhl/halko/simulator/engine"
	"github.com/rmkhl/halko/types"
)

func readAllTemperatureSensors(sensors map[string]engine.TemperatureSensor) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		resp := make(types.TemperatureResponse)
		for name, sensor := range sensors {
			resp[name] = sensor.Temperature()
		}
		ctx.JSON(http.StatusOK, types.APIResponse[types.TemperatureResponse]{Data: resp})
	}
}

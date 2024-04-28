package router

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rmkhl/halko/simulator/types"
)

func readAllTemperatureSensors(sensors map[string]types.TemperatureSensor) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		resp := make(types.TemperatureResponse)
		for name, sensor := range sensors {
			resp[name] = sensor.Temperature()
		}
		ctx.JSON(http.StatusOK, types.APIResponse[types.TemperatureResponse]{Data: resp})
	}
}

func readTemperatureSensor(sensors map[string]types.TemperatureSensor) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		sensorName, _ := ctx.Params.Get("sensor")
		sensor, known := sensors[sensorName]
		if !known {
			ctx.JSON(http.StatusNotFound, types.APIErrorResponse{Err: fmt.Sprintf("Unknown temperature sensor '%s'", sensorName)})
			return
		}
		resp := make(types.TemperatureResponse)
		resp[sensorName] = sensor.Temperature()
		ctx.JSON(http.StatusOK, types.APIResponse[types.TemperatureResponse]{Data: resp})
	}
}

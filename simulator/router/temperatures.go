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
		ctx.JSON(http.StatusOK, types.ApiResponse[types.TemperatureResponse]{Data: resp})
	}
}

func readTemperatureSensor(sensors map[string]types.TemperatureSensor) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		sensor_name, _ := ctx.Params.Get("sensor")
		sensor, known := sensors[sensor_name]
		if !known {
			ctx.JSON(http.StatusNotFound, types.ApiErrorResponse{Err: fmt.Sprintf("Unknown temperature sensor '%s'", sensor_name)})
			return
		}
		resp := make(types.TemperatureResponse)
		resp[sensor_name] = sensor.Temperature()
		ctx.JSON(http.StatusOK, types.ApiResponse[types.TemperatureResponse]{Data: resp})
	}
}

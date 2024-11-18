package router

import (
	"github.com/gin-gonic/gin"
	"github.com/rmkhl/halko/simulator/engine"
)

func SetupRoutes(r *gin.Engine, temperatureSensors map[string]engine.TemperatureSensor, powerSensors map[string]engine.PowerSensor, powerControls map[string]engine.PowerManager) {
	sensorAPI := r.Group("sensors/api")
	controlAPI := r.Group("controls/api")

	sensorAPIV1 := sensorAPI.Group("v1")
	controlAPIV1 := controlAPI.Group("v1")

	tempSensors := sensorAPIV1.Group("temperatures")
	tempSensors.GET("", readAllTemperatureSensors(temperatureSensors))
	tempSensors.GET(":sensor", readTemperatureSensor(temperatureSensors))

	psuSensors := sensorAPIV1.Group("powers")
	psuSensors.GET("", statusAllPowers(powerSensors))
	psuSensors.GET(":power", statusPower(powerSensors))

	controls := controlAPIV1.Group("powers")
	controls.POST(":power", operatePower(powerControls))
	controls.PUT(":power", operatePower(powerControls))
	controls.PATCH(":power", operatePower(powerControls))
}

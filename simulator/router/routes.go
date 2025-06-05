package router

import (
	"github.com/gin-gonic/gin"
	"github.com/rmkhl/halko/simulator/engine"
)

func SetupRoutes(r *gin.Engine, temperatureSensors map[string]engine.TemperatureSensor, shellyControls map[int8]interface{}) {
	sensorAPI := r.Group("sensors/api")
	sensorAPIV1 := sensorAPI.Group("v1")

	tempSensors := sensorAPIV1.Group("temperatures")
	tempSensors.GET("", readAllTemperatureSensors(temperatureSensors))
	tempSensors.GET(":sensor", readTemperatureSensor(temperatureSensors))

	shellyAPI := r.Group("rpc")
	shellyRead := shellyAPI.Group("Switch.GetStatus")
	shellyRead.GET("", readSwitchStatus(shellyControls))
	shellyWrite := shellyAPI.Group("Switch.Set")
	shellyWrite.GET("", setSwitchState(shellyControls))
}

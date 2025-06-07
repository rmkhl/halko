package router

import (
	"github.com/gin-gonic/gin"
	"github.com/rmkhl/halko/simulator/engine"
)

// Router is a placeholder for router dependencies, if any in the future.
// We define it to be consistent with other modules like sensorunit.
// Currently, it's empty as status handlers are simple.
type Router struct{}

func SetupRoutes(r *gin.Engine, temperatureSensors map[string]engine.TemperatureSensor, shellyControls map[int8]interface{}) {
	router := &Router{}
	sensorAPI := r.Group("sensors/api")
	sensorAPIV1 := sensorAPI.Group("v1")

	tempSensors := sensorAPIV1.Group("temperatures")
	tempSensors.GET("", readAllTemperatureSensors(temperatureSensors))
	tempSensors.GET(":sensor", readTemperatureSensor(temperatureSensors))

	statusAPI := sensorAPIV1.Group("status")
	statusAPI.GET("", router.getStatus)
	statusAPI.POST("", router.setStatus)

	shellyAPI := r.Group("rpc")
	shellyRead := shellyAPI.Group("Switch.GetStatus")
	shellyRead.GET("", readSwitchStatus(shellyControls))
	shellyWrite := shellyAPI.Group("Switch.Set")
	shellyWrite.GET("", setSwitchState(shellyControls))
}

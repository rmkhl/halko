package router

import (
	"github.com/gin-gonic/gin"
	"github.com/rmkhl/halko/simulator/engine"
)

type Router struct{}

func SetupRoutes(r *gin.Engine, temperatureSensors map[string]engine.TemperatureSensor, shellyControls map[int8]interface{}) {
	SetupSensorRoutes(r, temperatureSensors)
	SetupShellyRoutes(r, shellyControls)
}

func SetupSensorRoutes(r *gin.Engine, temperatureSensors map[string]engine.TemperatureSensor) {
	router := &Router{}
	r.GET("/temperatures", readAllTemperatureSensors(temperatureSensors))
	r.GET("/status", router.getStatus)
	r.POST("/status", router.setStatus)
}

func SetupShellyRoutes(r *gin.Engine, shellyControls map[int8]interface{}) {
	shellyAPI := r.Group("rpc")
	shellyRead := shellyAPI.Group("Switch.GetStatus")
	shellyRead.GET("", readSwitchStatus(shellyControls))
	shellyWrite := shellyAPI.Group("Switch.Set")
	shellyWrite.GET("", setSwitchState(shellyControls))
}

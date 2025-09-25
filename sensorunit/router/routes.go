package router

import (
	"github.com/gin-gonic/gin"
)

// SetupRoutes configures the Gin router with the API routes for the sensor unit.
func SetupRoutes(r *gin.Engine, api *API) {
	sensors := r.Group("sensors")

	temperatures := sensors.Group("temperatures")
	temperatures.GET("", api.getTemperatures)

	status := sensors.Group("status")
	status.GET("", api.getStatus)
	status.POST("", api.setStatus)
}

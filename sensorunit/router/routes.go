package router

import (
	"github.com/gin-gonic/gin"
)

// SetupRoutes configures the Gin router with the API routes for the sensor unit.
// It follows a similar pattern to powerunit, grouping routes under /api/v1.
func SetupRoutes(r *gin.Engine, api *API) {
	sensorsAPI := r.Group("sensors/api")
	apiV1 := sensorsAPI.Group("v1")

	temperatures := apiV1.Group("temperatures")
	temperatures.GET("", api.getTemperatures)

	status := apiV1.Group("status")
	status.GET("", api.getStatus)
	status.POST("", api.setStatus)
}

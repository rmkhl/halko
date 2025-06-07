package router

import (
	"github.com/gin-gonic/gin"
	"github.com/rmkhl/halko/sensorunit/serial"
)

// API represents the REST API for the sensor unit
type API struct {
	sensorUnit *serial.SensorUnit
}

// NewAPI creates a new API instance
func NewAPI(sensorUnit *serial.SensorUnit) *API {
	return &API{
		sensorUnit: sensorUnit,
	}
}

// SetupRouter configures the Gin router with the API routes
func SetupRouter(api *API) *gin.Engine {
	router := gin.Default()

	// Add CORS headers
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Setup routes
	setupTemperatureRoutes(router, api)
	setupStatusRoutes(router, api)

	return router
}

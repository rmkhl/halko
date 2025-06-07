package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rmkhaklab/halko/types"
)

// setupTemperatureRoutes configures the temperature API routes
func setupTemperatureRoutes(router *gin.Engine, api *API) {
	router.GET("/api/temperature", api.getTemperatures)
}

// getTemperatures handles GET requests to fetch temperature data
func (api *API) getTemperatures(c *gin.Context) {
	temperatures, err := api.sensorUnit.GetTemperatures()
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.APIErrorResponse{
			Err: err.Error(),
		})
		return
	}

	// Convert to the simulator format
	response := make(types.TemperatureResponse)

	// Map the sensor values to the expected keys
	for _, temp := range temperatures {
		switch temp.Name {
		case "OvenPrimary":
			response["oven"] = float32(temp.Value)
		case "Wood":
			response["material"] = float32(temp.Value)
			// We could add other mappings here if needed
		}
	}

	c.JSON(http.StatusOK, types.APIResponse[types.TemperatureResponse]{
		Data: response,
	})
}

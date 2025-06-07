package router

import (
	"net/http"

	"github.com/rmkhl/halko/types"

	"github.com/gin-gonic/gin"
)

// getTemperatures handles GET requests to fetch temperature data
// This function is now part of the API struct and called by SetupRoutes.
// No longer a standalone setupTemperatureRoutes function.
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

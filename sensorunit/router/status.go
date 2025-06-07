package router

import (
	"net/http"

	"github.com/rmkhl/halko/types"

	"github.com/gin-gonic/gin"
)

// setStatus handles POST requests to update the status text on the LCD
// This function is now part of the API struct and called by SetupRoutes.
// No longer a standalone setupStatusRoutes function.
func (api *API) setStatus(c *gin.Context) {
	var payload types.StatusRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, types.APIErrorResponse{
			Err: "Invalid request format",
		})
		return
	}

	err := api.sensorUnit.SetStatusText(payload.Message)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.APIErrorResponse{
			Err: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, types.APIResponse[types.StatusResponse]{
		Data: types.StatusResponse{
			Status: types.SensorStatusOK,
		},
	})
}

// getStatus handles GET requests to check the connection status
func (api *API) getStatus(c *gin.Context) {
	isConnected := api.sensorUnit.IsConnected()

	status := types.SensorStatusConnected
	if !isConnected {
		status = types.SensorStatusDisconnected
	}

	c.JSON(http.StatusOK, types.APIResponse[types.StatusResponse]{
		Data: types.StatusResponse{
			Status: status,
		},
	})
}

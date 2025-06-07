package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rmkhaklab/halko/types"
)

// Status represents a status message for the display
type Status struct {
	Message string `json:"message"`
}

// setupStatusRoutes configures the status API routes
func setupStatusRoutes(router *gin.Engine, api *API) {
	router.POST("/api/status", api.setStatus)
	router.GET("/api/status", api.getStatus)
}

// setStatus handles POST requests to update the status text on the LCD
func (api *API) setStatus(c *gin.Context) {
	var status Status
	if err := c.ShouldBindJSON(&status); err != nil {
		c.JSON(http.StatusBadRequest, types.APIErrorResponse{
			Err: "Invalid request format",
		})
		return
	}

	err := api.sensorUnit.SetStatusText(status.Message)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.APIErrorResponse{
			Err: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

// getStatus handles GET requests to check the connection status
func (api *API) getStatus(c *gin.Context) {
	isConnected := api.sensorUnit.IsConnected()

	status := "connected"
	if !isConnected {
		status = "disconnected"
	}

	c.JSON(http.StatusOK, gin.H{
		"status": status,
	})
}

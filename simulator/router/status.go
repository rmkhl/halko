package router

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rmkhl/halko/types"
)

func (r *Router) setStatus(c *gin.Context) {
	var payload types.StatusRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, types.APIErrorResponse{
			Err: "Invalid request format: " + err.Error(),
		})
		return
	}

	log.Printf("Simulator received status update: %s", payload.Message)

	c.JSON(http.StatusOK, types.APIResponse[types.StatusResponse]{
		Data: types.StatusResponse{
			Status: types.SensorStatusOK,
		},
	})
}

func (r *Router) getStatus(c *gin.Context) {
	c.JSON(http.StatusOK, types.APIResponse[types.StatusResponse]{
		Data: types.StatusResponse{
			Status: types.SensorStatusConnected,
		},
	})
}

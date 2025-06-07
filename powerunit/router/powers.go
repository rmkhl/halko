package router

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rmkhl/halko/powerunit/power"
	"github.com/rmkhl/halko/powerunit/shelly"
	"github.com/rmkhl/halko/types"
)

func statusAllPowers(p *power.Controller) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Get all power percentages from the controller
		cycles, err := p.GetAllCycles()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, types.APIErrorResponse{Err: err.Error()})
			return
		}

		// Convert to a readable map with power names and percentages
		readable := make(types.PowerStatusResponse)
		for id, percentage := range cycles {
			readable[id.String()] = types.PowerResponse{Percent: percentage}
		}

		ctx.JSON(http.StatusOK, types.APIResponse[types.PowerStatusResponse]{Data: readable})
	}
}

func statusPower(p *power.Controller) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		powerName, _ := ctx.Params.Get("power")
		id := getIDFromString(powerName)
		if id == shelly.UnknownID {
			ctx.JSON(http.StatusNotFound, types.APIErrorResponse{Err: fmt.Sprintf("Unknown power '%s'", powerName)})
			return
		}

		// Get power percentage from the controller
		percentage, err := p.GetCycle(id)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, types.APIErrorResponse{Err: err.Error()})
			return
		}

		// Return the percentage
		ctx.JSON(http.StatusOK, types.APIResponse[types.PowerResponse]{
			Data: types.PowerResponse{Percent: percentage},
		})
	}
}

func operatePower(p *power.Controller) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var command types.PowerCommand

		if err := ctx.ShouldBind(&command); err != nil {
			ctx.JSON(http.StatusBadRequest, types.APIErrorResponse{Err: err.Error()})
			return
		}
		powerName, _ := ctx.Params.Get("power")
		id := getIDFromString(powerName)
		if id == shelly.UnknownID {
			ctx.JSON(http.StatusBadRequest, types.APIErrorResponse{Err: fmt.Sprintf("Unknown power '%s'", powerName)})
			return
		}

		// Set the power percentage through the controller
		err := p.SetCycle(command.Percent, id)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, types.APIErrorResponse{Err: fmt.Sprintf("error setting power cycle: %s", err)})
			return
		}

		ctx.JSON(http.StatusOK, types.APIResponse[types.PowerOperationResponse]{
			Data: types.PowerOperationResponse{Message: "completed"},
		})
	}
}

// Helper function to convert string to ID
func getIDFromString(name string) shelly.ID {
	switch name {
	case "fan":
		return shelly.Fan
	case "heater":
		return shelly.Heater
	case "humidifier":
		return shelly.Humidifier
	default:
		return shelly.UnknownID
	}
}

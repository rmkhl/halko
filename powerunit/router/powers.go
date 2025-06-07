package router

import (
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

func operateAllPowers(p *power.Controller) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var commands types.PowersCommand

		if err := ctx.ShouldBind(&commands); err != nil {
			ctx.JSON(http.StatusBadRequest, types.APIErrorResponse{Err: err.Error()})
			return
		}

		// Convert the commands to the format expected by the controller
		cycles := make(map[shelly.ID]uint8)
		for powerName, command := range commands {
			id := getIDFromString(powerName)
			if id == shelly.UnknownID {
				ctx.JSON(http.StatusBadRequest, types.APIErrorResponse{Err: "Unknown power '" + powerName + "'"})
				return
			}
			cycles[id] = command.Percent
		}

		// Set all power percentages through the controller
		err := p.SetAllCycles(cycles)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, types.APIErrorResponse{Err: "Error setting power cycles: " + err.Error()})
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

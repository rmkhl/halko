package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rmkhl/halko/powerunit/power"
	"github.com/rmkhl/halko/types"
)

// powerMapping will be injected or read from config
var powerMapping map[string]int

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
			// Find the name for the ID
			var powerName string
			for name, mappedID := range powerMapping {
				if mappedID == id { // No need to cast to shelly.ID
					powerName = name
					break
				}
			}
			if powerName == "" {
				// Handle unknown ID if necessary, though controller should prevent this
				powerName = "unknown"
			}
			readable[powerName] = types.PowerResponse{Percent: percentage}
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
		cycles := make(map[int]uint8) // Changed shelly.ID to int
		for powerName, command := range commands {
			id, ok := powerMapping[powerName]
			if !ok {
				ctx.JSON(http.StatusBadRequest, types.APIErrorResponse{Err: "Unknown power '" + powerName + "'"})
				return
			}
			cycles[id] = command.Percent // No need to cast to shelly.ID
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

// SetPowerMapping allows main to inject the mapping
func SetPowerMapping(mapping map[string]int) {
	powerMapping = mapping
}

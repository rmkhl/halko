package router

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rmkhl/halko/types"
)

// PowerInfo defines the minimal interface needed for power status
type PowerInfo interface {
	Info() (bool, bool)
}

func readSwitchStatus(powers map[int8]interface{}) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Extract the switch ID from the query parameters
		switchID := ctx.Query("id")
		if switchID == "" {
			ctx.String(http.StatusBadRequest, "Switch ID is required")
			return
		}

		id, err := strconv.Atoi(switchID)
		if err != nil {
			ctx.String(http.StatusBadRequest, "Invalid Switch ID %s", switchID)
			return
		}

		if power, exists := powers[int8(id)]; exists {
			// Check if the power element implements PowerInfo
			if powerInfo, ok := power.(PowerInfo); ok {
				// Log the switch status
				_, turnedOn := powerInfo.Info()

				ctx.JSON(http.StatusOK, types.ShellySwitchGetStatusResponse{
					ID:     strconv.Itoa(id),
					Source: "HTTP_in",
					Output: turnedOn,
					Temperature: struct {
						TC float32 `json:"tC"`
						TF float32 `json:"tF"`
					}{
						TC: 20.0,
						TF: 68.0,
					},
				})
			} else {
				ctx.String(http.StatusInternalServerError, "Switch %d does not implement required interface", id)
			}
		} else {
			ctx.String(http.StatusNotFound, "Switch %d not found", id)
		}
	}
}

func setSwitchState(powers map[int8]interface{}) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Extract the switch ID from the query parameters
		switchID := ctx.Query("id")
		if switchID == "" {
			ctx.String(http.StatusBadRequest, "Switch ID is required")
			return
		}

		// Get the desired on/off state from the parameters
		turnOn := ctx.Query("on")
		if turnOn == "" {
			ctx.String(http.StatusBadRequest, "On parameter is required")
			return
		}

		id, err := strconv.Atoi(switchID)
		if err != nil {
			ctx.String(http.StatusBadRequest, "Invalid Switch ID %s", switchID)
			return
		}

		if power, exists := powers[int8(id)]; exists {
			// Check if the power element implements PowerInfo
			if powerInfo, ok := power.(PowerInfo); ok {
				// Get the current state before changing it
				_, previousState := powerInfo.Info()

				// Check if the object also has methods to change state
				if switcher, ok := power.(interface{ SwitchTo(bool) }); ok {
					// Convert parameter to boolean
					newState := turnOn == "true"

					// Change the state
					switcher.SwitchTo(newState)

					// Determine the previous state
					wasOn := previousState

					ctx.JSON(http.StatusOK, types.ShellySwitchSetResponse{
						WasOn: wasOn,
					})
				} else {
					ctx.String(http.StatusInternalServerError, "Switch %d does not support state changes", id)
				}
			} else {
				ctx.String(http.StatusInternalServerError, "Switch %d does not implement required interface", id)
			}
		} else {
			ctx.String(http.StatusNotFound, "Switch %d not found", id)
		}
	}
}

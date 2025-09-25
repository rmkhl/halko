package router

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type PowerInfo interface {
	Info() (bool, bool)
}

func readSwitchStatus(powers map[int8]interface{}) gin.HandlerFunc {
	return func(ctx *gin.Context) {
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
			if powerInfo, ok := power.(PowerInfo); ok {
				_, turnedOn := powerInfo.Info()

				response := struct {
					Code    int  `json:"code"`
					Message string `json:"message"`
					Output  bool `json:"output"`
				}{
					Code:    0,
					Message: "",
					Output:  turnedOn,
				}
				ctx.JSON(http.StatusOK, response)
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
		switchID := ctx.Query("id")
		if switchID == "" {
			ctx.String(http.StatusBadRequest, "Switch ID is required")
			return
		}

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
			if switcher, ok := power.(interface{ SwitchTo(bool) }); ok {
				newState := turnOn == "true"

				switcher.SwitchTo(newState)

				response := struct {
					Code    int  `json:"code"`
					Message string `json:"message"`
					Output  bool `json:"output"`
				}{
					Code:    0,
					Message: "",
					Output:  newState,
				}
				ctx.JSON(http.StatusOK, response)
			} else {
				ctx.String(http.StatusInternalServerError, "Switch %d does not support state changes", id)
			}
		} else {
			ctx.String(http.StatusNotFound, "Switch %d not found", id)
		}
	}
}

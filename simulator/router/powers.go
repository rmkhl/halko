package router

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rmkhl/halko/simulator/engine"
	"github.com/rmkhl/halko/types"
)

func statusAllPowers(powers map[string]engine.PowerSensor) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		resp := make(types.PowerStatusResponse)
		for name, power := range powers {
			if power.IsOn() {
				resp[name] = types.PowerResponse{Status: types.PowerOn, Percent: power.CurrentCycle()}
			} else {
				resp[name] = types.PowerResponse{Status: types.PowerOff}
			}
		}
		ctx.JSON(http.StatusOK, types.APIResponse[types.PowerStatusResponse]{Data: resp})
	}
}

func statusPower(powers map[string]engine.PowerSensor) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		powerName, _ := ctx.Params.Get("power")
		power, known := powers[powerName]
		if !known {
			ctx.JSON(http.StatusNotFound, types.APIErrorResponse{Err: fmt.Sprintf("Unknown power '%s'", powerName)})
			return
		}
		resp := make(types.PowerStatusResponse)
		if power.IsOn() {
			resp[powerName] = types.PowerResponse{Status: types.PowerOn, Percent: power.CurrentCycle()}
			return
		}
		resp[powerName] = types.PowerResponse{Status: types.PowerOff}
	}
}

func operatePower(powers map[string]engine.PowerManager) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var command types.PowerCommand

		if ctx.ShouldBind(&command) != nil {
			ctx.JSON(http.StatusBadRequest, types.APIErrorResponse{Err: "Does not compute"})
			return
		}
		powerName, _ := ctx.Params.Get("power")
		power, known := powers[powerName]
		if !known {
			ctx.JSON(http.StatusBadRequest, types.APIErrorResponse{Err: fmt.Sprintf("Unknown power '%s'", powerName)})
		}

		switch command.Command {
		case types.PowerOn:
			power.TurnOn(engine.NewCycle(command.Percent))
		case types.PowerOff:
			power.TurnOff()
		default:
			ctx.JSON(http.StatusBadRequest, types.APIErrorResponse{Err: fmt.Sprintf("Unknown command '%s'", command.Command)})
			return
		}
		ctx.JSON(http.StatusOK, types.APIResponse[types.PowerOperationResponse]{Data: types.PowerOperationResponse{Message: "completed"}})
	}
}

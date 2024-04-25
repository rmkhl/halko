package router

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rmkhl/halko/simulator/types"
)

func statusAllPowers(powers map[string]types.PowerSensor) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		resp := make(types.PowerStatusResponse)
		for name, power := range powers {
			if power.IsOn() {
				resp[name] = types.PowerResponse{Status: types.POWER_ON, Cycle: power.CurrentCycle()}
			} else {
				resp[name] = types.PowerResponse{Status: types.POWER_OFF}
			}
		}
		ctx.JSON(http.StatusOK, types.ApiResponse[types.PowerStatusResponse]{Data: resp})
	}
}

func statusPower(powers map[string]types.PowerSensor) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		power_name, _ := ctx.Params.Get("power")
		power, known := powers[power_name]
		if !known {
			ctx.JSON(http.StatusNotFound, types.ApiErrorResponse{Err: fmt.Sprintf("Unknown power '%s'", power_name)})
			return
		}
		resp := make(types.PowerStatusResponse)
		if power.IsOn() {
			resp[power_name] = types.PowerResponse{Status: types.POWER_ON, Cycle: power.CurrentCycle()}
			return
		}
		resp[power_name] = types.PowerResponse{Status: types.POWER_OFF}
	}
}

func operatePower(powers map[string]types.PowerManager) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var command types.PowerCommand

		if ctx.ShouldBind(&command) != nil {
			ctx.JSON(http.StatusBadRequest, types.ApiErrorResponse{Err: "Does not compute"})
			return
		}
		power_name, _ := ctx.Params.Get("power")
		power, known := powers[power_name]
		if !known {
			ctx.JSON(http.StatusBadRequest, types.ApiErrorResponse{Err: fmt.Sprintf("Unknown power '%s'", power_name)})
		}
		switch command.Command {
		case types.POWER_ON:
			power.TurnOn(types.NewCycle(command.Cycle.Name, command.Cycle.Ticks))
		case types.POWER_OFF:
			power.TurnOff()
		case types.POWER_NEXT:
			power.SwitchTo(types.NewCycle(command.Cycle.Name, command.Cycle.Ticks))
		default:
			ctx.JSON(http.StatusBadRequest, types.ApiErrorResponse{Err: fmt.Sprintf("Unknown command '%s'", command.Command)})
			return
		}
		ctx.JSON(http.StatusOK, types.ApiResponse[types.PowerOperationResponse]{Data: types.PowerOperationResponse{Message: "completed"}})
	}
}

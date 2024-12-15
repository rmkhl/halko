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
		resp, err := p.GetAllStates()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, types.APIErrorResponse{Err: err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, types.APIResponse[power.States]{Data: resp})
	}
}

func statusPower(p *power.Controller) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		powerName, _ := ctx.Params.Get("power")
		id, known := shelly.IDString(powerName).ID()
		if !known {
			ctx.JSON(http.StatusNotFound, types.APIErrorResponse{Err: fmt.Sprintf("Unknown power '%s'", powerName)})
			return
		}
		resp, err := p.GetState(id)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, types.APIErrorResponse{Err: err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, types.APIResponse[shelly.PowerState]{Data: resp})
	}
}

func operatePower(p *power.Controller) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var command types.PowerCommand

		if ctx.ShouldBind(&command) != nil {
			ctx.JSON(http.StatusBadRequest, types.APIErrorResponse{Err: "Does not compute"})
			return
		}
		powerName, _ := ctx.Params.Get("power")
		id, known := shelly.IDString(powerName).ID()
		if !known {
			ctx.JSON(http.StatusBadRequest, types.APIErrorResponse{Err: fmt.Sprintf("Unknown power '%s'", powerName)})
			return
		}
		err := p.SetCycle(command.Percent, id)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, types.APIErrorResponse{Err: fmt.Sprintf("error setting powercycle: ", err)})
			return
		}
		ctx.JSON(http.StatusOK, types.APIResponse[types.PowerOperationResponse]{Data: types.PowerOperationResponse{Message: "completed"}})
	}
}

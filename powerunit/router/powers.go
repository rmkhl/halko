package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rmkhl/halko/powerunit/power"
	"github.com/rmkhl/halko/powerunit/shelly"
	"github.com/rmkhl/halko/types"
)

func getAllPercentages(p *power.Controller, idMapping [shelly.NumberOfDevices]string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		percentages := p.GetAllPercentages()

		response := make(types.PowerStatusResponse)
		for id := range shelly.NumberOfDevices {
			response[idMapping[id]] = types.PowerResponse{Percent: percentages[id]}
		}

		ctx.JSON(http.StatusOK, types.APIResponse[types.PowerStatusResponse]{Data: response})
	}
}

func setAllPercentages(p *power.Controller, powerMapping map[string]int) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var commands types.PowersCommand

		if err := ctx.ShouldBind(&commands); err != nil {
			ctx.JSON(http.StatusBadRequest, types.APIErrorResponse{Err: err.Error()})
			return
		}

		var percentages [shelly.NumberOfDevices]uint8

		for powerName, command := range commands {
			id, ok := powerMapping[powerName]
			if !ok {
				ctx.JSON(http.StatusBadRequest, types.APIErrorResponse{Err: "Unknown power '" + powerName + "'"})
				return
			}
			percentages[id] = command.Percent
		}

		p.SetAllPercentages(percentages)

		ctx.JSON(http.StatusOK, types.APIResponse[types.PowerOperationResponse]{
			Data: types.PowerOperationResponse{Message: "completed"},
		})
	}
}

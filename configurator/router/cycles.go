package router

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rmkhl/halko/configurator/database"
)

func allCycles(cycles database.Cycles) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		cycles, err := cycles.All()
		status, err := statusAndError(err)
		if err != nil {
			ctx.JSON(status, errorJSON(err))
			return
		}

		ctx.JSON(status, cycles)
	}
}

func cycle(cycles database.Cycles) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id, ok := ctx.Params.Get("id")
		if !ok {
			ctx.JSON(http.StatusBadRequest, errors.New("no id provided"))
			return
		}

		cycle, err := cycles.ByID(id)
		status, err := statusAndError(err)
		if err != nil {
			ctx.JSON(status, errorJSON(err))
			return
		}

		ctx.JSON(status, cycle)
	}
}

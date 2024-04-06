package router

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rmkhl/halko/configurator/database"
)

func allPhases(phases database.Phases) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		phases, err := phases.All()
		status, err := statusAndError(err)
		if err != nil {
			ctx.JSON(status, errorJSON(err))
			return
		}

		ctx.JSON(status, phases)
	}
}

func phase(phases database.Phases) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id, ok := ctx.Params.Get("id")
		if !ok {
			ctx.JSON(http.StatusBadRequest, errors.New("no id provided"))
			return
		}

		phase, err := phases.ByID(id)
		status, err := statusAndError(err)
		if err != nil {
			ctx.JSON(status, errorJSON(err))
			return
		}

		ctx.JSON(status, phase)
	}
}

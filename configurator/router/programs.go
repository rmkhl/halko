package router

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rmkhl/halko/configurator/database"
)

func currentProgram(programs database.Programs) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		curr, err := programs.Current()
		status, err := statusAndError(err)
		if err != nil {
			ctx.JSON(status, errorJSON(err))
			return
		}

		ctx.JSON(status, curr)
	}
}

func allPrograms(programs database.Programs) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		programs, err := programs.All()
		status, err := statusAndError(err)
		if err != nil {
			ctx.JSON(status, errorJSON(err))
			return
		}

		ctx.JSON(status, programs)
	}
}

func program(programs database.Programs) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id, ok := ctx.Params.Get("id")
		if !ok {
			ctx.JSON(http.StatusBadRequest, errors.New("no id provided"))
			return
		}

		program, err := programs.ByID(id)
		status, err := statusAndError(err)
		if err != nil {
			ctx.JSON(status, errorJSON(err))
			return
		}

		ctx.JSON(status, program)
	}
}

package router

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rmkhl/halko/configurator/database"
	"github.com/rmkhl/halko/types"
)

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
		name, ok := ctx.Params.Get("name")
		if !ok {
			ctx.JSON(http.StatusBadRequest, ErrNoName)
			return
		}

		program, err := programs.ByName(name)
		status, err := statusAndError(err)
		if err != nil {
			ctx.JSON(status, errorJSON(err))
			return
		}

		ctx.JSON(status, program)
	}
}

func createProgram(programs database.Programs) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		defer ctx.Request.Body.Close()
		prog := types.Program{}

		if err := json.NewDecoder(ctx.Request.Body).Decode(&prog); err != nil {
			ctx.JSON(http.StatusBadRequest, errorJSON(err))
			return
		}

		created, err := programs.CreateOrUpdate(prog.ProgramName, &prog)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, errorJSON(err))
			return
		}

		ctx.JSON(http.StatusCreated, created)
	}
}

func updateProgram(programs database.Programs) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		name, ok := ctx.Params.Get("name")
		if !ok {
			ctx.JSON(http.StatusBadRequest, ErrNoName)
			return
		}

		defer ctx.Request.Body.Close()
		prog := types.Program{}

		if err := json.NewDecoder(ctx.Request.Body).Decode(&prog); err != nil {
			ctx.JSON(http.StatusBadRequest, errorJSON(err))
			return
		}

		updated, err := programs.CreateOrUpdate(name, &prog)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, errorJSON(err))
			return
		}

		ctx.JSON(http.StatusOK, updated)
	}
}

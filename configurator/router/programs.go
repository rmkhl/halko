package router

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rmkhl/halko/configurator/database"
	"github.com/rmkhl/halko/configurator/domain"
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

func createProgram(programs database.Programs) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		defer ctx.Request.Body.Close()
		prog := domain.Program{}

		if err := json.NewDecoder(ctx.Request.Body).Decode(&prog); err != nil {
			ctx.JSON(http.StatusBadRequest, errorJSON(err))
			return
		}

		prog.ID = ""

		created, err := programs.CreateOrUpdate(&prog)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, errorJSON(err))
			return
		}

		ctx.JSON(http.StatusCreated, created)
	}
}

func updateProgram(programs database.Programs) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id, ok := ctx.Params.Get("id")
		if !ok {
			ctx.JSON(http.StatusBadRequest, errors.New("no id provided"))
			return
		}

		defer ctx.Request.Body.Close()
		prog := domain.Program{}

		if err := json.NewDecoder(ctx.Request.Body).Decode(&prog); err != nil {
			ctx.JSON(http.StatusBadRequest, errorJSON(err))
			return
		}

		prog.ID = domain.ID(id)

		updated, err := programs.CreateOrUpdate(&prog)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, errorJSON(err))
			return
		}

		ctx.JSON(http.StatusCreated, updated)
	}
}

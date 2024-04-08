package router

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rmkhl/halko/configurator/database"
	"github.com/rmkhl/halko/configurator/domain"
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

func createCycle(cycles database.Cycles) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		defer ctx.Request.Body.Close()
		cycle := domain.Cycle{}

		if err := json.NewDecoder(ctx.Request.Body).Decode(&cycle); err != nil {
			ctx.JSON(http.StatusBadRequest, errorJSON(err))
			return
		}

		cycle.ID = ""

		created, err := cycles.CreateOrUpdate(&cycle)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, errorJSON(err))
			return
		}

		ctx.JSON(http.StatusCreated, created)
	}
}

func updateCycle(cycles database.Cycles) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id, ok := ctx.Params.Get("id")
		if !ok {
			ctx.JSON(http.StatusBadRequest, errors.New("no id provided"))
			return
		}

		defer ctx.Request.Body.Close()
		cycle := domain.Cycle{}

		if err := json.NewDecoder(ctx.Request.Body).Decode(&cycle); err != nil {
			ctx.JSON(http.StatusBadRequest, errorJSON(err))
			return
		}

		cycle.ID = domain.ID(id)

		updated, err := cycles.CreateOrUpdate(&cycle)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, errorJSON(err))
			return
		}

		ctx.JSON(http.StatusOK, updated)
	}
}

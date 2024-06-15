package router

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rmkhl/halko/configurator/database"
	"github.com/rmkhl/halko/configurator/domain"
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
		name, ok := ctx.Params.Get("name")
		if !ok {
			ctx.JSON(http.StatusBadRequest, ErrNoName)
			return
		}

		phase, err := phases.ByName(name)
		status, err := statusAndError(err)
		if err != nil {
			ctx.JSON(status, errorJSON(err))
			return
		}

		ctx.JSON(status, phase)
	}
}

func createPhase(phases database.Phases) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		defer ctx.Request.Body.Close()
		phase := domain.Phase{}

		if err := json.NewDecoder(ctx.Request.Body).Decode(&phase); err != nil {
			ctx.JSON(http.StatusBadRequest, errorJSON(err))
			return
		}

		created, err := phases.CreateOrUpdate("", &phase)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, errorJSON(err))
			return
		}

		ctx.JSON(http.StatusCreated, created)
	}
}

func updatePhase(phases database.Phases) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		name, ok := ctx.Params.Get("name")
		if !ok {
			ctx.JSON(http.StatusBadRequest, ErrNoName)
			return
		}

		defer ctx.Request.Body.Close()
		phase := domain.Phase{}

		if err := json.NewDecoder(ctx.Request.Body).Decode(&phase); err != nil {
			ctx.JSON(http.StatusBadRequest, errorJSON(err))
			return
		}

		updated, err := phases.CreateOrUpdate(name, &phase)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, errorJSON(err))
			return
		}

		ctx.JSON(http.StatusOK, updated)
	}
}

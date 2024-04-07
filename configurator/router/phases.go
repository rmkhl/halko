package router

import (
	"encoding/json"
	"errors"
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

func createPhase(phases database.Phases) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		defer ctx.Request.Body.Close()
		phase := domain.Phase{}

		if err := json.NewDecoder(ctx.Request.Body).Decode(&phase); err != nil {
			ctx.JSON(http.StatusBadRequest, errorJSON(err))
			return
		}

		phase.ID = ""

		created, err := phases.CreateOrUpdate(&phase)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, errorJSON(err))
			return
		}

		ctx.JSON(http.StatusCreated, created)
	}
}

func updatePhase(phases database.Phases) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id, ok := ctx.Params.Get("id")
		if !ok {
			ctx.JSON(http.StatusBadRequest, errors.New("no id provided"))
			return
		}

		defer ctx.Request.Body.Close()
		phase := domain.Phase{}

		if err := json.NewDecoder(ctx.Request.Body).Decode(&phase); err != nil {
			ctx.JSON(http.StatusBadRequest, errorJSON(err))
			return
		}

		phase.ID = domain.ID(id)

		updated, err := phases.CreateOrUpdate(&phase)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, errorJSON(err))
			return
		}

		ctx.JSON(http.StatusCreated, updated)
	}
}

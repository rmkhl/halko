package router

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rmkhl/halko/executor/engine"
	"github.com/rmkhl/halko/executor/types"
)

func getCurrentProgram(engine *engine.ControlEngine) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		currentStatus := engine.CurrentStatus()
		if currentStatus == nil {
			ctx.JSON(http.StatusNoContent, types.APIErrorResponse{Err: "No program running"})
			return
		}
		ctx.JSON(http.StatusOK, types.APIResponse[types.ProgramStatus]{Data: *currentStatus})
	}
}

func startNewProgram(engine *engine.ControlEngine) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var program types.Program

		err := ctx.ShouldBind(&program)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, types.APIErrorResponse{Err: fmt.Sprintf("Does not compute (%s)", err.Error())})
			return
		}
		err = program.Validate()
		if err != nil {
			ctx.JSON(http.StatusBadRequest, types.APIErrorResponse{Err: err.Error()})
			return
		}
		err = engine.StartEngine(&program)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, types.APIErrorResponse{Err: err.Error()})
			return
		}
		ctx.JSON(http.StatusCreated, types.APIResponse[types.Program]{Data: program})
	}
}

func cancelRunningProgram(engine *engine.ControlEngine) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		err := engine.StopEngine()
		if err != nil {
			ctx.JSON(http.StatusNotFound, types.APIErrorResponse{Err: err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, types.APIResponse[string]{Data: "Stopped"})
	}
}

package router

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rmkhl/halko/executor/engine"
	"github.com/rmkhl/halko/types"
)

func getCurrentProgram(engine *engine.ControlEngine) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		currentStatus := engine.CurrentStatus()
		if currentStatus == nil {
			ctx.JSON(http.StatusNoContent, types.APIErrorResponse{Err: "No program running"})
			return
		}
		ctx.JSON(http.StatusOK, types.APIResponse[types.ExecutionStatus]{Data: *currentStatus})
	}
}

func startNewProgram(engine *engine.ControlEngine) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Read and log the raw request body, then restore it for binding
		body, err := io.ReadAll(ctx.Request.Body)
		if err != nil {
			log.Printf("Failed to read request body: %v", err)
			ctx.JSON(http.StatusBadRequest, types.APIErrorResponse{Err: fmt.Sprintf("Failed to read request body (%s)", err.Error())})
			return
		}
		log.Printf("Raw request body: %s", string(body))

		// Restore the request body for ShouldBind to use
		ctx.Request.Body = io.NopCloser(bytes.NewBuffer(body))

		var program types.Program

		err = ctx.ShouldBind(&program)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, types.APIErrorResponse{Err: fmt.Sprintf("Does not compute (%s)", err.Error())})
			return
		}

		// Log the program before validation
		log.Printf("Received program: %s with %d steps", program.ProgramName, len(program.ProgramSteps))
		for i, step := range program.ProgramSteps {
			log.Printf("  Step %d: %s (%s) - Target: %d°C", i+1, step.Name, step.StepType, step.TargetTemperature)
		}

		// Apply defaults before validation
		program.ApplyDefaults(engine.GetDefaults())

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

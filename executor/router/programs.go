package router

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rmkhl/halko/executor/engine"
	"github.com/rmkhl/halko/executor/storage"
	"github.com/rmkhl/halko/types"
)

func listAllRuns(storage *storage.FileStorage) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var savedPrograms []types.RunHistory

		programs, err := storage.ListExecutedPrograms()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, types.APIErrorResponse{Err: err.Error()})
			return
		}

		for _, programName := range programs {
			state, updatedAt, _ := storage.LoadState(programName)
			savedPrograms = append(savedPrograms, types.RunHistory{
				Name:        programName,
				State:       state,
				CompletedAt: updatedAt,
				StartedAt:   startTimeFromName(programName),
			})
		}
		ctx.JSON(http.StatusOK, types.APIResponse[[]types.RunHistory]{Data: savedPrograms})
	}
}

func startTimeFromName(name string) int64 {
	nameParts := strings.Split(name, "@")
	if len(nameParts) == 2 {
		parsedTime, err := time.Parse(time.RFC3339, nameParts[1])
		if err == nil {
			return parsedTime.Unix()
		}
	}
	return 0
}

func getRun(storage *storage.FileStorage) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		programName, _ := ctx.Params.Get("name")
		program, err := storage.LoadExecutedProgram(programName)
		if err != nil {
			ctx.JSON(http.StatusNotFound, types.APIErrorResponse{Err: err.Error()})
			return
		}
		state, updatedAt, _ := storage.LoadState(programName)
		ctx.JSON(http.StatusOK, types.APIResponse[types.ExecutedProgram]{
			Data: types.ExecutedProgram{
				RunHistory: types.RunHistory{State: state, CompletedAt: updatedAt, StartedAt: startTimeFromName(programName)},
				Program:    *program,
			},
		})
	}
}

func deleteRun(storage *storage.FileStorage) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		programName, _ := ctx.Params.Get("name")

		err := storage.DeleteExecutedProgram(programName)
		storage.MaybeDeleteState(programName)
		storage.MaybeDeleteExecutionLog(programName)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, types.APIErrorResponse{Err: err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, types.APIResponse[string]{Data: "deleted"})
	}
}

func listAllPrograms(storage *storage.FileStorage) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		programs, err := storage.ListStoredPrograms()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, types.APIErrorResponse{Err: err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, types.APIResponse[[]string]{Data: programs})
	}
}

func getProgram(storage *storage.FileStorage) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		programName, _ := ctx.Params.Get("name")
		program, err := storage.LoadStoredProgram(programName)
		if err != nil {
			ctx.JSON(http.StatusNotFound, types.APIErrorResponse{Err: err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, types.APIResponse[types.Program]{Data: *program})
	}
}

func createProgram(storage *storage.FileStorage, engine *engine.ControlEngine) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var program types.Program

		err := ctx.ShouldBind(&program)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, types.APIErrorResponse{Err: "Invalid JSON: " + err.Error()})
			return
		}

		// Create a deep copy for validation
		programCopy, err := program.Duplicate()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, types.APIErrorResponse{Err: "Failed to copy program: " + err.Error()})
			return
		}

		// Apply defaults to the copy and validate
		programCopy.ApplyDefaults(engine.GetDefaults())
		err = programCopy.Validate()
		if err != nil {
			ctx.JSON(http.StatusBadRequest, types.APIErrorResponse{Err: err.Error()})
			return
		}

		// Store the original program without defaults applied
		err = storage.CreateStoredProgram(program.ProgramName, &program)
		if err != nil {
			ctx.JSON(http.StatusConflict, types.APIErrorResponse{Err: err.Error()})
			return
		}

		ctx.JSON(http.StatusCreated, types.APIResponse[types.Program]{Data: program})
	}
}

func updateProgram(storage *storage.FileStorage, engine *engine.ControlEngine) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		programName, _ := ctx.Params.Get("name")
		var program types.Program

		err := ctx.ShouldBind(&program)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, types.APIErrorResponse{Err: "Invalid JSON: " + err.Error()})
			return
		}

		// Create a deep copy for validation
		programCopy, err := program.Duplicate()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, types.APIErrorResponse{Err: "Failed to copy program: " + err.Error()})
			return
		}

		// Apply defaults to the copy and validate
		programCopy.ApplyDefaults(engine.GetDefaults())
		err = programCopy.Validate()
		if err != nil {
			ctx.JSON(http.StatusBadRequest, types.APIErrorResponse{Err: err.Error()})
			return
		}

		// Store the original program without defaults applied
		program.ProgramName = programName

		err = storage.UpdateStoredProgram(programName, &program)
		if err != nil {
			ctx.JSON(http.StatusNotFound, types.APIErrorResponse{Err: err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, types.APIResponse[types.Program]{Data: program})
	}
}

func deleteProgram(storage *storage.FileStorage) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		programName, _ := ctx.Params.Get("name")

		err := storage.DeleteStoredProgram(programName)
		if err != nil {
			ctx.JSON(http.StatusNotFound, types.APIErrorResponse{Err: err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, types.APIResponse[string]{Data: "deleted"})
	}
}

package router

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rmkhl/halko/executor/storage"
	"github.com/rmkhl/halko/executor/types"
)

func listAllPrograms(storage *storage.ProgramStorage) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var savedPrograms []types.SavedProgram

		programs, err := storage.ListPrograms()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, types.APIErrorResponse{Err: err.Error()})
			return
		}

		for _, programName := range programs {
			state, updated_at, _ := storage.LoadState(programName)
			savedPrograms = append(savedPrograms, types.SavedProgram{
				Name:        programName,
				State:       state,
				CompletedAt: updated_at,
				StartedAt:   startTimeFromName(programName),
			})
		}
		ctx.JSON(http.StatusBadRequest, types.APIResponse[[]types.SavedProgram]{Data: savedPrograms})
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

func getProgram(storage *storage.ProgramStorage) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		programName, _ := ctx.Params.Get("name")
		program, err := storage.LoadProgram(programName)
		if err != nil {
			ctx.JSON(http.StatusNotFound, types.APIErrorResponse{Err: err.Error()})
			return
		}
		state, updated_at, _ := storage.LoadState(programName)
		ctx.JSON(http.StatusOK, types.APIResponse[types.ExecutedProgram]{
			Data: types.ExecutedProgram{
				Program:     *program,
				State:       state,
				CompletedAt: updated_at,
				StartedAt:   startTimeFromName(programName),
			},
		})
	}
}

func deleteProgram(storage *storage.ProgramStorage) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		programName, _ := ctx.Params.Get("name")

		err := storage.DeleteProgram(programName)
		storage.MaybeDeleteState(programName)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, types.APIErrorResponse{Err: err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, types.APIResponse[string]{Data: "deleted"})
	}
}

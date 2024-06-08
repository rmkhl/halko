package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rmkhl/halko/executor/storage"
	"github.com/rmkhl/halko/executor/types"
)

func listAllPrograms(storage *storage.ProgramStorage) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		programs, err := storage.ListPrograms()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, types.APIErrorResponse{Err: err.Error()})
			return
		}
		ctx.JSON(http.StatusBadRequest, types.APIResponse[[]string]{Data: programs})
	}
}

func getProgram(storage *storage.ProgramStorage) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		programName, _ := ctx.Params.Get("name")
		program, err := storage.LoadProgram(programName)
		if err != nil {
			ctx.JSON(http.StatusNotFound, types.APIErrorResponse{Err: err.Error()})
			return
		}
		state, err := storage.LoadState(programName)
		ctx.JSON(http.StatusOK, types.APIResponse[types.ExecutedProgram]{Data: types.ExecutedProgram{Program: *program, State: state}})
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

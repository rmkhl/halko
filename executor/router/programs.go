package router

import (
	"net/http"
	"strings"
	"time"

	"github.com/rmkhl/halko/executor/storage"
	"github.com/rmkhl/halko/types"
)

func listAllRuns(storage *storage.FileStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var savedPrograms []types.RunHistory

		programs, err := storage.ListExecutedPrograms()
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
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
		writeJSON(w, http.StatusOK, types.APIResponse[[]types.RunHistory]{Data: savedPrograms})
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

func getRun(storage *storage.FileStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		programName := r.PathValue("name")
		program, err := storage.LoadExecutedProgram(programName)
		if err != nil {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		state, updatedAt, _ := storage.LoadState(programName)
		writeJSON(w, http.StatusOK, types.APIResponse[types.ExecutedProgram]{
			Data: types.ExecutedProgram{
				RunHistory: types.RunHistory{State: state, CompletedAt: updatedAt, StartedAt: startTimeFromName(programName)},
				Program:    *program,
			},
		})
	}
}

func deleteRun(storage *storage.FileStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		programName := r.PathValue("name")

		err := storage.DeleteExecutedProgram(programName)
		storage.MaybeDeleteState(programName)
		storage.MaybeDeleteExecutionLog(programName)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, types.APIResponse[string]{Data: "deleted"})
	}
}

// Storage-related program handlers have been moved to the independent storage service

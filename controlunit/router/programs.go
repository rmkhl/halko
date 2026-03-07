package router

import (
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/rmkhl/halko/controlunit/engine"
	"github.com/rmkhl/halko/types"
)

func listAllRuns(storage types.ExecutionStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
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

		// Sort by completion date (latest first)
		sort.Slice(savedPrograms, func(i, j int) bool {
			// Programs with CompletedAt come before those without
			if savedPrograms[i].CompletedAt == 0 && savedPrograms[j].CompletedAt != 0 {
				return false
			}
			if savedPrograms[i].CompletedAt != 0 && savedPrograms[j].CompletedAt == 0 {
				return true
			}
			// Both have CompletedAt: sort descending (latest first)
			if savedPrograms[i].CompletedAt != 0 && savedPrograms[j].CompletedAt != 0 {
				return savedPrograms[i].CompletedAt > savedPrograms[j].CompletedAt
			}
			// Neither has CompletedAt: sort by StartedAt descending
			return savedPrograms[i].StartedAt > savedPrograms[j].StartedAt
		})

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

func getRun(storage types.ExecutionStorage) http.HandlerFunc {
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

func deleteRun(storage types.ExecutionStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		programName := r.PathValue("name")

		err := storage.DeleteExecutedProgram(programName)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, types.APIResponse[string]{Data: "deleted"})
	}
}

func getRunLog(storage types.ExecutionStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		programName := r.PathValue("name")
		logPath := storage.GetLogPath(programName)

		content, err := os.ReadFile(logPath)
		if err != nil {
			writeError(w, http.StatusNotFound, "Log file not found")
			return
		}

		w.Header().Set("Content-Type", "text/csv")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(content)
	}
}

func getRunningLog(storage types.ExecutionStorage, engine *engine.ControlEngine) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		programName := engine.CurrentProgramName()
		if programName == "" {
			writeError(w, http.StatusNoContent, "No program running")
			return
		}

		logPath := storage.GetRunningLogPath(programName)
		content, err := os.ReadFile(logPath)
		if err != nil {
			writeError(w, http.StatusNotFound, "Log file not found")
			return
		}

		w.Header().Set("Content-Type", "text/csv")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(content)
	}
}

package router

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/rmkhl/halko/executor/engine"
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

func listAllPrograms(storage *storage.FileStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		programs, err := storage.ListStoredPrograms()
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, types.APIResponse[[]string]{Data: programs})
	}
}

func getProgram(storage *storage.FileStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		programName := r.PathValue("name")
		program, err := storage.LoadStoredProgram(programName)
		if err != nil {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, types.APIResponse[types.Program]{Data: *program})
	}
}

func createProgram(storage *storage.FileStorage, engine *engine.ControlEngine) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var program types.Program

		err := json.NewDecoder(r.Body).Decode(&program)
		if err != nil {
			writeError(w, http.StatusBadRequest, "Invalid JSON: "+err.Error())
			return
		}

		// Create a deep copy for validation
		programCopy, err := program.Duplicate()
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to copy program: "+err.Error())
			return
		}

		// Apply defaults to the copy and validate
		programCopy.ApplyDefaults(engine.GetDefaults())
		err = programCopy.Validate()
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		// Store the original program without defaults applied
		err = storage.CreateStoredProgram(program.ProgramName, &program)
		if err != nil {
			writeError(w, http.StatusConflict, err.Error())
			return
		}

		writeJSON(w, http.StatusCreated, types.APIResponse[types.Program]{Data: program})
	}
}

func updateProgram(storage *storage.FileStorage, engine *engine.ControlEngine) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		programName := r.PathValue("name")
		var program types.Program

		err := json.NewDecoder(r.Body).Decode(&program)
		if err != nil {
			writeError(w, http.StatusBadRequest, "Invalid JSON: "+err.Error())
			return
		}

		// Create a deep copy for validation
		programCopy, err := program.Duplicate()
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to copy program: "+err.Error())
			return
		}

		// Apply defaults to the copy and validate
		programCopy.ApplyDefaults(engine.GetDefaults())
		err = programCopy.Validate()
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		// Store the original program without defaults applied
		program.ProgramName = programName

		err = storage.UpdateStoredProgram(programName, &program)
		if err != nil {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}

		writeJSON(w, http.StatusOK, types.APIResponse[types.Program]{Data: program})
	}
}

func deleteProgram(storage *storage.FileStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		programName := r.PathValue("name")

		err := storage.DeleteStoredProgram(programName)
		if err != nil {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}

		writeJSON(w, http.StatusOK, types.APIResponse[string]{Data: "deleted"})
	}
}

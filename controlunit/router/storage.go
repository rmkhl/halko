package router

import (
	"encoding/json"
	"net/http"

	"github.com/rmkhl/halko/controlunit/storagefs"
	"github.com/rmkhl/halko/types"
)

func listAllStoredPrograms(storage *storagefs.ProgramStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		programs, err := storage.ListStoredPrograms()
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, types.APIResponse[[]string]{Data: programs})
	}
}

func getStoredProgram(storage *storagefs.ProgramStorage) http.HandlerFunc {
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

func createStoredProgram(storage *storagefs.ProgramStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var program types.Program

		err := json.NewDecoder(r.Body).Decode(&program)
		if err != nil {
			writeError(w, http.StatusBadRequest, "Invalid JSON: "+err.Error())
			return
		}

		if program.ProgramName == "" {
			writeError(w, http.StatusBadRequest, "Program name is required")
			return
		}

		err = storage.CreateStoredProgram(program.ProgramName, &program)
		if err != nil {
			writeError(w, http.StatusConflict, err.Error())
			return
		}

		writeJSON(w, http.StatusCreated, types.APIResponse[types.Program]{Data: program})
	}
}

func updateStoredProgram(storage *storagefs.ProgramStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		programName := r.PathValue("name")
		var program types.Program

		err := json.NewDecoder(r.Body).Decode(&program)
		if err != nil {
			writeError(w, http.StatusBadRequest, "Invalid JSON: "+err.Error())
			return
		}

		program.ProgramName = programName

		err = storage.UpdateStoredProgram(programName, &program)
		if err != nil {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}

		writeJSON(w, http.StatusOK, types.APIResponse[types.Program]{Data: program})
	}
}

func deleteStoredProgram(storage *storagefs.ProgramStorage) http.HandlerFunc {
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

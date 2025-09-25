package router

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/rmkhl/halko/storage/filestorage"
	"github.com/rmkhl/halko/types"
)

// writeJSON writes a JSON response
func writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// writeError writes an error response
func writeError(w http.ResponseWriter, statusCode int, message string) {
	writeJSON(w, statusCode, types.APIErrorResponse{Err: message})
}

// extractPathParam extracts a parameter from the URL path
func extractPathParam(path, pattern string) string {
	pathParts := strings.Split(strings.Trim(path, "/"), "/")
	patternParts := strings.Split(strings.Trim(pattern, "/"), "/")

	for i, part := range patternParts {
		if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") && i < len(pathParts) {
			return pathParts[i]
		}
	}
	return ""
}

func SetupRoutes(mux *http.ServeMux, storage *filestorage.FileStorage) {
	// Storage routes for programs
	mux.HandleFunc("GET /storage/programs", listAllPrograms(storage))
	mux.HandleFunc("GET /storage/programs/{name}", getProgram(storage))
	mux.HandleFunc("POST /storage/programs", createProgram(storage))
	mux.HandleFunc("POST /storage/programs/{name}", updateProgram(storage))
	mux.HandleFunc("DELETE /storage/programs/{name}", deleteProgram(storage))
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

func listAllPrograms(storage *filestorage.FileStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		programs, err := storage.ListStoredPrograms()
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, types.APIResponse[[]string]{Data: programs})
	}
}

func getProgram(storage *filestorage.FileStorage) http.HandlerFunc {
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

func createProgram(storage *filestorage.FileStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var program types.Program

		err := json.NewDecoder(r.Body).Decode(&program)
		if err != nil {
			writeError(w, http.StatusBadRequest, "Invalid JSON: "+err.Error())
			return
		}

		// Basic validation - we don't have access to engine defaults here
		// We'll validate that the program has a name and some basic structure
		if program.ProgramName == "" {
			writeError(w, http.StatusBadRequest, "Program name is required")
			return
		}

		// Store the program
		err = storage.CreateStoredProgram(program.ProgramName, &program)
		if err != nil {
			writeError(w, http.StatusConflict, err.Error())
			return
		}

		writeJSON(w, http.StatusCreated, types.APIResponse[types.Program]{Data: program})
	}
}

func updateProgram(storage *filestorage.FileStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		programName := r.PathValue("name")
		var program types.Program

		err := json.NewDecoder(r.Body).Decode(&program)
		if err != nil {
			writeError(w, http.StatusBadRequest, "Invalid JSON: "+err.Error())
			return
		}

		// Set the program name from the URL
		program.ProgramName = programName

		err = storage.UpdateStoredProgram(programName, &program)
		if err != nil {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}

		writeJSON(w, http.StatusOK, types.APIResponse[types.Program]{Data: program})
	}
}

func deleteProgram(storage *filestorage.FileStorage) http.HandlerFunc {
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

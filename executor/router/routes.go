package router

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/rmkhl/halko/executor/engine"
	"github.com/rmkhl/halko/executor/storage"
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

func SetupRoutes(mux *http.ServeMux, storage *storage.FileStorage, engine *engine.ControlEngine) {
	// Engine routes
	mux.HandleFunc("GET /engine/programs", listAllRuns(storage))
	mux.HandleFunc("GET /engine/programs/{name}", getRun(storage))
	mux.HandleFunc("DELETE /engine/programs/{name}", deleteRun(storage))

	mux.HandleFunc("GET /engine/running", getCurrentProgram(engine))
	mux.HandleFunc("POST /engine/running", startNewProgram(engine))
	mux.HandleFunc("DELETE /engine/running", cancelRunningProgram(engine))

	// Storage routes
	mux.HandleFunc("GET /storage/programs", listAllPrograms(storage))
	mux.HandleFunc("GET /storage/programs/{name}", getProgram(storage))
	mux.HandleFunc("POST /storage/programs", createProgram(storage, engine))
	mux.HandleFunc("POST /storage/programs/{name}", updateProgram(storage, engine))
	mux.HandleFunc("DELETE /storage/programs/{name}", deleteProgram(storage))
}

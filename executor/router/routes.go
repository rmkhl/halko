package router

import (
	"encoding/json"
	"net/http"

	"github.com/rmkhl/halko/executor/engine"
	"github.com/rmkhl/halko/executor/storage"
	"github.com/rmkhl/halko/types"
)

func writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		// Log the error but don't change the response as headers are already sent
		// In a real application, you might want to handle this more gracefully
		_ = err
	}
}

func writeError(w http.ResponseWriter, statusCode int, message string) {
	writeJSON(w, statusCode, types.APIErrorResponse{Err: message})
}

func SetupRoutes(mux *http.ServeMux, storage *storage.FileStorage, engine *engine.ControlEngine) {
	mux.HandleFunc("GET /engine/programs", listAllRuns(storage))
	mux.HandleFunc("GET /engine/programs/{name}", getRun(storage))
	mux.HandleFunc("DELETE /engine/programs/{name}", deleteRun(storage))

	mux.HandleFunc("GET /engine/running", getCurrentProgram(engine))
	mux.HandleFunc("POST /engine/running", startNewProgram(engine))
	mux.HandleFunc("DELETE /engine/running", cancelRunningProgram(engine))

	// Note: /storage/ endpoints are now handled by the independent storage service
}

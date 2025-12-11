package router

import (
	"encoding/json"
	"net/http"

	"github.com/rmkhl/halko/controlunit/engine"
	"github.com/rmkhl/halko/controlunit/storagefs"
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

func SetupRoutes(mux *http.ServeMux, storage *storagefs.ExecutorFileStorage, engine *engine.ControlEngine, endpoints *types.APIEndpoints) {
	mux.HandleFunc("GET "+endpoints.ControlUnit.Programs, listAllRuns(storage))
	mux.HandleFunc("GET "+endpoints.ControlUnit.Programs+"/{name}", getRun(storage))
	mux.HandleFunc("DELETE "+endpoints.ControlUnit.Programs+"/{name}", deleteRun(storage))

	mux.HandleFunc("GET "+endpoints.ControlUnit.Running, getCurrentProgram(engine))
	mux.HandleFunc("POST "+endpoints.ControlUnit.Running, startNewProgram(engine))
	mux.HandleFunc("DELETE "+endpoints.ControlUnit.Running, cancelRunningProgram(engine))

	mux.HandleFunc("GET "+endpoints.ControlUnit.Status, getStatus(engine))

	// Note: /storage/ endpoints are now handled by the independent storage service
}

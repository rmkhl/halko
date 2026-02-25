package router

import (
	"encoding/json"
	"net/http"

	"github.com/rmkhl/halko/controlunit/engine"
	"github.com/rmkhl/halko/controlunit/storagefs"
	"github.com/rmkhl/halko/types"
)

// corsMiddleware adds CORS headers to allow cross-origin requests from webapp
func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Allow requests from any origin (for development)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next(w, r)
	}
}

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

func SetupRoutes(mux *http.ServeMux, execStorage *storagefs.ExecutorFileStorage, programStorage *storagefs.ProgramStorage, engine *engine.ControlEngine, endpoints *types.APIEndpoints) {
	// WebSocket endpoint for live run log
	mux.HandleFunc("GET "+endpoints.ControlUnit.Engine+"/running/logws", StreamLiveRunLog(engine))
	// Engine execution endpoints
	mux.HandleFunc("GET "+endpoints.ControlUnit.Engine+"/running/log", corsMiddleware(getRunningLog(execStorage, engine)))
	mux.HandleFunc("GET "+endpoints.ControlUnit.Engine+"/history", corsMiddleware(listAllRuns(execStorage)))
	mux.HandleFunc("GET "+endpoints.ControlUnit.Engine+"/history/{name}", corsMiddleware(getRun(execStorage)))
	mux.HandleFunc("GET "+endpoints.ControlUnit.Engine+"/history/{name}/log", corsMiddleware(getRunLog(execStorage)))
	mux.HandleFunc("DELETE "+endpoints.ControlUnit.Engine+"/history/{name}", corsMiddleware(deleteRun(execStorage)))
	mux.HandleFunc("GET "+endpoints.ControlUnit.Engine+"/running", corsMiddleware(getCurrentProgram(engine)))
	mux.HandleFunc("POST "+endpoints.ControlUnit.Engine+"/running", corsMiddleware(startNewProgram(engine)))
	mux.HandleFunc("DELETE "+endpoints.ControlUnit.Engine+"/running", corsMiddleware(cancelRunningProgram(engine)))
	mux.HandleFunc("GET "+endpoints.ControlUnit.Engine+"/defaults", corsMiddleware(getDefaults(engine)))

	// Status endpoint
	mux.HandleFunc("GET "+endpoints.ControlUnit.Status, corsMiddleware(getStatus(engine)))

	// Program storage endpoints (stored/saved programs)
	mux.HandleFunc("GET "+endpoints.ControlUnit.Programs, corsMiddleware(listAllStoredPrograms(programStorage)))
	mux.HandleFunc("GET "+endpoints.ControlUnit.Programs+"/{name}", corsMiddleware(getStoredProgram(programStorage)))
	mux.HandleFunc("POST "+endpoints.ControlUnit.Programs, corsMiddleware(createStoredProgram(programStorage)))
	mux.HandleFunc("POST "+endpoints.ControlUnit.Programs+"/{name}", corsMiddleware(updateStoredProgram(programStorage)))
	mux.HandleFunc("DELETE "+endpoints.ControlUnit.Programs+"/{name}", corsMiddleware(deleteStoredProgram(programStorage)))
}

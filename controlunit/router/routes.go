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

func SetupRoutes(mux *http.ServeMux, execStorage *storagefs.ExecutorFileStorage, programStorage *storagefs.ProgramStorage, engine *engine.ControlEngine, endpoints *types.APIEndpoints) {
	// Engine execution endpoints
	mux.HandleFunc("GET "+endpoints.ControlUnit.Engine+"/history", listAllRuns(execStorage))
	mux.HandleFunc("GET "+endpoints.ControlUnit.Engine+"/history/{name}", getRun(execStorage))
	mux.HandleFunc("DELETE "+endpoints.ControlUnit.Engine+"/history/{name}", deleteRun(execStorage))
	mux.HandleFunc("GET "+endpoints.ControlUnit.Engine+"/running", getCurrentProgram(engine))
	mux.HandleFunc("POST "+endpoints.ControlUnit.Engine+"/running", startNewProgram(engine))
	mux.HandleFunc("DELETE "+endpoints.ControlUnit.Engine+"/running", cancelRunningProgram(engine))
	mux.HandleFunc("GET "+endpoints.ControlUnit.Engine+"/defaults", getDefaults(engine))

	// Status endpoint
	mux.HandleFunc("GET "+endpoints.ControlUnit.Status, getStatus(engine))

	// Program storage endpoints (stored/saved programs)
	mux.HandleFunc("GET "+endpoints.ControlUnit.Programs, listAllStoredPrograms(programStorage))
	mux.HandleFunc("GET "+endpoints.ControlUnit.Programs+"/{name}", getStoredProgram(programStorage))
	mux.HandleFunc("POST "+endpoints.ControlUnit.Programs, createStoredProgram(programStorage))
	mux.HandleFunc("POST "+endpoints.ControlUnit.Programs+"/{name}", updateStoredProgram(programStorage))
	mux.HandleFunc("DELETE "+endpoints.ControlUnit.Programs+"/{name}", deleteStoredProgram(programStorage))
}

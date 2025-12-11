package router

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/rmkhl/halko/controlunit/engine"
	"github.com/rmkhl/halko/types"
)

func getCurrentProgram(engine *engine.ControlEngine) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		currentStatus := engine.CurrentStatus()
		if currentStatus == nil {
			writeError(w, http.StatusNoContent, "No program running")
			return
		}
		writeJSON(w, http.StatusOK, types.APIResponse[types.ExecutionStatus]{Data: *currentStatus})
	}
}

func startNewProgram(engine *engine.ControlEngine) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("Failed to read request body: %v", err)
			writeError(w, http.StatusBadRequest, fmt.Sprintf("Failed to read request body (%s)", err.Error()))
			return
		}
		log.Printf("Raw request body: %s", string(body))

		r.Body = io.NopCloser(bytes.NewBuffer(body))

		var program types.Program

		err = json.NewDecoder(r.Body).Decode(&program)
		if err != nil {
			writeError(w, http.StatusBadRequest, fmt.Sprintf("Does not compute (%s)", err.Error()))
			return
		}

		log.Printf("Received program: %s with %d steps", program.ProgramName, len(program.ProgramSteps))
		for i, step := range program.ProgramSteps {
			log.Printf("  Step %d: %s (%s) - Target: %d°C", i+1, step.Name, step.StepType, step.TargetTemperature)
		}

		program.ApplyDefaults(engine.GetDefaults())

		err = program.Validate()
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		err = engine.StartEngine(&program)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, types.APIResponse[types.Program]{Data: program})
	}
}

func cancelRunningProgram(engine *engine.ControlEngine) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		err := engine.StopEngine()
		if err != nil {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, types.APIResponse[string]{Data: "Stopped"})
	}
}

func getDefaults(engine *engine.ControlEngine) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		defaults := engine.GetDefaults()
		if defaults == nil {
			writeError(w, http.StatusInternalServerError, "Defaults not configured")
			return
		}
		writeJSON(w, http.StatusOK, types.APIResponse[types.Defaults]{Data: *defaults})
	}
}

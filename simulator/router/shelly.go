package router

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/rmkhl/halko/types"
	"github.com/rmkhl/halko/types/log"
)

type PowerInfo interface {
	Info() (bool, bool)
}

func readSwitchStatus(powers map[int8]interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switchID := r.URL.Query().Get("id")
		log.Debug("Processing switch status request for ID: %s from %s", switchID, r.RemoteAddr)

		if switchID == "" {
			log.Warning("Switch status request missing ID parameter")
			http.Error(w, "Switch ID is required", http.StatusBadRequest)
			return
		}

		id, err := strconv.Atoi(switchID)
		if err != nil {
			log.Warning("Invalid switch ID in request: %s", switchID)
			http.Error(w, "Invalid Switch ID "+switchID, http.StatusBadRequest)
			return
		}

		power, exists := powers[int8(id)]
		if !exists {
			log.Warning("Switch %d not found", id)
			http.Error(w, "Switch "+strconv.Itoa(id)+" not found", http.StatusNotFound)
			return
		}

		powerInfo, ok := power.(PowerInfo)
		if !ok {
			log.Error("Switch %d does not implement required interface", id)
			http.Error(w, "Switch "+strconv.Itoa(id)+" does not implement required interface", http.StatusInternalServerError)
			return
		}

		_, turnedOn := powerInfo.Info()
		log.Debug("Switch %d status: output=%v", id, turnedOn)

		response := types.ShellySwitchGetStatusResponse{
			ID:     strconv.Itoa(id),
			Source: "HTTP_in",
			Output: turnedOn,
			Temperature: struct {
				TC float32 `json:"tC"`
				TF float32 `json:"tF"`
			}{
				TC: 20.0,
				TF: 68.0,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Error("Failed to encode JSON response: %v", err)
			http.Error(w, "Internal server error: failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}

func setSwitchState(powers map[int8]interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switchID := r.URL.Query().Get("id")
		turnOn := r.URL.Query().Get("on")
		log.Debug("Processing switch set request for ID: %s, state: %s from %s", switchID, turnOn, r.RemoteAddr)

		if switchID == "" {
			log.Warning("Switch set request missing ID parameter")
			http.Error(w, "Switch ID is required", http.StatusBadRequest)
			return
		}

		if turnOn == "" {
			log.Warning("Switch set request missing 'on' parameter")
			http.Error(w, "On parameter is required", http.StatusBadRequest)
			return
		}

		id, err := strconv.Atoi(switchID)
		if err != nil {
			log.Warning("Invalid switch ID in set request: %s", switchID)
			http.Error(w, "Invalid Switch ID "+switchID, http.StatusBadRequest)
			return
		}

		power, exists := powers[int8(id)]
		if !exists {
			log.Warning("Switch %d not found", id)
			http.Error(w, "Switch "+strconv.Itoa(id)+" not found", http.StatusNotFound)
			return
		}

		powerInfo, ok := power.(PowerInfo)
		if !ok {
			log.Error("Switch %d does not implement required interface", id)
			http.Error(w, "Switch "+strconv.Itoa(id)+" does not implement required interface", http.StatusInternalServerError)
			return
		}

		switcher, ok := power.(interface{ SwitchTo(bool) })
		if !ok {
			log.Error("Switch %d does not support state changes", id)
			http.Error(w, "Switch "+strconv.Itoa(id)+" does not support state changes", http.StatusInternalServerError)
			return
		}

		_, previousState := powerInfo.Info()
		newState := turnOn == "true"
		log.Info("Setting switch %d to %v (was %v)", id, newState, previousState)

		switcher.SwitchTo(newState)
		log.Debug("Switch %d state change queued successfully", id)

		response := types.ShellySwitchSetResponse{
			WasOn: previousState,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Error("Failed to encode JSON response: %v", err)
			http.Error(w, "Internal server error: failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}

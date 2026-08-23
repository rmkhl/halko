package router

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/rmkhl/halko/controlunit/engine"
	"github.com/rmkhl/halko/controlunit/storagefs"
	"github.com/rmkhl/halko/types"
)

type noteRequest struct {
	Text string `json:"text"`
}

// addRunningNote records an observation against whatever run is live. There is
// no run name in the path: the control unit knows which run that is, and a
// client naming one could otherwise post into a run that ended a moment ago.
func addRunningNote(controlEngine *engine.ControlEngine) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request noteRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid JSON: "+err.Error())
			return
		}

		text := strings.TrimSpace(request.Text)
		if text == "" {
			writeError(w, http.StatusBadRequest, "Note text is required")
			return
		}

		note, err := controlEngine.AddRunNote(text)
		// A run that ended between the request arriving and the write is the
		// same answer as no run at all: the note did not land, and saying so
		// beats filing it where nothing will read it.
		if errors.Is(err, engine.ErrNoProgramRunning) || errors.Is(err, storagefs.ErrNotesClosed) {
			writeError(w, http.StatusNotFound, "No program running")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		writeJSON(w, http.StatusCreated, types.APIResponse[types.RunNote]{Data: *note})
	}
}

func getRunningNotes(controlEngine *engine.ControlEngine) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		notes, err := controlEngine.RunNotes()
		if err != nil {
			// Matches GET /engine/running, which answers 204 when idle.
			writeError(w, http.StatusNoContent, "No program running")
			return
		}
		writeJSON(w, http.StatusOK, types.APIResponse[[]types.RunNote]{Data: notes})
	}
}

func getRunNotes(storage types.ExecutionStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		notes, err := storage.LoadRunNotes(r.PathValue("name"))
		if errors.Is(err, types.ErrInvalidStorageName) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, types.APIResponse[[]types.RunNote]{Data: notes})
	}
}

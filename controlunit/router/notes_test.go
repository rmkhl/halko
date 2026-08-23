package router

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rmkhl/halko/controlunit/engine"
	"github.com/rmkhl/halko/controlunit/storagefs"
)

// A POST that had no run to write to is a failure the operator must see, the
// same answer DELETE /engine/running already gives.
func TestAddRunningNoteWithNothingRunning(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/engine/running/notes", strings.NewReader(`{"text":"hello"}`))

	addRunningNote(&engine.ControlEngine{})(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404 (body %q)", rec.Code, rec.Body.String())
	}
}

// A GET matches its sibling GET /engine/running, which answers 204.
func TestGetRunningNotesWithNothingRunning(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/engine/running/notes", nil)

	getRunningNotes(&engine.ControlEngine{})(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("status = %d, want 204 (body %q)", rec.Code, rec.Body.String())
	}
}

// There is nothing to stamp for an empty observation, and the check must come
// before the engine is asked, so the operator gets the real reason.
func TestAddRunningNoteRejectsEmptyText(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/engine/running/notes", strings.NewReader(`{"text":"   "}`))

	addRunningNote(&engine.ControlEngine{})(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400 (body %q)", rec.Code, rec.Body.String())
	}
}

func TestAddRunningNoteRejectsBadJSON(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/engine/running/notes", strings.NewReader(`not json`))

	addRunningNote(&engine.ControlEngine{})(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400 (body %q)", rec.Code, rec.Body.String())
	}
}

// The name arrives from the URL, so it gets the same traversal treatment as
// every other name-keyed route (see traversal_test.go).
func TestGetRunNotesRejectsEscapingName(t *testing.T) {
	storage, err := storagefs.NewExecutorFileStorage(t.TempDir())
	if err != nil {
		t.Fatalf("creating storage: %v", err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/engine/history/x/notes", nil)
	req.SetPathValue("name", "../../etc/passwd")

	getRunNotes(storage)(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400 (body %q)", rec.Code, rec.Body.String())
	}
}

// A run nobody annotated is not an error.
func TestGetRunNotesOfARunWithNoneIsAnEmptyList(t *testing.T) {
	storage, err := storagefs.NewExecutorFileStorage(t.TempDir())
	if err != nil {
		t.Fatalf("creating storage: %v", err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/engine/history/x/notes", nil)
	req.SetPathValue("name", "never-annotated")

	getRunNotes(storage)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body %q)", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"data":[]`) {
		t.Errorf("body = %q, want an empty data array", rec.Body.String())
	}
}

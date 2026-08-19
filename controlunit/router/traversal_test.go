package router

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rmkhl/halko/controlunit/storagefs"
	"github.com/rmkhl/halko/types"
)

// A percent-encoded traversal survives ServeMux's path cleaning and reaches the
// handler as a literal "../..", so the handler must not turn it into a file
// path. The unencoded form is redirected by the mux and never gets this far.
const encodedEscape = "../../etc/passwd"

func TestGetRunLogRejectsEscapingName(t *testing.T) {
	storage, err := storagefs.NewExecutorFileStorage(t.TempDir())
	if err != nil {
		t.Fatalf("creating storage: %v", err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/engine/history/x/log", nil)
	req.SetPathValue("name", encodedEscape)

	getRunLog(storage)(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for an escaping name, got %d (body %q)", rec.Code, rec.Body.String())
	}
}

func TestCreateStoredProgramRejectsEscapingName(t *testing.T) {
	storage, err := storagefs.NewProgramStorage(t.TempDir())
	if err != nil {
		t.Fatalf("creating storage: %v", err)
	}

	body := `{"name":"../../escaped","steps":[]}`
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/programs", strings.NewReader(body))

	createStoredProgram(storage)(rec, req)

	// An unusable name is bad input, not a conflict with an existing program.
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for an escaping name, got %d (body %q)", rec.Code, rec.Body.String())
	}
}

// Guard the assumption the whole fix rests on: a name that is a plain segment
// still works end to end.
func TestPlainNameStillReachesStorage(t *testing.T) {
	storage, err := storagefs.NewProgramStorage(t.TempDir())
	if err != nil {
		t.Fatalf("creating storage: %v", err)
	}

	body := `{"name":"Oak drying","steps":[]}`
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/programs", strings.NewReader(body))

	createStoredProgram(storage)(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 for a plain name, got %d (body %q)", rec.Code, rec.Body.String())
	}
	if _, err := storage.LoadStoredProgram("Oak drying"); err != nil {
		t.Errorf("stored program not readable back: %v", err)
	}
	_ = types.ErrInvalidStorageName
}

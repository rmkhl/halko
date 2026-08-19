package router

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rmkhl/halko/types"
)

// The status endpoint is how a deployed powerunit identifies itself, so the
// version has to survive the whole request path, not just the struct literal.
func TestStatusReportsVersion(t *testing.T) {
	handler, _ := newTestRouter(t)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/status", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var response types.APIResponse[types.ServiceStatusResponse]
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("decoding status response: %v", err)
	}

	if response.Data.Version != types.Version {
		t.Errorf("version is %q, want %q", response.Data.Version, types.Version)
	}
}

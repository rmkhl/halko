package router

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rmkhl/halko/types"
)

// controlunit's own status is what /system/status reports for it, and what the
// webapp shows for the service the rest of the system is driven by.
func TestStatusReportsVersion(t *testing.T) {
	rec := httptest.NewRecorder()
	getStatus()(rec, httptest.NewRequest(http.MethodGet, "/status", nil))

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

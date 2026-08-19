package router

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rmkhl/halko/dbusunit/dbus"
	"github.com/rmkhl/halko/types"
)

// A zero-value Manager is a real manager that has not connected, which is the
// state dbusunit is in whenever the system bus is unreachable. Building one
// directly keeps the test self-contained: depending on a live system bus would
// make it pass or fail on whether the machine happens to have one.
func TestStatusReportsVersionWhenDisconnected(t *testing.T) {
	rec := httptest.NewRecorder()
	getStatus(&dbus.Manager{})(rec, httptest.NewRequest(http.MethodGet, "/status", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var response types.ServiceStatusResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("decoding status response: %v", err)
	}

	if response.Status != types.ServiceStatusUnavailable {
		t.Fatalf("expected an unavailable dbus unit, got %q", response.Status)
	}
	if response.Version != types.Version {
		t.Errorf("version is %q, want %q", response.Version, types.Version)
	}
}

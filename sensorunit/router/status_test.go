package router

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rmkhl/halko/sensorunit/serial"
	"github.com/rmkhl/halko/types"
)

// The sensor unit reports unavailable without a device attached, which is the
// interesting case: the version has to be present even when the service is
// unhealthy, or an operator cannot tell what is running on a broken kiln.
func TestStatusReportsVersionWhenDisconnected(t *testing.T) {
	sensorUnit, err := serial.NewSensorUnit(t.TempDir()+"/absent-device", 9600)
	if err != nil {
		t.Fatalf("creating sensor unit: %v", err)
	}
	t.Cleanup(func() { _ = sensorUnit.Shutdown() })

	api := NewAPI(sensorUnit)

	rec := httptest.NewRecorder()
	api.getStatus(rec, httptest.NewRequest(http.MethodGet, "/status", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var response types.APIResponse[types.ServiceStatusResponse]
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("decoding status response: %v", err)
	}

	if response.Data.Status != types.ServiceStatusUnavailable {
		t.Fatalf("expected an unavailable sensor unit, got %q", response.Data.Status)
	}
	if response.Data.Version != types.Version {
		t.Errorf("version is %q, want %q", response.Data.Version, types.Version)
	}
}

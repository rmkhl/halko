package engine

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/rmkhl/halko/types/log"
)

// captureLog redirects the shared logger for the duration of a test and returns
// everything written to it.
func captureLog(t *testing.T) *bytes.Buffer {
	t.Helper()

	var buf bytes.Buffer
	log.SetOutput(&buf)
	t.Cleanup(func() { log.SetOutput(os.Stdout) })

	return &buf
}

func newTestPSUController(t *testing.T, handler http.HandlerFunc) *psuController {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	return &psuController{
		client:          server.Client(),
		powerControlURL: server.URL + "/power",
	}
}

func TestSetPowerPostsThePercentageToTheNamedDevice(t *testing.T) {
	var gotPath, gotMethod, gotBody string

	p := newTestPSUController(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method

		body := make([]byte, r.ContentLength)
		_, _ = r.Body.Read(body)
		gotBody = string(body)

		_, _ = w.Write([]byte(`{"data":{"percent":40}}`))
	})

	p.setPower(psuOven, 40)

	if gotMethod != http.MethodPost {
		t.Fatalf("expected POST, got %s", gotMethod)
	}
	if gotPath != "/power/heater" {
		t.Fatalf("expected path /power/heater, got %q", gotPath)
	}
	if !strings.Contains(gotBody, `"percent":40`) {
		t.Fatalf("expected the percentage in the body, got %q", gotBody)
	}
}

// Regression: the branch that logs the power unit's message was guarded on the
// decode having *failed*, so the useful message was never printed.
func TestSetPowerLogsTheReportedError(t *testing.T) {
	buf := captureLog(t)

	p := newTestPSUController(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"Unknown power 'heater'"}`))
	})

	p.setPower(psuOven, 40)

	if !strings.Contains(buf.String(), "Unknown power 'heater'") {
		t.Fatalf("expected the reported error in the log, got %q", buf.String())
	}
}

func TestSetPowerFallsBackToTheStatusLine(t *testing.T) {
	buf := captureLog(t)

	p := newTestPSUController(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(`<html>gateway error</html>`))
	})

	p.setPower(psuOven, 40)

	if !strings.Contains(buf.String(), "502") {
		t.Fatalf("expected the status line in the log, got %q", buf.String())
	}
}

func TestSetPowerLogsNothingOnSuccess(t *testing.T) {
	buf := captureLog(t)

	p := newTestPSUController(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":{"percent":40}}`))
	})

	p.setPower(psuOven, 40)

	if strings.Contains(buf.String(), "Cannot set power") {
		t.Fatalf("expected no error log on success, got %q", buf.String())
	}
}

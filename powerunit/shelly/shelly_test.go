package shelly

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"
)

// newTestShelly starts a real HTTP server running the given handler and returns
// a client pointed at it, so every test below exercises the actual request and
// response path rather than a stand-in.
func newTestShelly(t *testing.T, handler http.HandlerFunc) *Shelly {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	return New(server.URL)
}

// respondWith serves a fixed status and body to every request.
func respondWith(status int, body string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}
}

func TestGetStateMapsOutputToPowerState(t *testing.T) {
	tests := []struct {
		name string
		body string
		want PowerState
	}{
		{"output true is on", `{"output":true}`, On},
		{"output false is off", `{"output":false}`, Off},
		{"missing output is off", `{}`, Off},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTestShelly(t, respondWith(http.StatusOK, tt.body))

			got, err := s.GetState(0)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("expected state %q, got %q", tt.want, got)
			}
		})
	}
}

func TestGetStateRequestsTheAddressedDevice(t *testing.T) {
	var gotPath, gotID string

	s := newTestShelly(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotID = r.URL.Query().Get("id")
		_, _ = w.Write([]byte(`{"output":true}`))
	})

	if _, err := s.GetState(2); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotPath != "/rpc/Switch.GetStatus" {
		t.Fatalf("expected path /rpc/Switch.GetStatus, got %q", gotPath)
	}
	if gotID != "2" {
		t.Fatalf("expected id=2, got id=%q", gotID)
	}
}

func TestGetStateReportsDeviceErrors(t *testing.T) {
	// A 200 carrying the Gen2 error shape is still a failure.
	s := newTestShelly(t, respondWith(http.StatusOK, `{"code":-105,"message":"no such switch"}`))

	got, err := s.GetState(0)
	if err == nil {
		t.Fatal("expected an error for a device error response, got nil")
	}
	if got != Unknown {
		t.Fatalf("expected state %q on error, got %q", Unknown, got)
	}
}

func TestGetStateRejectsAMalformedBody(t *testing.T) {
	s := newTestShelly(t, respondWith(http.StatusOK, `not json`))

	got, err := s.GetState(0)
	if err == nil {
		t.Fatal("expected an error for a malformed body, got nil")
	}
	if got != Unknown {
		t.Fatalf("expected state %q on error, got %q", Unknown, got)
	}
}

// A failing request whose body happens to decode must not read as a switch that
// is simply off — the caller has no reading at all in that case.
func TestGetStateRejectsANonOKStatus(t *testing.T) {
	s := newTestShelly(t, respondWith(http.StatusInternalServerError, `{}`))

	got, err := s.GetState(0)
	if err == nil {
		t.Fatal("expected an error for HTTP 500, got nil")
	}
	if got != Unknown {
		t.Fatalf("expected state %q on error, got %q", Unknown, got)
	}
}

func TestSetStateSendsTheRequestedState(t *testing.T) {
	tests := []struct {
		name   string
		state  PowerState
		wantOn string
	}{
		{"on", On, "true"},
		{"off", Off, "false"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotPath, gotID, gotOn string

			s := newTestShelly(t, func(w http.ResponseWriter, r *http.Request) {
				gotPath = r.URL.Path
				gotID = r.URL.Query().Get("id")
				gotOn = r.URL.Query().Get("on")
				_, _ = w.Write([]byte(`{"was_on":false}`))
			})

			got, err := s.SetState(tt.state, 1)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.state {
				t.Fatalf("expected %q echoed back, got %q", tt.state, got)
			}

			if gotPath != "/rpc/Switch.Set" {
				t.Fatalf("expected path /rpc/Switch.Set, got %q", gotPath)
			}
			if gotID != "1" {
				t.Fatalf("expected id=1, got id=%q", gotID)
			}
			if gotOn != tt.wantOn {
				t.Fatalf("expected on=%s, got on=%q", tt.wantOn, gotOn)
			}
		})
	}
}

// The startup sequence and every tick rely on SetState's error to know whether
// the relay actually moved. A 500 that decodes cleanly must not report success.
func TestSetStateRejectsANonOKStatus(t *testing.T) {
	s := newTestShelly(t, respondWith(http.StatusInternalServerError, `{}`))

	got, err := s.SetState(Off, 0)
	if err == nil {
		t.Fatal("expected an error for HTTP 500, got nil")
	}
	if got != Unknown {
		t.Fatalf("expected state %q on error, got %q", Unknown, got)
	}
}

func TestSetStateSurfacesTheDeviceMessage(t *testing.T) {
	s := newTestShelly(t, respondWith(http.StatusBadRequest, `{"code":-103,"message":"bad switch id"}`))

	_, err := s.SetState(On, 9)
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if !strings.Contains(err.Error(), "bad switch id") {
		t.Fatalf("expected the device message in the error, got %q", err.Error())
	}
}

func TestShutdownTurnsOffEveryDevice(t *testing.T) {
	var (
		mu    sync.Mutex
		calls []string
	)

	s := newTestShelly(t, func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		calls = append(calls, r.URL.Query().Get("id")+"="+r.URL.Query().Get("on"))
		mu.Unlock()
		_, _ = w.Write([]byte(`{}`))
	})

	if err := s.Shutdown(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(calls) != NumberOfDevices {
		t.Fatalf("expected %d calls, got %d: %v", NumberOfDevices, len(calls), calls)
	}
	for id, call := range calls {
		want := strconv.Itoa(id) + "=false"
		if call != want {
			t.Fatalf("call %d: expected %q, got %q", id, want, call)
		}
	}
}

func TestShutdownStopsAtTheFirstFailure(t *testing.T) {
	var (
		mu    sync.Mutex
		calls []string
	)

	s := newTestShelly(t, func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")

		mu.Lock()
		calls = append(calls, id)
		mu.Unlock()

		if id == "1" {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{}`))
			return
		}
		_, _ = w.Write([]byte(`{}`))
	})

	if err := s.Shutdown(); err == nil {
		t.Fatal("expected an error when a device fails to switch off, got nil")
	}

	// Device 2 must never be reached: the loop returns on the first failure.
	if len(calls) != 2 {
		t.Fatalf("expected to stop after 2 calls, got %d: %v", len(calls), calls)
	}
}

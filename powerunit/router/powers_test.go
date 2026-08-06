package router

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/rmkhl/halko/powerunit/power"
	"github.com/rmkhl/halko/powerunit/shelly"
	"github.com/rmkhl/halko/types"
)

const (
	heater     = "heater"
	fan        = "fan"
	humidifier = "humidifier"
)

// The mappings powerunit builds from halko.cfg.
var (
	testPowerMapping = map[string]int{heater: 0, fan: 1, humidifier: 2}
	testIDMapping    = [shelly.NumberOfDevices]string{heater, fan, humidifier}
)

// newTestRouter wires the real router to a real power controller talking to a
// real (if unattached) Shelly endpoint, so requests travel the whole path.
func newTestRouter(t *testing.T) (http.Handler, *power.Controller) {
	t.Helper()

	shellyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"output":false}`))
	}))
	t.Cleanup(shellyServer.Close)

	controller := power.New(time.Hour, 100*time.Second, shelly.New(shellyServer.URL))
	t.Cleanup(controller.Stop)

	endpoints := &types.APIEndpoints{}
	endpoints.PowerUnit.Power = "/power"
	endpoints.PowerUnit.Status = "/status"

	return New(controller, testPowerMapping, testIDMapping, endpoints), controller
}

func do(t *testing.T, handler http.Handler, method, target, body string) *httptest.ResponseRecorder {
	t.Helper()

	var reader *strings.Reader
	if body == "" {
		reader = strings.NewReader("")
	} else {
		reader = strings.NewReader(body)
	}

	req := httptest.NewRequest(method, target, reader)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

func TestGetAllPercentagesNamesEveryDevice(t *testing.T) {
	handler, controller := newTestRouter(t)
	controller.SetAllPercentages([shelly.NumberOfDevices]uint8{10, 20, 30})

	rec := do(t, handler, http.MethodGet, "/power", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var response types.APIResponse[types.PowerStatusResponse]
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	for name, want := range map[string]uint8{heater: 10, fan: 20, humidifier: 30} {
		got, ok := response.Data[name]
		if !ok {
			t.Fatalf("expected %q in the response, got %v", name, response.Data)
		}
		if got.Percent != want {
			t.Fatalf("expected %s at %d%%, got %d%%", name, want, got.Percent)
		}
	}
}

// Regression: a command naming only some devices used to zero the rest, so
// setting the heater silently switched the fan and humidifier off.
func TestPartialCommandLeavesUnnamedDevicesAlone(t *testing.T) {
	handler, controller := newTestRouter(t)
	controller.SetAllPercentages([shelly.NumberOfDevices]uint8{10, 60, 40})

	rec := do(t, handler, http.MethodPost, "/power", `{"heater":{"percent":75}}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	want := [shelly.NumberOfDevices]uint8{75, 60, 40}
	if got := controller.GetAllPercentages(); got != want {
		t.Fatalf("expected %v, got %v", want, got)
	}
}

func TestFullCommandSetsEveryDevice(t *testing.T) {
	handler, controller := newTestRouter(t)
	controller.SetAllPercentages([shelly.NumberOfDevices]uint8{10, 60, 40})

	rec := do(t, handler, http.MethodPost, "/power",
		`{"heater":{"percent":1},"fan":{"percent":2},"humidifier":{"percent":3}}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	want := [shelly.NumberOfDevices]uint8{1, 2, 3}
	if got := controller.GetAllPercentages(); got != want {
		t.Fatalf("expected %v, got %v", want, got)
	}
}

func TestBulkCommandRejectsAnUnknownDevice(t *testing.T) {
	handler, controller := newTestRouter(t)
	controller.SetAllPercentages([shelly.NumberOfDevices]uint8{10, 20, 30})

	rec := do(t, handler, http.MethodPost, "/power", `{"kettle":{"percent":50}}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}

	// A rejected command must not have moved anything.
	want := [shelly.NumberOfDevices]uint8{10, 20, 30}
	if got := controller.GetAllPercentages(); got != want {
		t.Fatalf("expected %v to be untouched, got %v", want, got)
	}
}

func TestBulkCommandRejectsMalformedJSON(t *testing.T) {
	handler, _ := newTestRouter(t)

	rec := do(t, handler, http.MethodPost, "/power", `{"heater":`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestGetPercentageForOneDevice(t *testing.T) {
	handler, controller := newTestRouter(t)
	controller.SetAllPercentages([shelly.NumberOfDevices]uint8{0, 45, 0})

	rec := do(t, handler, http.MethodGet, "/power/fan", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var response types.APIResponse[types.PowerResponse]
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if response.Data.Percent != 45 {
		t.Fatalf("expected 45%%, got %d%%", response.Data.Percent)
	}
}

func TestGetPercentageRejectsAnUnknownDevice(t *testing.T) {
	handler, _ := newTestRouter(t)

	rec := do(t, handler, http.MethodGet, "/power/kettle", "")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

// This is the endpoint the control unit actually drives.
func TestSetPercentageForOneDeviceLeavesTheOthersAlone(t *testing.T) {
	handler, controller := newTestRouter(t)
	controller.SetAllPercentages([shelly.NumberOfDevices]uint8{10, 60, 40})

	rec := do(t, handler, http.MethodPost, "/power/heater", `{"percent":90}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	want := [shelly.NumberOfDevices]uint8{90, 60, 40}
	if got := controller.GetAllPercentages(); got != want {
		t.Fatalf("expected %v, got %v", want, got)
	}
}

func TestSetPercentageRejectsAnUnknownDevice(t *testing.T) {
	handler, _ := newTestRouter(t)

	rec := do(t, handler, http.MethodPost, "/power/kettle", `{"percent":90}`)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestSetPercentageRejectsMalformedJSON(t *testing.T) {
	handler, _ := newTestRouter(t)

	rec := do(t, handler, http.MethodPost, "/power/heater", `{"percent":`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

// A command is what holds the idle watchdog off, so it has to reach the
// controller even when it asks for the percentage already in force. Steps that
// hold a constant power (Heating sends heater=100 and fan=50 every tick) send
// nothing but repeats, and dropping those starves the watchdog.
//
// A fresh controller reports itself idle and sits at 0%, so a command for 0%
// is a repeat; the idle flag clearing is proof it was forwarded anyway.
func TestRepeatedCommandStillReachesTheController(t *testing.T) {
	t.Run("single device", func(t *testing.T) {
		handler, controller := newTestRouter(t)

		if !controller.IsIdle() {
			t.Fatal("expected a fresh controller to report itself idle")
		}

		rec := do(t, handler, http.MethodPost, "/power/heater", `{"percent":0}`)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}

		if controller.IsIdle() {
			t.Fatal("a repeated command was dropped instead of refreshing the watchdog")
		}
	})

	t.Run("bulk", func(t *testing.T) {
		handler, controller := newTestRouter(t)

		rec := do(t, handler, http.MethodPost, "/power", `{"heater":{"percent":0}}`)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}

		if controller.IsIdle() {
			t.Fatal("a repeated command was dropped instead of refreshing the watchdog")
		}
	})
}

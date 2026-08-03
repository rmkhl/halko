package engine

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/rmkhl/halko/types"
)

const invalid = float32(types.InvalidTemperatureReading)

func TestObserveHoldsLastValidReading(t *testing.T) {
	var temps fsmTemperatures

	temps.observe(temperatureReadings{Kiln: 100, Material: 50}, 1000)
	if temps.reading.Kiln != 100 || temps.reading.Material != 50 {
		t.Fatalf("after valid sample reading = %+v, want {50 100}", temps.reading)
	}

	// Both sensors fail: the previous values must survive.
	temps.observe(temperatureReadings{Kiln: invalid, Material: invalid}, 1010)
	if temps.reading.Kiln != 100 || temps.reading.Material != 50 {
		t.Errorf("after invalid sample reading = %+v, want values held at {50 100}", temps.reading)
	}
	if temps.kilnValidAt != 1000 || temps.materialValidAt != 1000 {
		t.Errorf("validAt = (%d, %d), want (1000, 1000) — invalid samples must not stamp",
			temps.kilnValidAt, temps.materialValidAt)
	}
}

func TestObserveTracksSensorsIndependently(t *testing.T) {
	var temps fsmTemperatures
	temps.observe(temperatureReadings{Kiln: 100, Material: 50}, 1000)

	// Only the kiln fails; material must keep updating.
	temps.observe(temperatureReadings{Kiln: invalid, Material: 60}, 1010)

	if temps.reading.Kiln != 100 {
		t.Errorf("Kiln = %v, want 100 held", temps.reading.Kiln)
	}
	if temps.reading.Material != 60 {
		t.Errorf("Material = %v, want 60", temps.reading.Material)
	}
	if temps.kilnValidAt != 1000 {
		t.Errorf("kilnValidAt = %d, want 1000", temps.kilnValidAt)
	}
	if temps.materialValidAt != 1010 {
		t.Errorf("materialValidAt = %d, want 1010", temps.materialValidAt)
	}
}

func TestInvalidForMeasuresFromProgramStartWhenNeverValid(t *testing.T) {
	var temps fsmTemperatures // no sample ever observed

	sensor, seconds := temps.invalidFor(1130, 1000)
	if seconds != 130 {
		t.Errorf("seconds = %d, want 130 (measured from program start)", seconds)
	}
	if sensor == "" {
		t.Error("sensor = empty, want the offending sensor named")
	}
}

func TestInvalidForResetsOnValidReading(t *testing.T) {
	var temps fsmTemperatures
	temps.observe(temperatureReadings{Kiln: 100, Material: 50}, 1100)

	if _, seconds := temps.invalidFor(1130, 1000); seconds != 30 {
		t.Errorf("seconds = %d, want 30", seconds)
	}
}

func TestInvalidForReportsTheWorseSensor(t *testing.T) {
	var temps fsmTemperatures
	temps.observe(temperatureReadings{Kiln: 100, Material: 50}, 1000)
	// Material keeps reporting, kiln does not.
	temps.observe(temperatureReadings{Kiln: invalid, Material: 60}, 1100)

	sensor, seconds := temps.invalidFor(1130, 1000)
	if sensor != "kiln" {
		t.Errorf("sensor = %q, want \"kiln\"", sensor)
	}
	if seconds != 130 {
		t.Errorf("seconds = %d, want 130", seconds)
	}
}

func TestExecuteTickFailsProgramWhenSensorInvalidTooLong(t *testing.T) {
	fsm := &programFSMController{
		state:               fsmStateWaiting,
		started:             1000,
		currentPSUStatus:    &fsmPSUStatus{},
		currentTemperatures: &fsmTemperatures{},
	}
	fsm.stateHandlers = map[fsmState]fsmStateHandler{
		fsmStateWaiting: &waitingStateHandler{fsm: fsm},
		fsmStateFailed:  &failedStateHandler{fsm: fsm},
	}
	// No valid reading has ever arrived and the threshold has passed.
	fsm.currentTemperatures.updated = 1000 + maxInvalidTemperatureSeconds + 1

	fsm.executeTickAt(fsm.started + maxInvalidTemperatureSeconds + 1)

	if fsm.state != fsmStateFailed {
		t.Errorf("state = %q, want %q", fsm.state, fsmStateFailed)
	}
}

func TestExecuteTickKeepsRunningWhileReadingsAreValid(t *testing.T) {
	fsm := &programFSMController{
		state:               fsmStateWaiting,
		started:             1000,
		currentPSUStatus:    &fsmPSUStatus{},
		currentTemperatures: &fsmTemperatures{},
	}
	fsm.stateHandlers = map[fsmState]fsmStateHandler{
		fsmStateWaiting: &waitingStateHandler{fsm: fsm},
		fsmStateFailed:  &failedStateHandler{fsm: fsm},
	}
	fsm.currentTemperatures.observe(temperatureReadings{Kiln: 100, Material: 50}, 1100)

	fsm.executeTickAt(1150)

	if fsm.state == fsmStateFailed {
		t.Error("state = failed, want the program still running")
	}
}

// TestExecuteTickFailsafeCutsAllPower drives the failsafe against a real
// psuController backed by an httptest server, proving the failed state does
// not just flip a status field but actually commands the heater, fan and
// humidifier off. A gutted shutdown() body would pass every other test in
// this file while leaving the kiln powered.
func TestExecuteTickFailsafeCutsAllPower(t *testing.T) {
	var mu sync.Mutex
	commanded := map[string]uint8{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		psu := strings.TrimPrefix(r.URL.Path, "/")
		var cmd PowerCommand
		if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
			t.Errorf("decoding power command for %q: %v", psu, err)
		}
		mu.Lock()
		commanded[psu] = cmd.Percent
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	psu := &psuController{client: server.Client(), powerControlURL: server.URL}

	fsm := &programFSMController{
		state:               fsmStateWaiting,
		started:             1000,
		psuController:       psu,
		currentPSUStatus:    &fsmPSUStatus{},
		currentTemperatures: &fsmTemperatures{},
	}
	fsm.stateHandlers = map[fsmState]fsmStateHandler{
		fsmStateWaiting: &waitingStateHandler{fsm: fsm},
		fsmStateFailed:  &failedStateHandler{fsm: fsm},
	}
	// No valid reading has ever arrived and the threshold has passed.
	fsm.currentTemperatures.updated = 1000 + maxInvalidTemperatureSeconds + 1

	fsm.executeTickAt(fsm.started + maxInvalidTemperatureSeconds + 1)

	if fsm.state != fsmStateFailed {
		t.Fatalf("state = %q, want %q", fsm.state, fsmStateFailed)
	}

	mu.Lock()
	defer mu.Unlock()
	for _, psuName := range []string{psuOven, psuFan, psuHumidifier} {
		percent, ok := commanded[psuName]
		if !ok {
			t.Errorf("psu %q was never commanded", psuName)
			continue
		}
		if percent != 0 {
			t.Errorf("psu %q commanded to %d%%, want 0%%", psuName, percent)
		}
	}
}

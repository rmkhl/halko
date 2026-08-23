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

// testSensorTimeoutSeconds mirrors the sensor_timeout the shipped config
// carries; the FSM now reads it from the defaults rather than a constant.
const testSensorTimeoutSeconds = 120

const invalid = float32(types.InvalidTemperatureReading)

func TestObserveHoldsLastValidReading(t *testing.T) {
	var temps fsmTemperatures

	temps.observe(temperatureReadings{KilnPrimary: 100, KilnSecondary: 100, Material: 50}, 1000, types.KilnSensorHigher)
	if temps.reading.Kiln != 100 || temps.reading.Material != 50 {
		t.Fatalf("after valid sample reading = %+v, want {50 100}", temps.reading)
	}

	// Both sensors fail: the previous values must survive.
	temps.observe(temperatureReadings{KilnPrimary: invalid, KilnSecondary: invalid, Material: invalid}, 1010, types.KilnSensorHigher)
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
	temps.observe(temperatureReadings{KilnPrimary: 100, KilnSecondary: 100, Material: 50}, 1000, types.KilnSensorHigher)

	// Only the kiln fails; material must keep updating.
	temps.observe(temperatureReadings{KilnPrimary: invalid, KilnSecondary: invalid, Material: 60}, 1010, types.KilnSensorHigher)

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
	temps.observe(temperatureReadings{KilnPrimary: 100, KilnSecondary: 100, Material: 50}, 1100, types.KilnSensorHigher)

	if _, seconds := temps.invalidFor(1130, 1000); seconds != 30 {
		t.Errorf("seconds = %d, want 30", seconds)
	}
}

func TestInvalidForReportsTheWorseSensor(t *testing.T) {
	var temps fsmTemperatures
	temps.observe(temperatureReadings{KilnPrimary: 100, KilnSecondary: 100, Material: 50}, 1000, types.KilnSensorHigher)
	// Material keeps reporting, kiln does not.
	temps.observe(temperatureReadings{KilnPrimary: invalid, KilnSecondary: invalid, Material: 60}, 1100, types.KilnSensorHigher)

	sensor, seconds := temps.invalidFor(1130, 1000)
	if sensor != sensorNameKiln {
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
		defaults:            &types.Defaults{SensorTimeoutSeconds: testSensorTimeoutSeconds},
	}
	fsm.stateHandlers = map[fsmState]fsmStateHandler{
		fsmStateWaiting: &waitingStateHandler{fsm: fsm},
		fsmStateFailed:  &failedStateHandler{fsm: fsm},
	}
	// No valid reading has ever arrived and the threshold has passed.
	fsm.currentTemperatures.updated = 1000 + testSensorTimeoutSeconds + 1

	fsm.executeTickAt(fsm.started + testSensorTimeoutSeconds + 1)

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
		defaults:            &types.Defaults{SensorTimeoutSeconds: testSensorTimeoutSeconds},
	}
	fsm.stateHandlers = map[fsmState]fsmStateHandler{
		fsmStateWaiting: &waitingStateHandler{fsm: fsm},
		fsmStateFailed:  &failedStateHandler{fsm: fsm},
	}
	fsm.currentTemperatures.observe(temperatureReadings{KilnPrimary: 100, KilnSecondary: 100, Material: 50}, 1100, types.KilnSensorHigher)

	fsm.executeTickAt(1150)

	if fsm.state == fsmStateFailed {
		t.Error("state = failed, want the program still running")
	}
}

// TestExecuteTickFailsafeCutsAllPower drives the failsafe against a real
// psuController backed by an httptest server, proving the failed state does
// not just flip a status field but actually commands the heater, fan and
// steam off. A gutted shutdown() body would pass every other test in
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
		defaults:            &types.Defaults{SensorTimeoutSeconds: testSensorTimeoutSeconds},
		currentPSUStatus:    &fsmPSUStatus{},
		currentTemperatures: &fsmTemperatures{},
	}
	fsm.stateHandlers = map[fsmState]fsmStateHandler{
		fsmStateWaiting: &waitingStateHandler{fsm: fsm},
		fsmStateFailed:  &failedStateHandler{fsm: fsm},
	}
	// No valid reading has ever arrived and the threshold has passed.
	fsm.currentTemperatures.updated = 1000 + testSensorTimeoutSeconds + 1

	fsm.executeTickAt(fsm.started + testSensorTimeoutSeconds + 1)

	if fsm.state != fsmStateFailed {
		t.Fatalf("state = %q, want %q", fsm.state, fsmStateFailed)
	}

	mu.Lock()
	defer mu.Unlock()
	for _, psuName := range []string{psuOven, psuFan, psuSteam} {
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

func TestResolveKilnWithBothSensorsValid(t *testing.T) {
	tests := []struct {
		strategy types.KilnSensorStrategy
		want     float32
	}{
		{types.KilnSensorLower, 98.0},
		{types.KilnSensorHigher, 100.0},
		{types.KilnSensorAverage, 99.0},
	}

	for _, test := range tests {
		if got := resolveKiln(test.strategy, 98.0, 100.0); got != test.want {
			t.Errorf("resolveKiln(%s, 98, 100) = %v, want %v", test.strategy, got, test.want)
		}
	}
}

// The whole point of two sensors: one failing leaves the run on the other,
// whatever the strategy says. An average of one reading is that reading.
func TestResolveKilnFallsBackToTheSurvivor(t *testing.T) {
	invalid := float32(types.InvalidTemperatureReading)

	for _, strategy := range []types.KilnSensorStrategy{
		types.KilnSensorLower, types.KilnSensorHigher, types.KilnSensorAverage,
	} {
		if got := resolveKiln(strategy, invalid, 99.0); got != 99.0 {
			t.Errorf("resolveKiln(%s, invalid, 99) = %v, want 99", strategy, got)
		}
		if got := resolveKiln(strategy, 98.0, invalid); got != 98.0 {
			t.Errorf("resolveKiln(%s, 98, invalid) = %v, want 98", strategy, got)
		}
	}
}

func TestResolveKilnWithNeitherSensorValid(t *testing.T) {
	invalid := float32(types.InvalidTemperatureReading)

	if got := resolveKiln(types.KilnSensorAverage, invalid, invalid); got != invalid {
		t.Errorf("resolveKiln with both invalid = %v, want the invalid sentinel", got)
	}
}

// The trap this design exists to avoid. observe holds the last valid reading
// so the sentinel never reaches the power controllers. If resolution ran after
// that hold, a primary that dropped out would still be averaged against a live
// secondary and the controller would act on a number no sensor ever read.
func TestObserveDoesNotAverageAStaleSensorWithALiveOne(t *testing.T) {
	invalid := float32(types.InvalidTemperatureReading)
	var temperatures fsmTemperatures

	temperatures.observe(temperatureReadings{
		Material: 40, KilnPrimary: 90, KilnSecondary: 100,
	}, 1000, types.KilnSensorAverage)
	if temperatures.reading.Kiln != 95 {
		t.Fatalf("resolved kiln = %v, want the average 95", temperatures.reading.Kiln)
	}

	// Primary drops out. The average must now be the secondary alone, not the
	// mean of a ten-minute-old 90 and a live 100.
	temperatures.observe(temperatureReadings{
		Material: 40, KilnPrimary: invalid, KilnSecondary: 100,
	}, 1600, types.KilnSensorAverage)

	if temperatures.reading.Kiln != 100 {
		t.Errorf("resolved kiln = %v, want 100 - the stale primary was averaged in",
			temperatures.reading.Kiln)
	}
}

// The raw readings are evidence, not control inputs: holding one would hide
// the dropout it exists to reveal.
func TestObserveDoesNotHoldTheRawReadings(t *testing.T) {
	invalid := float32(types.InvalidTemperatureReading)
	var temperatures fsmTemperatures

	temperatures.observe(temperatureReadings{
		Material: 40, KilnPrimary: 90, KilnSecondary: 100,
	}, 1000, types.KilnSensorHigher)
	temperatures.observe(temperatureReadings{
		Material: 40, KilnPrimary: invalid, KilnSecondary: 100,
	}, 1060, types.KilnSensorHigher)

	if temperatures.reading.KilnPrimary != invalid {
		t.Errorf("raw primary = %v, want the invalid sentinel to show through",
			temperatures.reading.KilnPrimary)
	}
	if temperatures.reading.Kiln != 100 {
		t.Errorf("resolved kiln = %v, want 100", temperatures.reading.Kiln)
	}
}

// One sensor down is a degraded run, not a failed one. Both down is what the
// timeout is for.
func TestObserveKeepsTheKilnValidWhileOneSensorLives(t *testing.T) {
	invalid := float32(types.InvalidTemperatureReading)
	var temperatures fsmTemperatures

	temperatures.observe(temperatureReadings{
		Material: 40, KilnPrimary: invalid, KilnSecondary: 100,
	}, 1000, types.KilnSensorHigher)

	sensor, seconds := temperatures.invalidFor(1000, 900)
	if sensor == sensorNameKiln && seconds > 0 {
		t.Errorf("kiln counted as stale for %ds while a sensor was reporting", seconds)
	}
}

func TestObserveLetsTheKilnGoStaleWhenBothSensorsFail(t *testing.T) {
	invalid := float32(types.InvalidTemperatureReading)
	var temperatures fsmTemperatures

	temperatures.observe(temperatureReadings{
		Material: 40, KilnPrimary: 90, KilnSecondary: 100,
	}, 1000, types.KilnSensorHigher)
	temperatures.observe(temperatureReadings{
		Material: 40, KilnPrimary: invalid, KilnSecondary: invalid,
	}, 1600, types.KilnSensorHigher)

	sensor, seconds := temperatures.invalidFor(1600, 900)
	if sensor != sensorNameKiln || seconds != 600 {
		t.Errorf("invalidFor = (%q, %d), want (kiln, 600)", sensor, seconds)
	}
}

package simulation

import (
	"testing"
	"time"

	"github.com/rmkhl/halko/simulator/elements"
	"github.com/rmkhl/halko/simulator/faults"
	"github.com/rmkhl/halko/simulator/physics"
	"github.com/rmkhl/halko/types"
)

// newTestResetter builds a resetter over real elements started at 20°C, with
// failure injection enabled so the reset's effect on it is observable.
func newTestResetter() (*Resetter, *physics.SimulationState, *faults.Injector) {
	wood := elements.NewWood(20.0, 20.0)
	heater := elements.NewHeater("kiln", 20.0, 20.0, wood)
	fan := elements.NewPower("Fan")
	humidifier := elements.NewPower("Humidifier")
	state := &physics.SimulationState{KilnTemp: 20.0, MaterialTemp: 20.0, EnvironmentTemp: 20.0}
	injector := faults.New(true)

	return &Resetter{
		Heater:              heater,
		Wood:                wood,
		Fan:                 fan,
		Humidifier:          humidifier,
		PhysicsState:        state,
		Faults:              injector,
		InitialKilnTemp:     20.0,
		InitialMaterialTemp: 20.0,
		EnvironmentTemp:     20.0,
	}, state, injector
}

func TestResetRestoresPhysicsState(t *testing.T) {
	r, state, _ := newTestResetter()

	state.KilnTemp = 95.0
	state.MaterialTemp = 60.0
	state.HeaterIsOn = true
	state.FanIsOn = true
	state.HumidifierIsOn = true

	r.Reset()

	if state.KilnTemp != 20.0 || state.MaterialTemp != 20.0 {
		t.Fatalf("expected temperatures back at 20°C, got kiln=%v material=%v", state.KilnTemp, state.MaterialTemp)
	}
	if state.HeaterIsOn || state.FanIsOn || state.HumidifierIsOn {
		t.Fatalf("expected all power off, got heater=%v fan=%v humidifier=%v",
			state.HeaterIsOn, state.FanIsOn, state.HumidifierIsOn)
	}
}

func TestResetDisarmsFailureInjection(t *testing.T) {
	r, _, injector := newTestResetter()

	start := time.Now()
	injector.Observe(90.0, 20.0, start) // arms the schedule

	r.Reset()

	// Well past every stage boundary: a disarmed injector still alters
	// nothing.
	probes := types.TemperatureResponse{"KilnPrimary": 50.0, "KilnSecondary": 50.0, "Wood": 30.0}
	injector.Apply(probes, start.Add(10*time.Minute))
	for name, value := range probes {
		if value == types.InvalidTemperatureReading {
			t.Fatalf("expected %s untouched after reset, got %v", name, value)
		}
	}
}

func TestOnDisplayMessageResetsOnlyOnTheTransitionToIdle(t *testing.T) {
	r, state, _ := newTestResetter()

	state.KilnTemp = 95.0
	r.OnDisplayMessage("Pre-Heat")
	if state.KilnTemp != 95.0 {
		t.Fatalf("expected no reset while running, got kiln=%v", state.KilnTemp)
	}

	r.OnDisplayMessage("idle")
	if state.KilnTemp != 20.0 {
		t.Fatalf("expected a reset on the transition to idle, got kiln=%v", state.KilnTemp)
	}

	// A second idle message is not a transition, so it must not reset again.
	state.KilnTemp = 42.0
	r.OnDisplayMessage("idle")
	if state.KilnTemp != 42.0 {
		t.Fatalf("expected no second reset while already idle, got kiln=%v", state.KilnTemp)
	}
}

func TestOnDisplayMessageResetsAgainAfterAnotherRun(t *testing.T) {
	r, state, _ := newTestResetter()

	r.OnDisplayMessage("idle")
	r.OnDisplayMessage("Pre-Heat")
	state.KilnTemp = 95.0
	r.OnDisplayMessage("idle")

	if state.KilnTemp != 20.0 {
		t.Fatalf("expected a reset when the next run ends, got kiln=%v", state.KilnTemp)
	}
}

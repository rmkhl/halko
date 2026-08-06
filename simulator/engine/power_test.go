package engine

import "testing"

// The state machine holds a requested change until the end of a 10-tick cycle,
// so a relay cannot be switched more often than that.
const ticksPerCycle = 10

func advance(p *Power, ticks int) {
	for range ticks {
		p.Tick()
	}
}

func assertState(t *testing.T, p *Power, wantRunning, wantOn bool) {
	t.Helper()

	running, on := p.Info()
	if running != wantRunning || on != wantOn {
		t.Fatalf("expected running=%v on=%v, got running=%v on=%v",
			wantRunning, wantOn, running, on)
	}
}

func TestNewPowerIsIdle(t *testing.T) {
	p := NewPower("heater")

	if p.IsRunning() {
		t.Fatal("expected a new power to be stopped")
	}
	assertState(t, p, false, false)
}

func TestStartAdoptsTheInitialState(t *testing.T) {
	tests := []struct {
		name         string
		initialState bool
	}{
		{"started on", true},
		{"started off", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPower("heater")
			p.Start(tt.initialState)

			if !p.IsRunning() {
				t.Fatal("expected the power to be running")
			}
			assertState(t, p, true, tt.initialState)
		})
	}
}

func TestSwitchToTakesEffectAtTheEndOfTheCycle(t *testing.T) {
	p := NewPower("heater")
	p.Start(false)
	p.SwitchTo(true)

	// The request is pending for the rest of the cycle.
	advance(p, ticksPerCycle-1)
	assertState(t, p, true, false)

	// The tick that closes the cycle applies it.
	p.Tick()
	assertState(t, p, true, true)
}

func TestSwitchToAppliesTheLatestRequestOnly(t *testing.T) {
	p := NewPower("heater")
	p.Start(false)

	// Several changes of mind inside one cycle collapse into the last one.
	p.SwitchTo(true)
	advance(p, 3)
	p.SwitchTo(false)
	advance(p, 3)
	p.SwitchTo(true)

	advance(p, ticksPerCycle-6)
	assertState(t, p, true, true)
}

func TestStateHoldsUntilAnotherSwitch(t *testing.T) {
	p := NewPower("heater")
	p.Start(false)
	p.SwitchTo(true)
	advance(p, ticksPerCycle)
	assertState(t, p, true, true)

	// Nothing new requested, so it stays on across further cycles.
	advance(p, ticksPerCycle*3)
	assertState(t, p, true, true)
}

func TestStopClearsEverything(t *testing.T) {
	p := NewPower("heater")
	p.Start(true)
	p.SwitchTo(true)
	advance(p, ticksPerCycle)

	p.Stop()

	if p.IsRunning() {
		t.Fatal("expected the power to be stopped")
	}
	assertState(t, p, false, false)
}

func TestTickOnAStoppedPowerIsANoOp(t *testing.T) {
	p := NewPower("heater")
	p.Start(true)
	p.Stop()

	// A pending request must not revive a stopped power.
	p.SwitchTo(true)
	advance(p, ticksPerCycle*2)
	assertState(t, p, false, false)
}

func TestRestartResetsTheCycle(t *testing.T) {
	p := NewPower("heater")
	p.Start(false)
	p.SwitchTo(true)
	advance(p, ticksPerCycle-1) // one tick short of applying it

	// Restarting drops the pending request and begins a fresh cycle.
	p.Start(false)
	p.Tick()
	assertState(t, p, true, false)

	advance(p, ticksPerCycle-1)
	assertState(t, p, true, false)
}

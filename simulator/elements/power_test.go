package elements

import "testing"

const ticksPerCycle = 10

func advance(p *Power, ticks int) {
	for range ticks {
		p.Tick()
	}
}

// CurrentCycle is what the Shelly API reports as the switch output.
func TestCurrentCycleReportsTheSwitchOutput(t *testing.T) {
	p := NewPower("heater")

	if got := p.CurrentCycle(); got != 0 {
		t.Fatalf("expected a stopped power to report 0, got %d", got)
	}

	p.TurnOn(true)
	if got := p.CurrentCycle(); got != 1 {
		t.Fatalf("expected a power started on to report 1, got %d", got)
	}

	p.SwitchTo(false)
	advance(p, ticksPerCycle)
	if got := p.CurrentCycle(); got != 0 {
		t.Fatalf("expected 0 after switching off, got %d", got)
	}
}

func TestPowerIsOnTracksTheRunningState(t *testing.T) {
	p := NewPower("fan")

	if p.IsOn() {
		t.Fatal("expected a new power to be stopped")
	}

	p.TurnOn(false)
	if !p.IsOn() {
		t.Fatal("expected the power to be running after TurnOn")
	}

	// IsOn reports whether the power is running, not whether it is energised.
	if got := p.CurrentCycle(); got != 0 {
		t.Fatalf("expected a power started off to report cycle 0, got %d", got)
	}

	p.TurnOff()
	if p.IsOn() {
		t.Fatal("expected the power to be stopped after TurnOff")
	}
}

func TestPowerName(t *testing.T) {
	if got := NewPower("humidifier").Name(); got != "humidifier" {
		t.Fatalf("expected name humidifier, got %q", got)
	}
}

func TestHeaterHoldsItsTemperature(t *testing.T) {
	wood := NewWood(20, 20)
	heater := NewHeater("heater", 25, 20, wood)

	if got := heater.Temperature(); got != 25 {
		t.Fatalf("expected 25, got %v", got)
	}
	if got := heater.GetMinTemp(); got != 20 {
		t.Fatalf("expected a minimum of 20, got %v", got)
	}

	// The physics engine owns the temperature; ticking only advances the
	// power state machine and must leave the reading alone.
	heater.TurnOn(true)
	heater.Tick()
	if got := heater.Temperature(); got != 25 {
		t.Fatalf("expected the tick to leave the temperature at 25, got %v", got)
	}

	heater.SetTemperature(60)
	if got := heater.Temperature(); got != 60 {
		t.Fatalf("expected 60, got %v", got)
	}
}

func TestWoodHoldsItsTemperature(t *testing.T) {
	wood := NewWood(22, 18)

	if got := wood.Temperature(); got != 22 {
		t.Fatalf("expected 22, got %v", got)
	}
	if got := wood.GetMinTemp(); got != 18 {
		t.Fatalf("expected a minimum of 18, got %v", got)
	}

	wood.SetTemperature(45)
	if got := wood.Temperature(); got != 45 {
		t.Fatalf("expected 45, got %v", got)
	}
}

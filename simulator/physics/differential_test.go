package physics

import "testing"

// A deliberately round configuration, so the expected temperatures below are
// arrived at by hand rather than by recording what the code happens to emit:
//
//	kiln change     = (heater - loss*(kiln-env) - transfer*(kiln-material)) / kiln mass
//	material change = (transfer*(kiln-material) - loss*(material-env)) / material mass
func arithmeticDifferentialConfig() map[string]interface{} {
	return map[string]interface{}{
		keyHeaterPower:             1000.0,
		keyHeatLossCoefficient:     10.0,
		keyHeatTransferCoefficient: 20.0,
		keyKilnThermalMass:         100.0,
		keyMaterialThermalMass:     200.0,
	}
}

func newDifferential(t *testing.T) SimulationEngine {
	t.Helper()

	engine, err := NewSimulationEngine(engineDifferential, arithmeticDifferentialConfig())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return engine
}

func TestDifferentialTickWithTheHeaterOn(t *testing.T) {
	engine := newDifferential(t)
	state := &SimulationState{KilnTemp: 50, MaterialTemp: 30, EnvironmentTemp: 20, HeaterIsOn: true}

	engine.Tick(state)

	// kiln:     (1000 - 10*30 - 20*20) / 100 = +3.0
	// material: (20*20 - 10*10) / 200        = +1.5
	if !nearly(state.KilnTemp, 53.0) {
		t.Fatalf("expected kiln 53.0, got %v", state.KilnTemp)
	}
	if !nearly(state.MaterialTemp, 31.5) {
		t.Fatalf("expected material 31.5, got %v", state.MaterialTemp)
	}
}

func TestDifferentialTickWithTheHeaterOff(t *testing.T) {
	engine := newDifferential(t)
	state := &SimulationState{KilnTemp: 50, MaterialTemp: 30, EnvironmentTemp: 20, HeaterIsOn: false}

	engine.Tick(state)

	// Only the heater term drops out: (0 - 300 - 400) / 100 = -7.0.
	if !nearly(state.KilnTemp, 43.0) {
		t.Fatalf("expected kiln 43.0, got %v", state.KilnTemp)
	}
	// The material still receives the transfer computed from the old kiln
	// temperature, so it is unchanged from the heater-on case.
	if !nearly(state.MaterialTemp, 31.5) {
		t.Fatalf("expected material 31.5, got %v", state.MaterialTemp)
	}
}

// Heat flows back into the kiln when the material is the hotter of the two.
func TestDifferentialTickWithTheMaterialHotterThanTheKiln(t *testing.T) {
	engine := newDifferential(t)
	state := &SimulationState{KilnTemp: 30, MaterialTemp: 50, EnvironmentTemp: 20, HeaterIsOn: false}

	engine.Tick(state)

	// transfer is negative, so the kiln gains despite the heater being off:
	// kiln:     (0 - 10*10 - 20*(-20)) / 100 = +3.0
	// material: (20*(-20) - 10*30) / 200     = -3.5
	if !nearly(state.KilnTemp, 33.0) {
		t.Fatalf("expected kiln 33.0, got %v", state.KilnTemp)
	}
	if !nearly(state.MaterialTemp, 46.5) {
		t.Fatalf("expected material 46.5, got %v", state.MaterialTemp)
	}
}

// The fan is not part of this engine's model; it must not shift the numbers.
func TestDifferentialTickIgnoresTheFan(t *testing.T) {
	withFan := newDifferential(t)
	withoutFan := newDifferential(t)

	fanState := &SimulationState{KilnTemp: 50, MaterialTemp: 30, EnvironmentTemp: 20, HeaterIsOn: true, FanIsOn: true}
	plainState := &SimulationState{KilnTemp: 50, MaterialTemp: 30, EnvironmentTemp: 20, HeaterIsOn: true}

	withFan.Tick(fanState)
	withoutFan.Tick(plainState)

	if fanState.KilnTemp != plainState.KilnTemp || fanState.MaterialTemp != plainState.MaterialTemp {
		t.Fatalf("expected the fan to make no difference, got %v/%v against %v/%v",
			fanState.KilnTemp, fanState.MaterialTemp, plainState.KilnTemp, plainState.MaterialTemp)
	}
}

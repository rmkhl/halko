package physics

import "testing"

// A round configuration so the expected temperatures can be derived by hand.
// air.volume is zero, which keeps the kiln's thermal capacity at exactly
// mass*specific_heat (1000 J/K) instead of varying with air density, and
// emissivity is a parameter so radiation can be switched out of the sum.
func arithmeticThermodynamicConfig(emissivity float64) map[string]interface{} {
	return map[string]interface{}{
		sectionKiln: map[string]interface{}{
			keyMass: 100.0, keySpecificHeat: 10.0, keySurfaceArea: 2.0,
			keyWallUValue: 5.0, keyEmissivity: emissivity,
		},
		sectionAir: map[string]interface{}{
			keyVolume: 0.0, keySpecificHeat: 1000.0,
		},
		sectionMaterial: map[string]interface{}{
			keyMass: 50.0, keySpecificHeat: 10.0, keySurfaceArea: 3.0,
		},
		sectionHeater: map[string]interface{}{
			keyWattage: 1000.0, keyEfficiency: 0.5,
		},
		sectionConvection: map[string]interface{}{
			keyNatural: 2.0, keyForced: 6.0, keyFanWasteHeat: 100.0,
		},
		sectionEnvironment: map[string]interface{}{
			keyTemperature: 20.0,
		},
		sectionPhysics: map[string]interface{}{
			keyStefanBoltzmann: 5.67e-8, keyTimeStep: 2.0,
		},
	}
}

func newThermodynamic(t *testing.T, emissivity float64) SimulationEngine {
	t.Helper()

	engine, err := NewSimulationEngine(engineThermodynamic, arithmeticThermodynamicConfig(emissivity))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return engine
}

// With radiation switched out, one tick is:
//
//	heater      = 1000*0.5*2                 = 1000 J
//	conduction  = 5*2*(40-20)*2              =  400 J
//	convection  = 2*3*(40-30)*2              =  120 J
//	kiln        = (1000-400-120) / 1000      = +0.48 °C
//	material    = (120 - 0.1*2*3*10*2) / 500 = +0.216 °C
func TestThermodynamicTickWithTheHeaterOn(t *testing.T) {
	engine := newThermodynamic(t, 0)
	state := &SimulationState{KilnTemp: 40, MaterialTemp: 30, EnvironmentTemp: 20, HeaterIsOn: true}

	engine.Tick(state)

	if !nearly(state.KilnTemp, 40.48) {
		t.Fatalf("expected kiln 40.48, got %v", state.KilnTemp)
	}
	if !nearly(state.MaterialTemp, 30.216) {
		t.Fatalf("expected material 30.216, got %v", state.MaterialTemp)
	}
}

func TestThermodynamicTickWithTheHeaterOff(t *testing.T) {
	engine := newThermodynamic(t, 0)
	state := &SimulationState{KilnTemp: 40, MaterialTemp: 30, EnvironmentTemp: 20, HeaterIsOn: false}

	engine.Tick(state)

	// Losing the 1000 J heater input costs the kiln a full degree.
	if !nearly(state.KilnTemp, 39.48) {
		t.Fatalf("expected kiln 39.48, got %v", state.KilnTemp)
	}
	if !nearly(state.MaterialTemp, 30.216) {
		t.Fatalf("expected material 30.216, got %v", state.MaterialTemp)
	}
}

// The fan does two things at once: it switches convection from the natural
// coefficient to the forced one, and its motor adds waste heat.
func TestThermodynamicTickWithTheFanOn(t *testing.T) {
	engine := newThermodynamic(t, 0)
	state := &SimulationState{
		KilnTemp: 40, MaterialTemp: 30, EnvironmentTemp: 20, HeaterIsOn: true, FanIsOn: true,
	}

	engine.Tick(state)

	// heater     = 1000 + 100*2 (fan waste)   = 1200 J
	// convection = 6*3*(40-30)*2              =  360 J
	// kiln       = (1200-400-360) / 1000      = +0.44 °C
	// material   = (360 - 0.1*6*3*10*2) / 500 = +0.648 °C
	if !nearly(state.KilnTemp, 40.44) {
		t.Fatalf("expected kiln 40.44, got %v", state.KilnTemp)
	}
	if !nearly(state.MaterialTemp, 30.648) {
		t.Fatalf("expected material 30.648, got %v", state.MaterialTemp)
	}

	// The fan moves more heat into the material than it costs the kiln.
	if state.MaterialTemp <= 30.216 {
		t.Fatalf("expected the fan to warm the material faster, got %v", state.MaterialTemp)
	}
}

func TestThermodynamicTickWithTheMaterialHotterThanTheKiln(t *testing.T) {
	engine := newThermodynamic(t, 0)
	state := &SimulationState{KilnTemp: 30, MaterialTemp: 50, EnvironmentTemp: 20, HeaterIsOn: false}

	engine.Tick(state)

	// Convection reverses, so the kiln gains while the material sheds heat
	// both to the kiln and to the environment.
	if !nearly(state.KilnTemp, 30.04) {
		t.Fatalf("expected kiln 30.04, got %v", state.KilnTemp)
	}
	if !nearly(state.MaterialTemp, 49.448) {
		t.Fatalf("expected material 49.448, got %v", state.MaterialTemp)
	}
}

// Radiation is the one term that is not linear in the temperature difference,
// so it is pinned as the difference it makes rather than folded into the sums
// above:
//
//	ε·σ·A·(T_kiln⁴ - T_env⁴)·Δt / capacity
//	= 1 · 5.67e-8 · 2 · (313.15⁴ - 293.15⁴) · 2 / 1000
//	≈ 0.506 °C
func TestThermodynamicRadiationCoolsTheKiln(t *testing.T) {
	withoutRadiation := newThermodynamic(t, 0)
	withRadiation := newThermodynamic(t, 1)

	plainState := &SimulationState{KilnTemp: 40, MaterialTemp: 30, EnvironmentTemp: 20, HeaterIsOn: true}
	radiatingState := &SimulationState{KilnTemp: 40, MaterialTemp: 30, EnvironmentTemp: 20, HeaterIsOn: true}

	withoutRadiation.Tick(plainState)
	withRadiation.Tick(radiatingState)

	drop := plainState.KilnTemp - radiatingState.KilnTemp
	if !nearly(drop, 0.506031) {
		t.Fatalf("expected radiation to cost the kiln 0.506031, got %v", drop)
	}

	// Radiation acts on the kiln alone; the material sees the same convection.
	if plainState.MaterialTemp != radiatingState.MaterialTemp {
		t.Fatalf("expected the material to be unaffected, got %v against %v",
			radiatingState.MaterialTemp, plainState.MaterialTemp)
	}
}

// Air is part of the kiln's thermal mass, so giving it a volume slows the
// kiln's response to the same energy input.
func TestThermodynamicAirVolumeAddsThermalMass(t *testing.T) {
	config := arithmeticThermodynamicConfig(0)
	config[sectionAir].(map[string]interface{})[keyVolume] = 1.0

	engine, err := NewSimulationEngine(engineThermodynamic, config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	state := &SimulationState{KilnTemp: 40, MaterialTemp: 30, EnvironmentTemp: 20, HeaterIsOn: true}
	engine.Tick(state)

	if state.KilnTemp >= 40.48 {
		t.Fatalf("expected the added air mass to slow the kiln below 40.48, got %v", state.KilnTemp)
	}
	if state.KilnTemp <= 40 {
		t.Fatalf("expected the kiln to still warm, got %v", state.KilnTemp)
	}
}

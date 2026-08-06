package physics

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// Valid configurations for each engine, matching the shipped .conf files.

func validSimpleConfig() map[string]interface{} {
	return map[string]interface{}{
		"heating_rate":  1.5,
		"cooling_rate":  0.5,
		"transfer_rate": 0.3,
	}
}

func validDifferentialConfig() map[string]interface{} {
	return map[string]interface{}{
		keyHeaterPower:              2000.0,
		"heat_loss_coefficient":     0.08,
		"heat_transfer_coefficient": 15.0,
		"kiln_thermal_mass":         70000.0,
		"material_thermal_mass":     34000.0,
	}
}

func validThermodynamicConfig() map[string]interface{} {
	return map[string]interface{}{
		sectionKiln: map[string]interface{}{
			keyMass: 140.0, keySpecificHeat: 500.0, keySurfaceArea: 4.5,
			"wall_u_value": 0.08, "emissivity": 0.08,
		},
		sectionAir: map[string]interface{}{
			"volume": 0.5, keySpecificHeat: 1005.0,
		},
		sectionMaterial: map[string]interface{}{
			keyMass: 20.0, keySpecificHeat: 1700.0, keySurfaceArea: 1.5,
		},
		sectionHeater: map[string]interface{}{
			"wattage": 2000.0, "efficiency": 0.95,
		},
		sectionConvection: map[string]interface{}{
			"natural": 5.0, "forced": 15.0, "fan_waste_heat": 50.0,
		},
		sectionEnvironment: map[string]interface{}{
			"temperature": 20.0,
		},
		sectionPhysics: map[string]interface{}{
			"stefan_boltzmann": 5.67e-8, keyTimeStep: 6.0,
		},
	}
}

func engineConfigs() map[string]func() map[string]interface{} {
	return map[string]func() map[string]interface{}{
		engineSimple:        validSimpleConfig,
		engineDifferential:  validDifferentialConfig,
		engineThermodynamic: validThermodynamicConfig,
	}
}

// clone deep-copies a config so mutations cannot leak between cases.
func clone(config map[string]interface{}) map[string]interface{} {
	copied := make(map[string]interface{}, len(config))
	for key, value := range config {
		if nested, ok := value.(map[string]interface{}); ok {
			copied[key] = clone(nested)
			continue
		}
		copied[key] = value
	}
	return copied
}

func TestNewSimulationEngineBuildsEveryKnownEngine(t *testing.T) {
	for name, config := range engineConfigs() {
		t.Run(name, func(t *testing.T) {
			engine, err := NewSimulationEngine(name, config())
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if engine.Name() != name {
				t.Fatalf("expected engine %q, got %q", name, engine.Name())
			}
		})
	}
}

func TestNewSimulationEngineRejectsAnUnknownName(t *testing.T) {
	if _, err := NewSimulationEngine("perpetual-motion", nil); err == nil {
		t.Fatal("expected an error for an unknown engine name")
	}
}

func TestNewSimulationEngineRejectsABadConfig(t *testing.T) {
	for name := range engineConfigs() {
		t.Run(name, func(t *testing.T) {
			if _, err := NewSimulationEngine(name, nil); err == nil {
				t.Fatalf("expected %s to reject a nil config", name)
			}
		})
	}
}

// mutations returns configs derived from a valid one by breaking a single
// field: removing it, giving it the wrong type, or zeroing/negating it.
func mutations(config map[string]interface{}) map[string]map[string]interface{} {
	out := make(map[string]map[string]interface{})

	// Collect every leaf path; one level of nesting is all the engines use.
	type leaf struct{ section, key string }
	var leaves []leaf
	for key, value := range config {
		nested, ok := value.(map[string]interface{})
		if !ok {
			leaves = append(leaves, leaf{"", key})
			continue
		}
		for nestedKey := range nested {
			leaves = append(leaves, leaf{key, nestedKey})
		}
	}

	set := func(target map[string]interface{}, l leaf, value interface{}, remove bool) {
		holder := target
		if l.section != "" {
			holder = target[l.section].(map[string]interface{})
		}
		if remove {
			delete(holder, l.key)
			return
		}
		holder[l.key] = value
	}

	for _, l := range leaves {
		name := l.key
		if l.section != "" {
			name = l.section + "." + l.key
		}

		for label, mutate := range map[string]struct {
			value  interface{}
			remove bool
		}{
			"missing":      {nil, true},
			"wrong type":   {"not a number", false},
			"zero":         {0.0, false},
			"negative":     {-1.0, false},
			"json integer": {json.Number("5"), false},
		} {
			mutated := clone(config)
			set(mutated, l, mutate.value, mutate.remove)
			out[name+" "+label] = mutated
		}
	}

	// Whole sections replaced by a scalar, for the nested engines.
	for key, value := range config {
		if _, ok := value.(map[string]interface{}); !ok {
			continue
		}
		mutated := clone(config)
		mutated[key] = "not an object"
		out[key+" is not an object"] = mutated

		removed := clone(config)
		delete(removed, key)
		out[key+" section missing"] = removed
	}

	return out
}

// The contract between the two halves of every engine: Initialize reads the
// config with unchecked type assertions, so anything ValidateConfig lets
// through has to be safe to initialize. A field added to one and not the other
// shows up here as a panic.
func TestInitializeNeverPanicsOnAnAcceptedConfig(t *testing.T) {
	for name, validConfig := range engineConfigs() {
		t.Run(name, func(t *testing.T) {
			for label, config := range mutations(validConfig()) {
				t.Run(label, func(_ *testing.T) {
					engine, err := NewSimulationEngine(name, config)
					if err != nil {
						return // Rejected, which is a fine outcome.
					}

					// Accepted, so it must also survive a tick.
					state := &SimulationState{KilnTemp: 20, MaterialTemp: 20, EnvironmentTemp: 20, HeaterIsOn: true}
					engine.Tick(state)
				})
			}
		})
	}
}

func TestValidateConfigRejectsAMissingField(t *testing.T) {
	tests := []struct {
		engine string
		config map[string]interface{}
	}{
		{engineSimple, map[string]interface{}{"heating_rate": 1.0, "cooling_rate": 1.0}},
		{engineDifferential, map[string]interface{}{"heater_power": 2000.0}},
		{engineThermodynamic, map[string]interface{}{sectionKiln: map[string]interface{}{}}},
	}

	for _, tt := range tests {
		t.Run(tt.engine, func(t *testing.T) {
			if _, err := NewSimulationEngine(tt.engine, tt.config); err == nil {
				t.Fatalf("expected %s to reject an incomplete config", tt.engine)
			}
		})
	}
}

// The shipped configs have to keep passing validation. time_step is injected by
// the simulator from controlunit.tick_length, so it is added here the same way.
func TestShippedConfigsValidate(t *testing.T) {
	entries, err := filepath.Glob(filepath.Join("..", "..", "simulator*.conf"))
	if err != nil {
		t.Fatalf("failed to look for configs: %v", err)
	}
	if len(entries) == 0 {
		t.Skip("no simulator configs found")
	}

	for _, path := range entries {
		t.Run(filepath.Base(path), func(t *testing.T) {
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("failed to read %s: %v", path, err)
			}

			var parsed struct {
				SimulationEngine string                 `json:"simulation_engine"`
				EngineConfig     map[string]interface{} `json:"engine_config"`
			}
			if err := json.Unmarshal(data, &parsed); err != nil {
				t.Fatalf("failed to parse %s: %v", path, err)
			}

			if parsed.EngineConfig[sectionPhysics] == nil {
				parsed.EngineConfig[sectionPhysics] = make(map[string]interface{})
			}
			if physics, ok := parsed.EngineConfig[sectionPhysics].(map[string]interface{}); ok {
				physics[keyTimeStep] = 6.0
			}

			if _, err := NewSimulationEngine(parsed.SimulationEngine, parsed.EngineConfig); err != nil {
				t.Fatalf("%s no longer builds an engine: %v", path, err)
			}
		})
	}
}

// Every engine has to settle rather than run away, and must never take a
// temperature below ambient.
func TestEnginesStayAboveAmbientWithTheHeaterOff(t *testing.T) {
	for name, config := range engineConfigs() {
		t.Run(name, func(t *testing.T) {
			engine, err := NewSimulationEngine(name, config())
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			state := &SimulationState{
				KilnTemp: 80, MaterialTemp: 70, EnvironmentTemp: 20, HeaterIsOn: false,
			}

			for tick := range 5000 {
				engine.Tick(state)

				if state.KilnTemp < state.EnvironmentTemp {
					t.Fatalf("tick %d: kiln fell below ambient: %v", tick, state.KilnTemp)
				}
				if state.MaterialTemp < state.EnvironmentTemp {
					t.Fatalf("tick %d: material fell below ambient: %v", tick, state.MaterialTemp)
				}
			}
		})
	}
}

func TestEnginesHeatTheKilnWhenTheHeaterIsOn(t *testing.T) {
	for name, config := range engineConfigs() {
		t.Run(name, func(t *testing.T) {
			engine, err := NewSimulationEngine(name, config())
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			state := &SimulationState{
				KilnTemp: 20, MaterialTemp: 20, EnvironmentTemp: 20, HeaterIsOn: true,
			}
			start := state.KilnTemp

			for range 20 {
				engine.Tick(state)
			}

			if state.KilnTemp <= start {
				t.Fatalf("expected the kiln to warm from %v, got %v", start, state.KilnTemp)
			}
		})
	}
}

func TestSimpleEngineTransfersHeatTowardsTheMaterial(t *testing.T) {
	engine, err := NewSimulationEngine(engineSimple, validSimpleConfig())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	state := &SimulationState{KilnTemp: 60, MaterialTemp: 20, EnvironmentTemp: 20}
	engine.Tick(state)

	// transfer_rate is applied as a fixed step whenever the kiln is warmer.
	if want := float32(20.3); state.MaterialTemp != want {
		t.Fatalf("expected material at %v, got %v", want, state.MaterialTemp)
	}
}

// Characterisation: the simple engine moves the material by a fixed
// transfer_rate step rather than by a fraction of the gap, so a gap narrower
// than the step overshoots instead of converging.
func TestSimpleEngineOvershootsAGapNarrowerThanItsStep(t *testing.T) {
	engine, err := NewSimulationEngine(engineSimple, validSimpleConfig())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Tick updates the kiln first: with the heater off it cools by 0.5 to 30.1,
	// leaving a 0.1° gap above the material.
	state := &SimulationState{KilnTemp: 30.6, MaterialTemp: 30.0, EnvironmentTemp: 20}
	engine.Tick(state)

	if !nearly(state.KilnTemp, 30.1) {
		t.Fatalf("expected the kiln to cool to 30.1, got %v", state.KilnTemp)
	}

	// The material still moves a full 0.3° step and ends up above the kiln.
	if !nearly(state.MaterialTemp, 30.3) {
		t.Fatalf("expected the material at 30.3, got %v", state.MaterialTemp)
	}
	if state.MaterialTemp <= state.KilnTemp {
		t.Fatalf("expected the material (%v) to overshoot the kiln (%v)",
			state.MaterialTemp, state.KilnTemp)
	}
}

func nearly(got, want float32) bool {
	const tolerance = 0.001
	diff := got - want
	return diff < tolerance && diff > -tolerance
}

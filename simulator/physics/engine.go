package physics

import (
	"fmt"
)

// SimulationEngine defines the interface for physics simulation engines
type SimulationEngine interface {
	// Initialize sets up the engine with configuration
	Initialize(config map[string]interface{}) error

	// ValidateConfig checks if the engine configuration is valid
	ValidateConfig(config map[string]interface{}) error

	// Tick performs one physics simulation step
	Tick(state *SimulationState)

	// Name returns the engine name for logging
	Name() string
}

// SimulationState holds the current state of all simulated elements
type SimulationState struct {
	// Oven/heater state
	OvenTemp        float32
	HeaterIsOn      bool
	EnvironmentTemp float32

	// Material/wood state
	MaterialTemp float32

	// Fan state (for future use in heat transfer)
	FanIsOn bool

	// Humidifier state (for future expansion)
	HumidifierIsOn bool
}

// NewSimulationEngine creates a simulation engine by name
func NewSimulationEngine(engineName string, config map[string]interface{}) (SimulationEngine, error) {
	var engine SimulationEngine

	switch engineName {
	case "simple":
		engine = &SimpleSimulation{}
	case "differential":
		engine = &DifferentialSimulation{}
	case "thermodynamic":
		engine = &ThermodynamicSimulation{}
	default:
		return nil, fmt.Errorf("unknown simulation engine: %s", engineName)
	}

	// Validate config before initialization
	if err := engine.ValidateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid config for %s engine: %w", engineName, err)
	}

	// Initialize the engine
	if err := engine.Initialize(config); err != nil {
		return nil, fmt.Errorf("failed to initialize %s engine: %w", engineName, err)
	}

	return engine, nil
}

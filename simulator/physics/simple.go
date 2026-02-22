package physics

import (
	"errors"
	"fmt"

	"github.com/rmkhl/halko/types/log"
)

// SimpleSimulation implements the current fixed-rate temperature simulation
type SimpleSimulation struct {
	heatingRate  float32 // Temperature increase per tick when heater is on
	coolingRate  float32 // Temperature decrease per tick when heater is off
	transferRate float32 // Heat transfer rate between oven and material
}

func (s *SimpleSimulation) Name() string {
	return "simple"
}

func (s *SimpleSimulation) ValidateConfig(config map[string]interface{}) error {
	// Check for required fields with proper types
	if config == nil {
		return errors.New("config cannot be nil")
	}

	// Validate heating_rate
	heatingRate, ok := config["heating_rate"].(float64)
	if !ok {
		return errors.New("heating_rate is required and must be a number")
	}
	if heatingRate <= 0 {
		return fmt.Errorf("heating_rate must be positive, got: %f", heatingRate)
	}

	// Validate cooling_rate
	coolingRate, ok := config["cooling_rate"].(float64)
	if !ok {
		return errors.New("cooling_rate is required and must be a number")
	}
	if coolingRate <= 0 {
		return fmt.Errorf("cooling_rate must be positive, got: %f", coolingRate)
	}

	// Validate transfer_rate
	transferRate, ok := config["transfer_rate"].(float64)
	if !ok {
		return errors.New("transfer_rate is required and must be a number")
	}
	if transferRate <= 0 {
		return fmt.Errorf("transfer_rate must be positive, got: %f", transferRate)
	}

	return nil
}

func (s *SimpleSimulation) Initialize(config map[string]interface{}) error {
	// Config already validated, safe to extract
	s.heatingRate = float32(config["heating_rate"].(float64))
	s.coolingRate = float32(config["cooling_rate"].(float64))
	s.transferRate = float32(config["transfer_rate"].(float64))

	log.Info("Simple simulation initialized: heating=%.2f°C/tick, cooling=%.2f°C/tick, transfer=%.2f°C/tick",
		s.heatingRate, s.coolingRate, s.transferRate)
	return nil
}

func (s *SimpleSimulation) Tick(state *SimulationState) {
	log.Debug("Simulation[simple] tick - Heater:%v Fan:%v Humidifier:%v", state.HeaterIsOn, state.FanIsOn, state.HumidifierIsOn)

	// Update oven temperature based on heater state
	oldOvenTemp := state.OvenTemp
	if state.HeaterIsOn {
		state.OvenTemp += s.heatingRate
		log.Debug("Simulation[simple]: Oven heating %.1f°C -> %.1f°C", oldOvenTemp, state.OvenTemp)
	} else {
		state.OvenTemp = max(state.EnvironmentTemp, state.OvenTemp-s.coolingRate)
		if state.OvenTemp != oldOvenTemp {
			log.Debug("Simulation[simple]: Oven cooling %.1f°C -> %.1f°C (min: %.1f°C)",
				oldOvenTemp, state.OvenTemp, state.EnvironmentTemp)
		}
	}

	// Update material temperature based on oven temperature
	oldMaterialTemp := state.MaterialTemp
	delta := state.OvenTemp - state.MaterialTemp

	var effectiveDelta float32
	switch {
	case delta > 0.0:
		effectiveDelta = s.transferRate
	case delta < 0.0:
		effectiveDelta = -s.transferRate
	}

	state.MaterialTemp = max(state.EnvironmentTemp, state.MaterialTemp+effectiveDelta)
	if state.MaterialTemp != oldMaterialTemp {
		log.Debug("Simulation[simple]: Material %.1f°C -> %.1f°C (oven: %.1f°C, delta: %.2f°C)",
			oldMaterialTemp, state.MaterialTemp, state.OvenTemp, delta)
	}
}

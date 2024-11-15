package types

import (
	"fmt"
	"time"
)

const (
	StepTypeHeating   StepType = "heating"
	StepTypeCooling   StepType = "cooling"
	StepTypeAcclimate StepType = "acclimate"
)

type (
	StepType string

	PowerSetting struct {
		Power int8 `json:"power"`
	}

	PidSettings struct {
		Kp float64 `json:"kp"`
		Ki float64 `json:"ki"`
		Kd float64 `json:"kd"`
	}

	PowerPidSettings struct {
		MaxDelta uint8        `json:"max_delta,omitempty"`
		Power    uint8        `json:"power,omitempty"`
		Pid      *PidSettings `json:"pid,omitempty"`
	}

	ProgramStep struct {
		Name              string           `json:"name"`
		StepType          StepType         `json:"type"`
		TargetTemperature uint8            `json:"temperature_target"`
		Duration          *time.Duration   `json:"duration,omitempty"`
		Heater            PowerPidSettings `json:"heater"`
		Fan               PowerSetting     `json:"fan"`
		Humidifier        PowerSetting     `json:"humidifier"`
	}

	Program struct {
		ProgramName  string        `json:"name"`
		ProgramSteps []ProgramStep `json:"steps"`
	}
)

func (p *ProgramStep) Validate() error {
	// Do some rudimentary validation for different step types, purposefully not "optimized" for code brevity
	//
	// For all steps the power setting for the fan and humidifier must be set between 0 and 100.
	if p.Fan.Power < 0 || p.Fan.Power > 100 {
		return fmt.Errorf("fan power must be between 0 and 100")
	}
	if p.Humidifier.Power < 0 || p.Humidifier.Power > 100 {
		return fmt.Errorf("humidifier power must be between 0 and 100")
	}

	// Target temperature cannot be above 200
	if p.TargetTemperature > 200 {
		return fmt.Errorf("target temperature must be between 0 and 200")
	}

	// For heating step the PowerPidSettings must be set and it needs to have a max delta or constant power
	if p.StepType == StepTypeHeating {
		if p.Heater.MaxDelta == 0 && p.Heater.Power == 0 {
			return fmt.Errorf("max delta or constant power must be set for heating step")
		}
		if p.TargetTemperature == 0 {
			return fmt.Errorf("target temperature must be set for heating step")
		}
	}

	// For acclimate step the duration, heater power and target temperature needs to be set
	if p.StepType == StepTypeAcclimate {
		if p.Duration == nil {
			return fmt.Errorf("duration must be set for acclimate step")
		}
		if p.Heater.MaxDelta == 0 && p.Heater.Power == 0 {
			return fmt.Errorf("heater must be set for acclimate step, either max delta or power")
		}
		if p.TargetTemperature == 0 {
			return fmt.Errorf("target temperature must be set for acclimate step")
		}
	}

	// For cooling step either the duration or target temperature needs to be set
	if p.StepType == StepTypeCooling {
		if p.Duration == nil && p.TargetTemperature == 0 {
			return fmt.Errorf("either duration or target temperature must be set for cooling step")
		}
		// If target temperature is set, it must be below 50
		if p.TargetTemperature > 50 {
			return fmt.Errorf("target temperature must be below 50 for cooling step")
		}
	}

	return nil
}

func (p *Program) Validate() error {
	for _, step := range p.ProgramSteps {
		err := step.Validate()
		if err != nil {
			return err
		}
	}
	return nil
}

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
		MaxDelta *uint8       `json:"max_delta,omitempty"`
		Power    *uint8       `json:"power,omitempty"`
		Pid      *PidSettings `json:"pid,omitempty"`
	}

	ProgramStep struct {
		Name              string           `json:"name"`
		StepType          StepType         `json:"type"`
		TargetTemperature uint             `json:"temperature_target"`
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
	if p.StepType == StepTypeHeating || p.StepType == StepTypeAcclimate {
		if p.Heater.Power != nil {
			return fmt.Errorf("setting power value for heater %s step is not allowed", p.StepType)
		}
		if p.Heater.MaxDelta == nil {
			return fmt.Errorf("setting max delta value for heater %s in step required", p.StepType)
		}
	}
	if p.StepType == StepTypeCooling {
		if p.Heater.Power == nil {
			return fmt.Errorf("sower must be off for cooling")
		}
		if p.Heater.MaxDelta == nil {
			return fmt.Errorf("setting max delta value for heater in cooling is not allowed")
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

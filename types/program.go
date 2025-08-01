package types

import (
	"errors"
	"time"
)

const (
	StepTypeHeating   StepType = "heating"
	StepTypeCooling   StepType = "cooling"
	StepTypeAcclimate StepType = "acclimate"

	PowerSettingTypeSimple PowerSettingType = "simple"
	PowerSettingTypeDelta  PowerSettingType = "delta"
	PowerSettingTypePid    PowerSettingType = "pid"
)

type (
	StepType         string
	PowerSettingType string

	PidSettings struct {
		Kp float32 `json:"kp"`
		Ki float32 `json:"ki"`
		Kd float32 `json:"kd"`
	}

	PowerPidSettings struct {
		Type     PowerSettingType `json:"type,omitempty"`
		MinDelta *float32         `json:"min_delta,omitempty"`
		MaxDelta *float32         `json:"max_delta,omitempty"`
		Power    *uint8           `json:"power,omitempty"`
		Pid      *PidSettings     `json:"pid,omitempty"`
	}

	ProgramStep struct {
		Name              string            `json:"name"`
		StepType          StepType          `json:"type"`
		TargetTemperature uint8             `json:"temperature_target,omitempty"`
		Runtime           *time.Duration    `json:"runtime,omitempty"`
		Heater            *PowerPidSettings `json:"heater,omitempty"`
		Fan               *PowerPidSettings `json:"fan,omitempty"`
		Humidifier        *PowerPidSettings `json:"humidifier,omitempty"`
	}

	Program struct {
		ProgramName     string        `json:"name"`
		ProgramSteps    []ProgramStep `json:"steps"`
		DefaultsApplied bool          `json:"-"`
	}
)

func isValidPercentage(value uint8) bool {
	return value <= 100
}

func (p *PowerPidSettings) Validate(component string) error {
	hasDeltas := p.MinDelta != nil || p.MaxDelta != nil

	controlMethods := 0
	if p.Power != nil {
		controlMethods++
		p.Type = PowerSettingTypeSimple
	}
	if hasDeltas {
		controlMethods++
		p.Type = PowerSettingTypeDelta
	}
	if p.Pid != nil {
		controlMethods++
		p.Type = PowerSettingTypePid
	}

	if controlMethods != 1 {
		return errors.New(component + " must define exactly one control method: power, min/max deltas, or PID")
	}

	if hasDeltas && (p.MinDelta == nil || p.MaxDelta == nil) {
		return errors.New(component + " min and max delta must both be defined")
	}

	if p.Power != nil && !isValidPercentage(*p.Power) {
		return errors.New(component + " power must be between 0 and 100")
	}

	return nil
}

func (p *ProgramStep) Validate() error {
	// Validate fan - call validate first to capture any errors
	fanErr := p.Fan.Validate("fan")
	if p.Fan.Type != PowerSettingTypeSimple {
		return errors.New("fan must use simple power control")
	}
	if fanErr != nil {
		return fanErr
	}

	// Validate humidifier - call validate first to capture any errors
	humidifierErr := p.Humidifier.Validate("humidifier")
	if p.Humidifier.Type != PowerSettingTypeSimple {
		return errors.New("humidifier must use simple power control")
	}
	if humidifierErr != nil {
		return humidifierErr
	}

	switch p.StepType {
	case StepTypeHeating:
		return p.validateHeatingStep()
	case StepTypeAcclimate:
		return p.validateAcclimateStep()
	case StepTypeCooling:
		return p.validateCoolingStep()
	default:
		return errors.New("unknown step type")
	}
}

func (p *ProgramStep) validateHeatingStep() error {
	if p.Runtime != nil {
		return errors.New("heating step cannot have runtime")
	}
	return p.Heater.Validate("heater")
}

func (p *ProgramStep) validateAcclimateStep() error {
	if p.Runtime == nil {
		return errors.New("acclimate step must have runtime")
	}
	return p.Heater.Validate("heater")
}

func (p *ProgramStep) validateCoolingStep() error {
	if p.Runtime != nil {
		return errors.New("cooling step cannot have runtime")
	}
	// Validate heater - call validate first to capture any errors
	heaterErr := p.Heater.Validate("heater")
	if p.Heater.Type != PowerSettingTypeSimple {
		return errors.New("cooling step heater must use simple power control")
	}
	if heaterErr != nil {
		return heaterErr
	}
	return nil
}

func (p *Program) ApplyDefaults(defaults *Defaults) {
	zeroPower := uint8(0)

	for i := range p.ProgramSteps {
		step := &p.ProgramSteps[i]

		if step.Heater == nil {
			step.Heater = &PowerPidSettings{}
		}

		if step.Heater.Pid == nil && step.Heater.Power == nil &&
			step.Heater.MinDelta == nil && step.Heater.MaxDelta == nil {
			switch step.StepType {
			case StepTypeAcclimate:
				defaultPid := defaults.PidSettings[StepTypeAcclimate]
				step.Heater.Pid = &PidSettings{
					Kp: defaultPid.Kp,
					Ki: defaultPid.Ki,
					Kd: defaultPid.Kd,
				}
			case StepTypeHeating:
				step.Heater.MinDelta = &defaults.MinDeltaHeating
				step.Heater.MaxDelta = &defaults.MaxDeltaHeating
			case StepTypeCooling:
				step.Heater.Power = &zeroPower
			}
		}

		if step.Fan == nil {
			step.Fan = &PowerPidSettings{}
		}
		if step.Fan.Power == nil {
			step.Fan.Power = &zeroPower
		}

		if step.Humidifier == nil {
			step.Humidifier = &PowerPidSettings{}
		}
		if step.Humidifier.Power == nil {
			step.Humidifier.Power = &zeroPower
		}
	}

	p.DefaultsApplied = true
}

func (p *Program) Validate() error {
	if !p.DefaultsApplied {
		return errors.New("defaults must be applied before validation")
	}

	for _, step := range p.ProgramSteps {
		err := step.Validate()
		if err != nil {
			return err
		}
	}

	return p.validateStepOrderAndTemperatureProgression()
}

func (p *Program) validateStepOrderAndTemperatureProgression() error {
	if len(p.ProgramSteps) <= 2 {
		return errors.New("program must have at least two step")
	}

	if p.ProgramSteps[0].StepType != StepTypeHeating {
		return errors.New("first step must be a heating step")
	}

	if p.ProgramSteps[len(p.ProgramSteps)-1].StepType != StepTypeCooling {
		return errors.New("last step must be a cooling step")
	}

	for i := 0; i < len(p.ProgramSteps)-1; i++ {
		currentStep := p.ProgramSteps[i]

		if currentStep.TargetTemperature > 200 {
			return errors.New("target temperature must not exceed 200 degrees")
		}

		nextStep := p.ProgramSteps[i+1]

		switch nextStep.StepType {
		case StepTypeHeating:
			if nextStep.TargetTemperature <= currentStep.TargetTemperature {
				return errors.New("heating step temperature must be higher than previous step")
			}
		case StepTypeAcclimate:
			if nextStep.TargetTemperature < currentStep.TargetTemperature {
				return errors.New("acclimate step temperature must be greater than or equal to previous step")
			}
		case StepTypeCooling:
			if nextStep.TargetTemperature >= currentStep.TargetTemperature {
				return errors.New("cooling step temperature must be lower than previous step")
			}
		}
	}

	return nil
}

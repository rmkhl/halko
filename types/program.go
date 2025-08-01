package types

import (
	"errors"
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
		Power *uint8 `json:"power,omitempty"`
	}

	PidSettings struct {
		Kp float32 `json:"kp"`
		Ki float32 `json:"ki"`
		Kd float32 `json:"kd"`
	}

	PowerPidSettings struct {
		MinDelta *float32     `json:"min_delta,omitempty"`
		MaxDelta *float32     `json:"max_delta,omitempty"`
		Power    *uint8       `json:"power,omitempty"`
		Pid      *PidSettings `json:"pid,omitempty"`
	}

	ProgramStep struct {
		Name              string           `json:"name"`
		StepType          StepType         `json:"type"`
		TargetTemperature uint8            `json:"temperature_target,omitempty"`
		Runtime           *time.Duration   `json:"runtime,omitempty"`
		Heater            PowerPidSettings `json:"heater"`
		Fan               PowerSetting     `json:"fan"`
		Humidifier        PowerSetting     `json:"humidifier"`
	}

	Program struct {
		ProgramName  string        `json:"name"`
		ProgramSteps []ProgramStep `json:"steps"`
	}
)

// PowerSetting helper methods
func (ps *PowerSetting) IsDefined() bool {
	return ps.Power != nil
}

func (ps *PowerSetting) GetPower() uint8 {
	if ps.Power == nil {
		return 0
	}
	return *ps.Power
}

func (ps *PowerSetting) SetPower(power uint8) {
	ps.Power = &power
}

// PowerPidSettings helper methods
func (pps *PowerPidSettings) IsPowerDefined() bool {
	return pps.Power != nil
}

func (pps *PowerPidSettings) IsMinDeltaDefined() bool {
	return pps.MinDelta != nil
}

func (pps *PowerPidSettings) IsMaxDeltaDefined() bool {
	return pps.MaxDelta != nil
}

func (pps *PowerPidSettings) GetPower() uint8 {
	if pps.Power == nil {
		return 0
	}
	return *pps.Power
}

func (pps *PowerPidSettings) GetMinDelta() float32 {
	if pps.MinDelta == nil {
		return 0
	}
	return *pps.MinDelta
}

func (pps *PowerPidSettings) GetMaxDelta() float32 {
	if pps.MaxDelta == nil {
		return 0
	}
	return *pps.MaxDelta
}

func (pps *PowerPidSettings) SetPower(power uint8) {
	pps.Power = &power
}

func (pps *PowerPidSettings) SetMinDelta(delta float32) {
	pps.MinDelta = &delta
}

func (pps *PowerPidSettings) SetMaxDelta(delta float32) {
	pps.MaxDelta = &delta
}

func (p *ProgramStep) Validate() error {
	// Do some rudimentary validation for different step types, purposefully not "optimized" for code brevity
	//
	// For all steps the power setting for the fan and humidifier must be set between 0 and 100.
	if p.Fan.Power != nil && *p.Fan.Power > 100 {
		return errors.New("fan power must be between 0 and 100")
	}
	if p.Humidifier.Power != nil && *p.Humidifier.Power > 100 {
		return errors.New("humidifier power must be between 0 and 100")
	}

	// Target temperature cannot be above 200
	if p.TargetTemperature > 200 {
		return errors.New("target temperature must be between 0 and 200")
	}

	// For heating step: require temperature_target (no runtime) and heater must have power (no PID)
	if p.StepType == StepTypeHeating {
		if p.TargetTemperature == 0 {
			return errors.New("target temperature must be set for heating step")
		}
		if p.Runtime != nil {
			return errors.New("runtime cannot be set for heating step")
		}
		if p.Heater.Power == nil || *p.Heater.Power == 0 {
			return errors.New("heater power must be set for heating step")
		}
		if p.Heater.Pid != nil {
			return errors.New("pid cannot be set for heating step")
		}
	}

	// For acclimate step: require both temperature_target and runtime
	// nolint:nestif  // I'll take the blame on this and refactor it later
	if p.StepType == StepTypeAcclimate {
		if p.TargetTemperature == 0 {
			return errors.New("target temperature must be set for acclimate step")
		}
		if p.Runtime == nil {
			return errors.New("runtime must be set for acclimate step")
		}
		// When using PID, min_delta and max_delta cannot be defined
		if p.Heater.Pid != nil {
			if (p.Heater.MinDelta != nil && *p.Heater.MinDelta != 0) || (p.Heater.MaxDelta != nil && *p.Heater.MaxDelta != 0) {
				return errors.New("min_delta and max_delta cannot be set when using pid")
			}
			// If PID is not empty, all properties must be present
			if p.Heater.Pid.Kp == 0 && (p.Heater.Pid.Ki != 0 || p.Heater.Pid.Kd != 0) {
				return errors.New("all pid properties (kp, ki, kd) must be set if any are specified")
			}
			if p.Heater.Pid.Ki == 0 && (p.Heater.Pid.Kp != 0 || p.Heater.Pid.Kd != 0) {
				return errors.New("all pid properties (kp, ki, kd) must be set if any are specified")
			}
			if p.Heater.Pid.Kd == 0 && (p.Heater.Pid.Kp != 0 || p.Heater.Pid.Ki != 0) {
				return errors.New("all pid properties (kp, ki, kd) must be set if any are specified")
			}
		}
	}

	// For cooling step: require either temperature_target or runtime
	if p.StepType == StepTypeCooling {
		if p.Runtime == nil && p.TargetTemperature == 0 {
			return errors.New("either runtime or target temperature must be set for cooling step")
		}
		// If target temperature is set, it must be below 50
		if p.TargetTemperature > 50 {
			return errors.New("target temperature must be below 50 for cooling step")
		}
	}

	return nil
}

func (p *Program) Validate() error {
	// Validate each step
	for _, step := range p.ProgramSteps {
		err := step.Validate()
		if err != nil {
			return err
		}
	}

	// Validate the logical order of steps
	return p.validateStepOrder()
}

// validateStepOrder ensures that steps follow a logical temperature progression:
// 1. First step must be heating
// 2. For heating steps: temperature must be higher than previous step
// 3. For acclimate steps: temperature must be >= previous step
// 4. For cooling steps: temperature must be lower than previous step
// 5. Program must end with a cooling step
func (p *Program) validateStepOrder() error {
	if len(p.ProgramSteps) <= 2 {
		return errors.New("program must have at least two step")
	}

	// Check that the first step is a heating step
	if p.ProgramSteps[0].StepType != StepTypeHeating {
		return errors.New("first step must be a heating step")
	}

	// Check that the last step is a cooling step
	if p.ProgramSteps[len(p.ProgramSteps)-1].StepType != StepTypeCooling {
		return errors.New("last step must be a cooling step")
	}

	// Check temperature progression
	for i := 1; i < len(p.ProgramSteps); i++ {
		currentStep := p.ProgramSteps[i]
		previousStep := p.ProgramSteps[i-1]

		switch currentStep.StepType {
		case StepTypeHeating:
			if currentStep.TargetTemperature <= previousStep.TargetTemperature {
				return errors.New("heating step temperature must be higher than previous step")
			}
		case StepTypeAcclimate:
			if currentStep.TargetTemperature < previousStep.TargetTemperature {
				return errors.New("acclimate step temperature must be greater than or equal to previous step")
			}
		case StepTypeCooling:
			if currentStep.TargetTemperature >= previousStep.TargetTemperature {
				return errors.New("cooling step temperature must be lower than previous step")
			}
		}
	}

	return nil
}

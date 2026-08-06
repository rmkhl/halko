package types

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

const (
	StepTypeHeating   StepType = "heating"
	StepTypeCooling   StepType = "cooling"
	StepTypeAcclimate StepType = "acclimate"

	PowerSettingTypeSimple PowerSettingType = "simple"
	PowerSettingTypeDelta  PowerSettingType = "delta"
	PowerSettingTypePid    PowerSettingType = "pid"

	// steamCeilingCelsius is the temperature steam cannot heat the kiln past.
	// Above it steam is thermally neutral, since it must itself be raised to
	// kiln temperature. Below it steam outruns the heater and is the most
	// immediate contributor to kiln temperature.
	steamCeilingCelsius = 100
)

type (
	StepType         string
	PowerSettingType string

	StepDuration struct {
		time.Duration
	}

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
		Runtime           *StepDuration     `json:"runtime,omitempty"`
		Heater            *PowerPidSettings `json:"heater,omitempty"`
		Fan               *PowerPidSettings `json:"fan,omitempty"`
		Steam             *PowerPidSettings `json:"steam,omitempty"`
	}

	Program struct {
		ProgramName     string        `json:"name"`
		ProgramSteps    []ProgramStep `json:"steps"`
		DefaultsApplied bool          `json:"-"`
	}
)

// Duplicate creates a deep copy of a Program using JSON marshaling/unmarshaling
func (p Program) Duplicate() (Program, error) {
	data, err := json.Marshal(p)
	if err != nil {
		return Program{}, err
	}

	var duplicate Program
	err = json.Unmarshal(data, &duplicate)
	if err != nil {
		return Program{}, err
	}

	return duplicate, nil
}

// MarshalJSON implements json.Marshaler for StepDuration
func (d StepDuration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

// UnmarshalJSON implements json.Unmarshaler for StepDuration
func (d *StepDuration) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	duration, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	d.Duration = duration
	return nil
}

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

	// Validate steam - this resolves Type, which the step validators restrict
	if steamErr := p.Steam.Validate("steam"); steamErr != nil {
		return steamErr
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

// steamIsOff reports whether the step holds steam at a constant zero.
func (p *ProgramStep) steamIsOff() bool {
	return p.Steam.Type == PowerSettingTypeSimple && p.Steam.Power != nil && *p.Steam.Power == 0
}

// steamRunsOpenLoop reports whether the step holds steam at a constant non-zero
// power, with nothing modulating it against the kiln/material delta.
func (p *ProgramStep) steamRunsOpenLoop() bool {
	return p.Steam.Type == PowerSettingTypeSimple && p.Steam.Power != nil && *p.Steam.Power > 0
}

// kilnExitTemperature is the lowest kiln temperature the step can hand over at.
// A heating step's delta controller holds the kiln within
// [material+min_delta, material+max_delta] and the step ends once the material
// reaches target, so the kiln floor at handover is target+min_delta - the top
// of the band is a maximum, not something the next step can rely on. An
// acclimate settles the kiln back onto the material, so it hands over at
// target. Cooling always runs steam off, so its exit temperature gates nothing.
func (p *ProgramStep) kilnExitTemperature() float32 {
	if p.StepType == StepTypeHeating && p.Heater.Type == PowerSettingTypeDelta && p.Heater.MinDelta != nil {
		return float32(p.TargetTemperature) + *p.Heater.MinDelta
	}
	return float32(p.TargetTemperature)
}

func (p *ProgramStep) validateHeatingStep() error {
	if p.Runtime != nil {
		return errors.New("heating step cannot have runtime")
	}
	// Steam may be held constant or modulated against the delta; which of the
	// two is allowed depends on the entry temperature, checked per program.
	if p.Steam.Type != PowerSettingTypeSimple && p.Steam.Type != PowerSettingTypeDelta {
		return errors.New("heating step steam must use simple or delta power control")
	}
	if err := p.Heater.Validate("heater"); err != nil {
		return err
	}
	// Only delta control bounds the kiln/material delta. Simple runs the heater
	// open-loop, and PID regulates the kiln to target while ignoring the
	// material entirely, so both let the kiln run arbitrarily far ahead.
	if p.Heater.Type != PowerSettingTypeDelta {
		return errors.New("heating step heater must use delta power control")
	}
	return nil
}

func (p *ProgramStep) validateAcclimateStep() error {
	if p.Runtime == nil {
		return errors.New("acclimate step must have runtime")
	}
	if p.Steam.Type != PowerSettingTypeSimple {
		return errors.New("acclimate step steam must use simple power control")
	}
	// An acclimate settles the kiln onto the material, so a target below the
	// ceiling leaves steam free to heat the kiln with nothing modulating it.
	if p.TargetTemperature < steamCeilingCelsius && !p.steamIsOff() {
		return fmt.Errorf("acclimate step below %d°C must switch steam off", steamCeilingCelsius)
	}
	return p.Heater.Validate("heater")
}

func (p *ProgramStep) validateCoolingStep() error {
	// Runtime is optional for cooling steps - if specified, step progresses when
	// either target temperature is reached OR runtime expires (whichever comes first)
	// Validate heater - call validate first to capture any errors
	heaterErr := p.Heater.Validate("heater")
	if p.Heater.Type != PowerSettingTypeSimple {
		return errors.New("cooling step heater must use simple power control")
	}
	if heaterErr != nil {
		return heaterErr
	}
	// The kiln descends back through the ceiling during a cooling step, so any
	// steam at all would start heating it again on the way down.
	if !p.steamIsOff() {
		return errors.New("cooling step must switch steam off")
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

		if step.Steam == nil {
			step.Steam = &PowerPidSettings{}
		}
		// Only fill in a constant power when the step named no control method
		// at all; a step asking for delta steam must keep Power unset, or it
		// ends up defining two methods and fails validation.
		if step.Steam.Pid == nil && step.Steam.Power == nil &&
			step.Steam.MinDelta == nil && step.Steam.MaxDelta == nil {
			step.Steam.Power = &zeroPower
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

	if err := p.validateSteamAgainstKilnTemperature(); err != nil {
		return err
	}

	return p.validateStepOrderAndTemperatureProgression()
}

// validateSteamAgainstKilnTemperature enforces where steam may run open-loop.
// A heating step may hold steam at constant power only when the kiln is already
// at or above the ceiling as the step begins, since below it steam outruns the
// heater and opens a delta the heater cannot close by backing off. Each step
// enters where its predecessor handed over; the first starts cold, because
// nothing constrains what temperature a charge is loaded at.
func (p *Program) validateSteamAgainstKilnTemperature() error {
	entry := float32(0)

	for i := range p.ProgramSteps {
		step := &p.ProgramSteps[i]

		if step.StepType == StepTypeHeating && step.steamRunsOpenLoop() && entry < steamCeilingCelsius {
			return fmt.Errorf(
				"step %q enters with the kiln at %.1f°C, below %d°C, so it must switch steam off or use delta power control",
				step.Name, entry, steamCeilingCelsius)
		}

		entry = step.kilnExitTemperature()
	}

	return nil
}

func (p *Program) validateStepOrderAndTemperatureProgression() error {
	if len(p.ProgramSteps) <= 2 {
		return errors.New("program must have at least three steps")
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
			if currentStep.StepType == StepTypeHeating && nextStep.TargetTemperature != currentStep.TargetTemperature {
				return errors.New("acclimate step temperature must match the preceding heating step temperature")
			}
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

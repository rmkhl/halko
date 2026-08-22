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

	// Startup step types. The control unit synthesizes these in front of a
	// program's own steps; a program may not name them.
	StepTypeEqualize     StepType = "equalize"
	StepTypeSteamPrewarm StepType = "steam_prewarm"

	PowerSettingTypeSimple PowerSettingType = "simple"
	PowerSettingTypeDelta  PowerSettingType = "delta"
)

type (
	StepType         string
	PowerSettingType string

	StepDuration struct {
		time.Duration
	}

	PowerPidSettings struct {
		Type     PowerSettingType `json:"type,omitempty"`
		MinDelta *float32         `json:"min_delta,omitempty"`
		MaxDelta *float32         `json:"max_delta,omitempty"`
		Power    *uint8           `json:"power,omitempty"`
	}

	// EqualizeSettings configures the startup steps the control unit runs
	// before a program's own first step. Delta is the band within which the
	// kiln and the material count as equal; SteamPrewarm asks for the steam
	// generator to be brought up to boiling and proved before the run starts.
	EqualizeSettings struct {
		Delta        *float32 `json:"delta,omitempty"`
		SteamPrewarm *bool    `json:"steam_prewarm,omitempty"`
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
		ProgramName string `json:"name"`
		// Description is free text the operator writes about the program -
		// what charge it suits, why the steps are shaped the way they are. It
		// is carried along with the program into a run and its history so the
		// note is still there when the run is looked at later.
		Description     string            `json:"description,omitempty"`
		Equalize        *EqualizeSettings `json:"equalize,omitempty"`
		ProgramSteps    []ProgramStep     `json:"steps"`
		DefaultsApplied bool              `json:"-"`
		// Captured from the defaults so Validate does not need them passed in.
		maxTargetTemperature uint8
		steamCeiling         uint8
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
	if controlMethods != 1 {
		return errors.New(component + " must define exactly one control method: power or min/max deltas")
	}

	if hasDeltas && (p.MinDelta == nil || p.MaxDelta == nil) {
		return errors.New(component + " min and max delta must both be defined")
	}

	// A reversed or collapsed band leaves the controller with a bound it can
	// never cross, which silently disables it rather than failing loudly.
	if hasDeltas && *p.MinDelta >= *p.MaxDelta {
		return errors.New(component + " min delta must be below max delta")
	}

	if p.Power != nil && !isValidPercentage(*p.Power) {
		return errors.New(component + " power must be between 0 and 100")
	}

	return nil
}

// Validate checks a step in isolation. steamCeiling is the temperature above
// which steam stops being able to heat the kiln, which the step needs to know
// to decide whether it may hold steam at constant power.
func (p *ProgramStep) Validate(steamCeiling uint8) error {
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
		return p.validateAcclimateStep(steamCeiling)
	case StepTypeCooling:
		return p.validateCoolingStep()
	case StepTypeEqualize, StepTypeSteamPrewarm:
		return errors.New("equalize and steam_prewarm steps are created by the control unit and cannot be part of a program")
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
// Both delta-controlled step types bottom out at target+min_delta, for
// different reasons: a heating step holds the kiln within
// [material+min_delta, material+max_delta] and ends once the material reaches
// target, while an acclimate holds it within [target+min_delta,
// target+max_delta] throughout. Either way the floor of the band is what the
// next step can rely on - the top is a maximum, not a guarantee. Cooling always
// runs steam off, so its exit temperature gates nothing.
func (p *ProgramStep) kilnExitTemperature() float32 {
	usesDeltaBand := p.StepType == StepTypeHeating || p.StepType == StepTypeAcclimate
	if usesDeltaBand && p.Heater.Type == PowerSettingTypeDelta && p.Heater.MinDelta != nil {
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
	// Only delta control bounds the kiln/material delta; simple runs the heater
	// open-loop and lets the kiln run arbitrarily far ahead.
	if p.Heater.Type != PowerSettingTypeDelta {
		return errors.New("heating step heater must use delta power control")
	}
	return nil
}

func (p *ProgramStep) validateAcclimateStep(steamCeiling uint8) error {
	if p.Runtime == nil {
		return errors.New("acclimate step must have runtime")
	}
	if p.Steam.Type != PowerSettingTypeSimple {
		return errors.New("acclimate step steam must use simple power control")
	}
	// An acclimate holds the kiln at its target, so a target below the ceiling
	// leaves steam free to heat the kiln with nothing modulating it.
	if p.TargetTemperature < steamCeiling && !p.steamIsOff() {
		return fmt.Errorf("acclimate step below %d°C must switch steam off", steamCeiling)
	}
	if err := p.Heater.Validate("heater"); err != nil {
		return err
	}
	if p.Heater.Type != PowerSettingTypeDelta {
		return errors.New("acclimate step heater must use delta power control")
	}
	// Unlike every other step type, an acclimate's deltas are offsets from the
	// step target rather than from the material: the controller holds the kiln
	// in [target+min_delta, target+max_delta] and reads the material only to
	// decide when a pulse may stop. The band must therefore straddle the
	// target. A zero min_delta puts the floor on the stop point and chatters; a
	// zero max_delta leaves the heater unable to drive the air above target,
	// which is the only way it can pull the wood back up.
	if *p.Heater.MinDelta >= 0 {
		return errors.New("acclimate step heater min delta must be negative, the kiln floor below target")
	}
	if *p.Heater.MaxDelta <= 0 {
		return errors.New("acclimate step heater max delta must be positive, the kiln ceiling above target")
	}
	return nil
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
	for i := range p.ProgramSteps {
		step := &p.ProgramSteps[i]

		if step.Heater == nil {
			step.Heater = &PowerPidSettings{}
		}

		if step.Heater.Power == nil &&
			step.Heater.MinDelta == nil && step.Heater.MaxDelta == nil {
			switch step.StepType {
			case StepTypeAcclimate, StepTypeHeating:
				band := defaults.Deltas[step.StepType]
				step.Heater.MinDelta = &band.MinDelta
				step.Heater.MaxDelta = &band.MaxDelta
			case StepTypeCooling:
				// A cooling step never drives the heater; it waits for the
				// charge to fall, so there is nothing to configure.
				off := uint8(0)
				step.Heater.Power = &off
			}
		}

		if step.Fan == nil {
			step.Fan = &PowerPidSettings{}
		}
		if step.Fan.Power == nil {
			step.Fan.Power = defaults.FanPower
		}

		if step.Steam == nil {
			step.Steam = &PowerPidSettings{}
		}
		// Only fill in a constant power when the step named no control method
		// at all; a step asking for delta steam must keep Power unset, or it
		// ends up defining two methods and fails validation.
		if step.Steam.Power == nil &&
			step.Steam.MinDelta == nil && step.Steam.MaxDelta == nil {
			step.Steam.Power = defaults.SteamPower
		}
	}

	if p.Equalize == nil {
		p.Equalize = &EqualizeSettings{}
	}
	if p.Equalize.Delta == nil {
		p.Equalize.Delta = defaults.Equalize.Delta
	}
	if p.Equalize.SteamPrewarm == nil {
		p.Equalize.SteamPrewarm = defaults.Equalize.SteamPrewarm
	}

	p.maxTargetTemperature = *defaults.MaxTargetTemperature
	p.steamCeiling = *defaults.SteamCeiling
	p.DefaultsApplied = true
}

// PrependStartupSteps puts the control unit's own startup steps in front of the
// program's. They are synthesized rather than authored so that everything
// downstream - the status endpoint, the execution log, the history record and
// the webapp - sees ordinary steps instead of special-casing a hidden state.
//
// Call this only after Validate has passed. The authored program is what the
// rules apply to; a synthesized equalize step in front of it would defeat
// "first step must be a heating step".
//
// The steps carry no power settings: the FSM states that run them drive the
// channels from the kiln/material delta, not from a PowerController.
func (p *Program) PrependStartupSteps() {
	startup := []ProgramStep{{Name: "Equalize", StepType: StepTypeEqualize}}
	if *p.Equalize.SteamPrewarm {
		startup = append(startup, ProgramStep{Name: "Steam warm-up", StepType: StepTypeSteamPrewarm})
	}
	p.ProgramSteps = append(startup, p.ProgramSteps...)
}

func (p *Program) Validate() error {
	if !p.DefaultsApplied {
		return errors.New("defaults must be applied before validation")
	}

	if *p.Equalize.Delta <= 0 {
		return errors.New("equalize delta must be greater than zero")
	}

	for _, step := range p.ProgramSteps {
		err := step.Validate(p.steamCeiling)
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

		if step.StepType == StepTypeHeating && step.steamRunsOpenLoop() && entry < float32(p.steamCeiling) {
			return fmt.Errorf(
				"step %q enters with the kiln at %.1f°C, below %d°C, so it must switch steam off or use delta power control",
				step.Name, entry, p.steamCeiling)
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

		if currentStep.TargetTemperature > p.maxTargetTemperature {
			return fmt.Errorf("target temperature must not exceed %d degrees", p.maxTargetTemperature)
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

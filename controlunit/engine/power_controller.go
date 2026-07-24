// Implements simple, delta and PID based power controller.
package engine

import (
	"time"

	"github.com/rmkhl/halko/types"
)

type (
	// PidControllerState holds mutable state for a PidController.
	PidControllerState struct {
		CurrentError           float32
		CurrentErrorIntegral   float32
		CurrentErrorDerivative float32
		PreviousUpdate         int64
	}

	PidController struct {
		Config *types.PidSettings
		State  PidControllerState
	}

	PowerController struct {
		PidController     *PidController
		TargetTemperature float32
		Settings          *types.PowerPidSettings
	}
)

func NewPidController(config *types.PidSettings) *PidController {
	return &PidController{
		Config: config,
	}
}

// Update the controller state.
func (c *PidController) Update(reference float32, actual float32) float32 {
	if c.State.PreviousUpdate == 0 {
		c.State.PreviousUpdate = time.Now().Unix()
		return 0
	}
	previousError := c.State.CurrentError
	sampleInterval := time.Now().Unix() - c.State.PreviousUpdate
	c.State.CurrentError = reference - actual
	c.State.CurrentErrorDerivative = (c.State.CurrentError - previousError) / float32(sampleInterval)
	c.State.CurrentErrorIntegral += c.State.CurrentError * float32(sampleInterval)
	return c.Config.Kp*c.State.CurrentError +
		c.Config.Ki*c.State.CurrentErrorIntegral +
		c.Config.Kd*c.State.CurrentErrorDerivative
}

// New power controller with the given configuration and settings.
func NewPowerController(targetTemperature float32, settings *types.PowerPidSettings) *PowerController {
	controller := &PowerController{
		TargetTemperature: targetTemperature,
		Settings:          settings,
	}

	if settings.Type == types.PowerSettingTypePid {
		controller.PidController = NewPidController(settings.Pid)
	}

	return controller
}

// Update the power controller with the current temperature and power.
// Returns the new power percentage to use.
func (c *PowerController) Update(power uint8, owenTemperature float32, woodTemperature float32) uint8 {
	switch c.Settings.Type {
	case types.PowerSettingTypeSimple:
		return *c.Settings.Power

	case types.PowerSettingTypeDelta:
		targetTemperature := c.TargetTemperature

		maxOvenTemp := woodTemperature + *c.Settings.MaxDelta
		targetTemperature = min(targetTemperature, maxOvenTemp)

		minOvenTemp := woodTemperature + *c.Settings.MinDelta
		targetTemperature = max(targetTemperature, minOvenTemp)

		if owenTemperature < targetTemperature {
			return 100
		}
		return 0

	case types.PowerSettingTypePid:
		powerDelta := c.PidController.Update(c.TargetTemperature, owenTemperature)
		return uint8(min(100, max(int(float32(power)+powerDelta), 0)))

	default:
		return 0 // Safeguard, this should not happen, but lets turn everything off
	}
}

// heatingDeltaController keeps the kiln inside the band
// [material+minDelta, material+maxDelta] with hysteresis: the heater turns
// off at the upper bound, back on at the lower bound, and holds its previous
// state inside the band.
type heatingDeltaController struct {
	minDelta float32
	maxDelta float32
	heaterOn bool
}

func (c *heatingDeltaController) Update(kilnTemperature, materialTemperature float32) uint8 {
	switch {
	case kilnTemperature >= materialTemperature+c.maxDelta:
		c.heaterOn = false
	case kilnTemperature <= materialTemperature+c.minDelta:
		c.heaterOn = true
	}
	if c.heaterOn {
		return 100
	}
	return 0
}

// acclimateDeltaController holds the material at the step target while
// keeping the kiln inside the delta safety envelope relative to the material.
// Heating requires both demand (material below target, no hysteresis) and an
// armed envelope (same hysteresis band as heating: disarm at
// material+maxDelta, re-arm at material+minDelta).
type acclimateDeltaController struct {
	target     float32
	minDelta   float32
	maxDelta   float32
	envelopeOK bool
}

func (c *acclimateDeltaController) Update(kilnTemperature, materialTemperature float32) uint8 {
	switch {
	case kilnTemperature >= materialTemperature+c.maxDelta:
		c.envelopeOK = false
	case kilnTemperature <= materialTemperature+c.minDelta:
		c.envelopeOK = true
	}
	if c.envelopeOK && materialTemperature < c.target {
		return 100
	}
	return 0
}

// simplePowerController always returns its configured power.
type simplePowerController struct {
	power uint8
}

func (c *simplePowerController) Update(_, _ float32) uint8 {
	return c.power
}

// pidPowerController adjusts its own previous output by the PID delta,
// clamped to 0-100. PID error is computed on kiln temperature vs target.
type pidPowerController struct {
	pid       *PidController
	target    float32
	lastPower uint8
}

func (c *pidPowerController) Update(kilnTemperature, _ float32) uint8 {
	powerDelta := c.pid.Update(c.target, kilnTemperature)
	c.lastPower = uint8(min(100, max(int(float32(c.lastPower)+powerDelta), 0)))
	return c.lastPower
}

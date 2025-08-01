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

// Implements simple PID based power controller.
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
		ConstantPower     uint8
		TargetTemperature float32
		MaxDelta          float32
		MinDelta          float32
	}
)

func NewPidController(config *types.PidSettings) *PidController {
	return &PidController{
		Config: config,
	}
}

// Update the controller state.
func (c *PidController) Update(reference float32, actual float32) float32 {
	// first call, lets just keep the status quo
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

func (c *PidController) Reset() {
	c.State = PidControllerState{}
}

// Simple constant power controller, basically used to control the fan and the humidifier.
func newConstantPowerController(power uint8) *PowerController {
	return &PowerController{
		ConstantPower: power,
	}
}

// New power controller with the given configuration and settings. If the Power is set to non zero value, the PID settings are ignored.
func NewPowerController(targetTemperature float32, settings *types.PowerPidSettings, defaultPidSettings *types.PidSettings, maxDeltaHeating float32, minDeltaHeating float32) *PowerController {
	// Use default pid values is none are defined in the settings
	var controllerConfig *types.PidSettings
	var controller *PidController
	var maxDelta float32
	var minDelta float32

	if settings.Pid == nil {
		controllerConfig = defaultPidSettings
	} else {
		controllerConfig = settings.Pid
	}

	if controllerConfig != nil {
		controller = NewPidController(controllerConfig)
	}

	// Use the delta values from the step settings, or fall back to heating config values
	if settings.MaxDelta != 0 {
		maxDelta = settings.MaxDelta
		if settings.MinDelta != 0 {
			minDelta = settings.MinDelta
		} else {
			minDelta = 0 // Default for non-heating steps
		}
	} else {
		maxDelta = maxDeltaHeating
		minDelta = minDeltaHeating
	}

	return &PowerController{
		PidController:     controller,
		ConstantPower:     settings.Power,
		TargetTemperature: targetTemperature,
		MaxDelta:          maxDelta,
		MinDelta:          minDelta,
	}
}

// Update the power controller with the current temperature and power.
// Returns the new power percentage to use.
func (c *PowerController) Update(power uint8, owenTemperature float32, woodTemperature float32) uint8 {
	// If the power is set to a constant value, return that value
	if c.PidController == nil {
		return c.ConstantPower
	}

	// Calculate target temperature based on wood temperature and max/min delta
	var targetTemperature float32

	if c.MinDelta != 0 {
		// When both min and max delta are specified, respect both constraints
		// Don't let the oven get too far ahead or behind the wood temperature
		targetTemperature = min(c.TargetTemperature,
			max(woodTemperature+c.MinDelta,
				min(woodTemperature+c.MaxDelta, c.TargetTemperature)))
	} else {
		// For other step types, just use the max delta as before
		targetTemperature = min(c.TargetTemperature, woodTemperature+c.MaxDelta)
	}

	powerDelta := c.PidController.Update(targetTemperature, owenTemperature)
	// Limit the power to be between 0 and 100
	return uint8(min(100, max(int(float32(power)+powerDelta), 0)))
}

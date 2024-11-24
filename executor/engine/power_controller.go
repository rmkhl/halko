// Implements simple PID based power controller.
package engine

import (
	"time"

	"github.com/rmkhl/halko/types"
)

type (
	// PidControllerState holds mutable state for a PidController.
	PidControllerState struct {
		CurrentError           float64
		CurrentErrorIntegral   float64
		CurrentErrorDerivative float64
		PreviousUpdate         int64
	}

	PidController struct {
		Config *types.PidSettings
		State  PidControllerState
	}

	PowerController struct {
		PidController     *PidController
		ConstantPower     uint8
		TargetTemperature float64
		MaxDelta          float64
	}
)

func NewPidController(config *types.PidSettings) *PidController {
	return &PidController{
		Config: config,
	}
}

// Update the controller state.
func (c *PidController) Update(reference float64, actual float64) float64 {
	// first call, lets just keep the status quo
	if c.State.PreviousUpdate == 0 {
		c.State.PreviousUpdate = time.Now().Unix()
		return 0
	}
	previousError := c.State.CurrentError
	sampleInterval := time.Now().Unix() - c.State.PreviousUpdate
	c.State.CurrentError = reference - actual
	c.State.CurrentErrorDerivative = (c.State.CurrentError - previousError) / float64(sampleInterval)
	c.State.CurrentErrorIntegral += c.State.CurrentError * float64(sampleInterval)
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
func NewPowerController(targetTemperature float64, settings *types.PowerPidSettings, defaultPidSettings *types.PidSettings) *PowerController {
	// Use default pid values is none are defined in the settings
	var controllerConfig *types.PidSettings
	var controller *PidController

	if settings.Pid == nil {
		controllerConfig = defaultPidSettings
	} else {
		controllerConfig = settings.Pid
	}
	if controllerConfig != nil {
		controller = NewPidController(controllerConfig)
	}
	return &PowerController{
		PidController:     controller,
		ConstantPower:     settings.Power,
		TargetTemperature: targetTemperature,
		MaxDelta:          float64(settings.MaxDelta),
	}
}

// Update the power controller with the current temperature and power.
// Returns the new power percentage to use.
func (c *PowerController) Update(power uint8, owenTemperature float32, woodTemperature float32) uint8 {
	// If the power is set to a constant value, return that value
	if c.PidController == nil {
		return c.ConstantPower
	}
	// Adjust the target temperature for the owen based on the wood temperature
	targetTemperature := min(c.TargetTemperature, float64(woodTemperature)+c.MaxDelta)
	powerDelta := c.PidController.Update(targetTemperature, float64(owenTemperature))
	// Limit the power to be between 0 and 100
	return uint8(min(100, max(int(float64(power)+powerDelta), 0)))
}

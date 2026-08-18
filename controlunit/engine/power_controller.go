// Implements simple and per-step delta based power controllers.
package engine

import (
	"github.com/rmkhl/halko/types"
)

type (
	// PowerController decides the power percentage from the latest temperature
	// readings. Implementations own whatever state they need; the returned value
	// is re-commanded to the power unit on every reading.
	PowerController interface {
		Update(kilnTemperature, materialTemperature float32) uint8
	}
)

func NewPowerController(stepType types.StepType, targetTemperature float32, settings *types.PowerPidSettings) PowerController {
	failSafe := &simplePowerController{power: 0}
	if settings == nil {
		return failSafe
	}

	switch settings.Type {
	case types.PowerSettingTypeSimple:
		if settings.Power == nil {
			return failSafe
		}
		return &simplePowerController{power: *settings.Power}

	case types.PowerSettingTypeDelta:
		if settings.MinDelta == nil || settings.MaxDelta == nil {
			return failSafe
		}
		switch stepType {
		case types.StepTypeHeating:
			return &heatingDeltaController{minDelta: *settings.MinDelta, maxDelta: *settings.MaxDelta, heaterOn: true}
		case types.StepTypeAcclimate:
			return &acclimateDeltaController{target: targetTemperature, minDelta: *settings.MinDelta, maxDelta: *settings.MaxDelta}
		default:
			return failSafe
		}

	default:
		return failSafe
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

// acclimateDeltaController does two different jobs, and takes the reference
// and the half of the delta band that belongs to whichever one is in effect.
//
// Heating the kiln closes a distance to the step target, so it bands the kiln
// in [target+minDelta, target]: 0 is on target and minDelta, which is negative,
// is the most sag tolerated before the heater fires.
//
// Heating the material drives the gradient that pushes heat into the wood, and
// that gradient is measured from the wood, so it bands the kiln in
// [material, material+maxDelta]: 0 is no transfer and maxDelta, which is
// positive, is the most gradient the charge can safely take.
//
// The two bands meet at the target and together span
// [target+minDelta, target+maxDelta] without a gap, so the moment the wood
// reaches target a kiln still up in the heating band is above its new one and
// the heater stops.
type acclimateDeltaController struct {
	target   float32
	minDelta float32
	maxDelta float32
	heaterOn bool
}

func (c *acclimateDeltaController) Update(kilnTemperature, materialTemperature float32) uint8 {
	lower, upper := c.target+c.minDelta, c.target
	if materialTemperature < c.target {
		lower, upper = materialTemperature, materialTemperature+c.maxDelta
	}
	switch {
	case kilnTemperature >= upper:
		c.heaterOn = false
	case kilnTemperature <= lower:
		c.heaterOn = true
	}
	if c.heaterOn {
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

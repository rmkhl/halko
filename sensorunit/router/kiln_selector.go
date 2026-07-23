package router

import "github.com/rmkhl/halko/types"

// kilnSelectionHysteresis is how much hotter (in °C) the non-selected kiln
// sensor must read before the selection switches to it. Without this margin
// two sensors idling within noise of each other make the reported kiln
// temperature flip between them on every poll.
const kilnSelectionHysteresis = 0.5

// kilnSelector picks which of the two kiln sensors the reported temperature
// comes from. The zero value starts on the primary sensor.
type kilnSelector struct {
	secondarySelected bool
}

// Select returns the kiln temperature to report given the two sensor
// readings, either of which may be types.InvalidTemperatureReading. It keeps
// reporting the currently selected sensor unless the other one is valid and
// exceeds it by more than the hysteresis margin, or the selected sensor's
// reading is invalid.
func (s *kilnSelector) Select(primary, secondary float32) float32 {
	primaryValid := primary != types.InvalidTemperatureReading
	secondaryValid := secondary != types.InvalidTemperatureReading

	switch {
	case !primaryValid && !secondaryValid:
		return types.InvalidTemperatureReading
	case !secondaryValid:
		s.secondarySelected = false
	case !primaryValid:
		s.secondarySelected = true
	case s.secondarySelected && primary > secondary+kilnSelectionHysteresis:
		s.secondarySelected = false
	case !s.secondarySelected && secondary > primary+kilnSelectionHysteresis:
		s.secondarySelected = true
	}

	if s.secondarySelected {
		return secondary
	}
	return primary
}

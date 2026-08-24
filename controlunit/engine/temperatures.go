package engine

import (
	"github.com/rmkhl/halko/types"
)

type (
	// fsmTemperatures holds the most recent valid reading for each sensor.
	// A failed probe reports types.InvalidTemperatureReading; keeping the
	// previous value stops that sentinel reaching the power controllers,
	// and the per-sensor timestamps say how long that has been going on.
	fsmTemperatures struct {
		updated         int64
		reading         temperatureReadings
		kilnValidAt     int64
		materialValidAt int64
	}
)

// The sensor names invalidFor reports, and the failsafe logs and fails a run
// with. They name a sensor to a human, so they live in one place.
const (
	sensorNameKiln     = "kiln"
	sensorNameMaterial = "material"
)

func validReading(value float32) bool {
	return value != types.InvalidTemperatureReading
}

// resolveKiln turns the kiln's two sensor readings into the single temperature
// the controllers act on, by the strategy the installation configured. A
// sensor that failed is left out entirely: with one reading left there is
// nothing to compare or average, so the survivor is the answer whatever the
// strategy says.
func resolveKiln(strategy types.KilnSensorStrategy, primary, secondary float32) float32 {
	primaryValid := validReading(primary)
	secondaryValid := validReading(secondary)

	switch {
	case !primaryValid && !secondaryValid:
		return types.InvalidTemperatureReading
	case !secondaryValid:
		return primary
	case !primaryValid:
		return secondary
	}

	switch strategy {
	case types.KilnSensorLower:
		if secondary < primary {
			return secondary
		}
		return primary
	case types.KilnSensorAverage:
		return (primary + secondary) / 2
	case types.KilnSensorHigher:
		fallthrough
	default:
		if secondary > primary {
			return secondary
		}
		return primary
	}
}

// observe records a sample, keeping the previous value for any sensor that
// reported an invalid reading. The cold junction readings are the exception:
// they are logged, never controlled or failsafed on.
func (t *fsmTemperatures) observe(sample temperatureReadings, now int64, strategy types.KilnSensorStrategy) {
	// Resolve the sample that just arrived, never the held values: a sensor
	// that dropped out ten minutes ago must not be averaged against a live one.
	// The hold below then applies to the resolved value, so a dropout narrows
	// the resolution to the survivor on the very next poll.
	if kiln := resolveKiln(strategy, sample.KilnPrimary, sample.KilnSecondary); validReading(kiln) {
		t.reading.Kiln = kiln
		t.kilnValidAt = now
	}
	if validReading(sample.Material) {
		t.reading.Material = sample.Material
		t.materialValidAt = now
	}
	// The raw pair and the cold junctions are copied straight through rather
	// than held: a held raw reading would hide the dropout it exists to show,
	// a stale cold junction would mask the drift it is here to expose, and
	// nothing controls on any of them.
	t.reading.KilnPrimary = sample.KilnPrimary
	t.reading.KilnSecondary = sample.KilnSecondary
	t.reading.MaterialDie = sample.MaterialDie
	t.reading.KilnPrimaryDie = sample.KilnPrimaryDie
	t.reading.KilnSecondaryDie = sample.KilnSecondaryDie
}

// invalidFor names the sensor that has gone longest without a valid reading
// and returns how many seconds it has been. A sensor that has never reported
// a valid reading is measured from programStart, so a run that never gets one
// still trips the failsafe.
func (t *fsmTemperatures) invalidFor(now, programStart int64) (string, int64) {
	kilnSince := t.kilnValidAt
	if kilnSince < programStart {
		kilnSince = programStart
	}
	materialSince := t.materialValidAt
	if materialSince < programStart {
		materialSince = programStart
	}

	if kilnSince <= materialSince {
		return sensorNameKiln, now - kilnSince
	}
	return sensorNameMaterial, now - materialSince
}

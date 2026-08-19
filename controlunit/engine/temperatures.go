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

func validReading(value float32) bool {
	return value != types.InvalidTemperatureReading
}

// observe records a sample, keeping the previous value for any sensor that
// reported an invalid reading. The cold junction readings are the exception:
// they are logged, never controlled or failsafed on.
func (t *fsmTemperatures) observe(sample temperatureReadings, now int64) {
	if validReading(sample.Kiln) {
		t.reading.Kiln = sample.Kiln
		t.kilnValidAt = now
	}
	if validReading(sample.Material) {
		t.reading.Material = sample.Material
		t.materialValidAt = now
	}
	// Copied straight through rather than held: a stale cold junction value
	// would mask the very drift the reading is here to expose, and nothing
	// controls on it.
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
		return "kiln", now - kilnSince
	}
	return "material", now - materialSince
}

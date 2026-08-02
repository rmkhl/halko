package engine

import (
	"testing"

	"github.com/rmkhl/halko/types"
)

const invalid = float32(types.InvalidTemperatureReading)

func TestObserveHoldsLastValidReading(t *testing.T) {
	var temps fsmTemperatures

	temps.observe(temperatureReadings{Kiln: 100, Material: 50}, 1000)
	if temps.reading.Kiln != 100 || temps.reading.Material != 50 {
		t.Fatalf("after valid sample reading = %+v, want {50 100}", temps.reading)
	}

	// Both sensors fail: the previous values must survive.
	temps.observe(temperatureReadings{Kiln: invalid, Material: invalid}, 1010)
	if temps.reading.Kiln != 100 || temps.reading.Material != 50 {
		t.Errorf("after invalid sample reading = %+v, want values held at {50 100}", temps.reading)
	}
	if temps.kilnValidAt != 1000 || temps.materialValidAt != 1000 {
		t.Errorf("validAt = (%d, %d), want (1000, 1000) — invalid samples must not stamp",
			temps.kilnValidAt, temps.materialValidAt)
	}
}

func TestObserveTracksSensorsIndependently(t *testing.T) {
	var temps fsmTemperatures
	temps.observe(temperatureReadings{Kiln: 100, Material: 50}, 1000)

	// Only the kiln fails; material must keep updating.
	temps.observe(temperatureReadings{Kiln: invalid, Material: 60}, 1010)

	if temps.reading.Kiln != 100 {
		t.Errorf("Kiln = %v, want 100 held", temps.reading.Kiln)
	}
	if temps.reading.Material != 60 {
		t.Errorf("Material = %v, want 60", temps.reading.Material)
	}
	if temps.kilnValidAt != 1000 {
		t.Errorf("kilnValidAt = %d, want 1000", temps.kilnValidAt)
	}
	if temps.materialValidAt != 1010 {
		t.Errorf("materialValidAt = %d, want 1010", temps.materialValidAt)
	}
}

func TestInvalidForMeasuresFromProgramStartWhenNeverValid(t *testing.T) {
	var temps fsmTemperatures // no sample ever observed

	sensor, seconds := temps.invalidFor(1130, 1000)
	if seconds != 130 {
		t.Errorf("seconds = %d, want 130 (measured from program start)", seconds)
	}
	if sensor == "" {
		t.Error("sensor = empty, want the offending sensor named")
	}
}

func TestInvalidForResetsOnValidReading(t *testing.T) {
	var temps fsmTemperatures
	temps.observe(temperatureReadings{Kiln: 100, Material: 50}, 1100)

	if _, seconds := temps.invalidFor(1130, 1000); seconds != 30 {
		t.Errorf("seconds = %d, want 30", seconds)
	}
}

func TestInvalidForReportsTheWorseSensor(t *testing.T) {
	var temps fsmTemperatures
	temps.observe(temperatureReadings{Kiln: 100, Material: 50}, 1000)
	// Material keeps reporting, kiln does not.
	temps.observe(temperatureReadings{Kiln: invalid, Material: 60}, 1100)

	sensor, seconds := temps.invalidFor(1130, 1000)
	if sensor != "kiln" {
		t.Errorf("sensor = %q, want \"kiln\"", sensor)
	}
	if seconds != 130 {
		t.Errorf("seconds = %d, want 130", seconds)
	}
}

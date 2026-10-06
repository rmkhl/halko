package router

import (
	"testing"

	"github.com/rmkhl/halko/sensorunit/serial"
	"github.com/rmkhl/halko/types"
)

// Which of the two readings to control on is the control unit's decision now,
// so this service reports both and resolves neither.
func TestTemperatureResponseCarriesBothKilnSensors(t *testing.T) {
	response := temperatureResponseFrom([]serial.Temperature{
		{Name: probeKilnPrimary, Value: 98.2},
		{Name: probeKilnSecondary, Value: 99.1},
		{Name: probeWood, Value: 42.5},
	})

	if response["kiln_primary"] != 98.2 {
		t.Errorf("kiln_primary = %v, want 98.2", response["kiln_primary"])
	}
	if response["kiln_secondary"] != 99.1 {
		t.Errorf("kiln_secondary = %v, want 99.1", response["kiln_secondary"])
	}
	if response["material"] != 42.5 {
		t.Errorf("material = %v, want 42.5", response["material"])
	}
	if _, present := response["kiln"]; present {
		t.Errorf("response still carries a resolved kiln key: %v", response)
	}
}

// A probe the device did not report must arrive as the invalid sentinel rather
// than as a plausible zero: the control unit decides what to do about it.
func TestTemperatureResponseReportsAMissingKilnSensorAsInvalid(t *testing.T) {
	response := temperatureResponseFrom([]serial.Temperature{
		{Name: probeKilnSecondary, Value: 99.1},
		{Name: probeWood, Value: 42.5},
	})

	if response["kiln_primary"] != types.InvalidTemperatureReading {
		t.Errorf("kiln_primary = %v, want the invalid sentinel", response["kiln_primary"])
	}
	if response["kiln_secondary"] != 99.1 {
		t.Errorf("kiln_secondary = %v, want 99.1", response["kiln_secondary"])
	}
}

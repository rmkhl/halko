package router

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rmkhl/halko/sensorunit/serial"
	"github.com/rmkhl/halko/types"
)

// materialDie is repeated across these tests often enough that goconst asks
// for a name.
const materialDie = "material_die"

// The die endpoint answers from what the last device read recorded, so a
// caller gets values without a serial round trip of its own.
func TestDieTemperaturesServeTheLastRecordedReadings(t *testing.T) {
	api := NewAPI(nil)
	api.storeDieReadings(types.TemperatureResponse{
		"kiln_primary_die":   27.5,
		"kiln_secondary_die": 28.125,
		materialDie:          41.875,
	})

	recorder := httptest.NewRecorder()
	api.getDieTemperatures(recorder, httptest.NewRequest(http.MethodGet, "/temperatures/die", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}

	var response types.APIResponse[types.TemperatureResponse]
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("decoding response: %v", err)
	}

	want := map[string]float32{
		"kiln_primary_die":   27.5,
		"kiln_secondary_die": 28.125,
		materialDie:          41.875,
	}
	for name, value := range want {
		if response.Data[name] != value {
			t.Errorf("%s = %v, want %v", name, response.Data[name], value)
		}
	}
}

// Before any read has happened there is nothing to serve, which must be an
// empty answer rather than three readings of 0 °C.
func TestDieTemperaturesAreEmptyBeforeTheFirstRead(t *testing.T) {
	api := NewAPI(nil)

	recorder := httptest.NewRecorder()
	api.getDieTemperatures(recorder, httptest.NewRequest(http.MethodGet, "/temperatures/die", nil))

	var response types.APIResponse[types.TemperatureResponse]
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if len(response.Data) != 0 {
		t.Fatalf("readings = %v, want none", response.Data)
	}
}

// Handing out the live map would let a caller read it while the next device
// read is writing to it.
func TestDieReadingsReturnsACopy(t *testing.T) {
	api := NewAPI(nil)
	api.storeDieReadings(types.TemperatureResponse{materialDie: 41.875})

	readings := api.dieReadings()
	readings[materialDie] = 99.0

	if got := api.dieReadings()[materialDie]; got != 41.875 {
		t.Fatalf("stored reading = %v, want 41.875 (caller mutated the source)", got)
	}
}

// Which of the two readings to control on is the control unit's decision now,
// so this service reports both and resolves neither.
func TestTemperatureResponseCarriesBothKilnSensors(t *testing.T) {
	response, _ := temperatureResponseFrom([]serial.Temperature{
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
	response, _ := temperatureResponseFrom([]serial.Temperature{
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

// The cold junctions ride along on the same read and stay in their own map,
// which the die endpoint serves.
func TestTemperatureResponseSeparatesTheColdJunctions(t *testing.T) {
	response, dies := temperatureResponseFrom([]serial.Temperature{
		{Name: probeKilnPrimary, Value: 98.2},
		{Name: probeKilnPrimaryDie, Value: 27.5},
		{Name: probeWoodDie, Value: 41.8},
	})

	if dies["kiln_primary_die"] != 27.5 || dies[materialDie] != 41.8 {
		t.Errorf("dies = %v, want the cold junctions", dies)
	}
	if _, present := response["kiln_primary_die"]; present {
		t.Errorf("a cold junction leaked into the temperature response: %v", response)
	}
}

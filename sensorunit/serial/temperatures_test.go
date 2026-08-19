package serial

import (
	"testing"

	"github.com/rmkhl/halko/types"
)

// dieSuffix is the cold junction half of a read response, appended to the
// thermocouple half in tests that do not care about its contents.
const dieSuffix = ",KilnPrimaryDie=27.5C,KilnSecondaryDie=28.125C,WoodDie=41.875C"

func TestParseTemperatureResponseValid(t *testing.T) {
	got, err := parseTemperatureResponse(
		"KilnPrimary=20.5C,KilnSecondary=21.0C,Wood=19.25C" + dieSuffix)
	if err != nil {
		t.Fatalf("parseTemperatureResponse() error = %v", err)
	}
	want := []Temperature{
		{Name: "KilnPrimary", Value: 20.5, Unit: "C"},
		{Name: "KilnSecondary", Value: 21.0, Unit: "C"},
		{Name: "Wood", Value: 19.25, Unit: "C"},
		{Name: "KilnPrimaryDie", Value: 27.5, Unit: "C"},
		{Name: "KilnSecondaryDie", Value: 28.125, Unit: "C"},
		{Name: "WoodDie", Value: 41.875, Unit: "C"},
	}
	if len(got) != len(want) {
		t.Fatalf("got %d readings, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("reading %d = %+v, want %+v", i, got[i], want[i])
		}
	}
}

// NaN is a valid report of a failed sensor, not a parse failure.
func TestParseTemperatureResponseNaN(t *testing.T) {
	tests := []struct {
		name     string
		response string
		want     [3]float32
	}{
		{
			name:     "one sensor failed",
			response: "KilnPrimary=20.5C,KilnSecondary=NaN,Wood=19.0C" + dieSuffix,
			want:     [3]float32{20.5, types.InvalidTemperatureReading, 19.0},
		},
		{
			name:     "all sensors failed",
			response: "KilnPrimary=NaN,KilnSecondary=NaN,Wood=NaN" + dieSuffix,
			want: [3]float32{
				types.InvalidTemperatureReading,
				types.InvalidTemperatureReading,
				types.InvalidTemperatureReading,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseTemperatureResponse(tt.response)
			if err != nil {
				t.Fatalf("parseTemperatureResponse() error = %v", err)
			}
			if len(got) != expectedReadings {
				t.Fatalf("got %d readings, want %d", len(got), expectedReadings)
			}
			for i, want := range tt.want {
				if got[i].Value != want {
					t.Errorf("reading %d value = %v, want %v", i, got[i].Value, want)
				}
			}
		})
	}
}

// A thermocouple faults independently of the chip reporting it, so NaN in
// one half of the line says nothing about the other.
func TestParseTemperatureResponseDieFailsIndependently(t *testing.T) {
	got, err := parseTemperatureResponse(
		"KilnPrimary=131.75C,KilnSecondary=NaN,Wood=122.75C," +
			"KilnPrimaryDie=27.5C,KilnSecondaryDie=28.125C,WoodDie=NaN")
	if err != nil {
		t.Fatalf("parseTemperatureResponse() error = %v", err)
	}
	if got[1].Value != types.InvalidTemperatureReading {
		t.Errorf("failed thermocouple = %v, want invalid", got[1].Value)
	}
	if got[4].Value != 28.125 {
		t.Errorf("die reading beside a failed thermocouple = %v, want 28.125", got[4].Value)
	}
	if got[5].Value != types.InvalidTemperatureReading {
		t.Errorf("failed die reading = %v, want invalid", got[5].Value)
	}
}

// A corrupted line is not salvageable: rejecting it lets the caller retry
// rather than inventing a value for the sensor that went missing.
func TestParseTemperatureResponseRejectsCorruption(t *testing.T) {
	tests := []struct {
		name     string
		response string
	}{
		{"unparseable value", "KilnPrimary=20.5C,KilnSecondary=1x.5C,Wood=19.0C" + dieSuffix},
		{"missing equals", "KilnPrimary=20.5C,KilnSecondary,Wood=19.0C" + dieSuffix},
		{"empty value", "KilnPrimary=20.5C,KilnSecondary=,Wood=19.0C" + dieSuffix},
		{"thermocouples only", "KilnPrimary=20.5C,KilnSecondary=21.0C,Wood=19.0C"},
		{"truncated die group", "KilnPrimary=20.5C,KilnSecondary=21.0C,Wood=19.0C,KilnPrimaryDie=27.5C"},
		{"too many fields", "KilnPrimary=20.5C,KilnSecondary=21.0C,Wood=19.0C" + dieSuffix + ",Extra=1.0C"},
		{"empty response", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseTemperatureResponse(tt.response)
			if err == nil {
				t.Fatalf("parseTemperatureResponse(%q) = %+v, want error", tt.response, got)
			}
			if got != nil {
				t.Errorf("readings = %+v, want nil on error", got)
			}
		})
	}
}

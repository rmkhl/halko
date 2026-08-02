package serial

import (
	"testing"

	"github.com/rmkhl/halko/types"
)

func TestParseTemperatureResponseValid(t *testing.T) {
	got, err := parseTemperatureResponse("KilnPrimary=20.5C,KilnSecondary=21.0C,Wood=19.25C")
	if err != nil {
		t.Fatalf("parseTemperatureResponse() error = %v", err)
	}
	want := []Temperature{
		{Name: "KilnPrimary", Value: 20.5, Unit: "C"},
		{Name: "KilnSecondary", Value: 21.0, Unit: "C"},
		{Name: "Wood", Value: 19.25, Unit: "C"},
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
			response: "KilnPrimary=20.5C,KilnSecondary=NaN,Wood=19.0C",
			want:     [3]float32{20.5, types.InvalidTemperatureReading, 19.0},
		},
		{
			name:     "all sensors failed",
			response: "KilnPrimary=NaN,KilnSecondary=NaN,Wood=NaN",
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
			if len(got) != 3 {
				t.Fatalf("got %d readings, want 3", len(got))
			}
			for i, want := range tt.want {
				if got[i].Value != want {
					t.Errorf("reading %d value = %v, want %v", i, got[i].Value, want)
				}
			}
		})
	}
}

// A corrupted line is not salvageable: rejecting it lets the caller retry
// rather than inventing a value for the sensor that went missing.
func TestParseTemperatureResponseRejectsCorruption(t *testing.T) {
	tests := []struct {
		name     string
		response string
	}{
		{"unparseable value", "KilnPrimary=20.5C,KilnSecondary=1x.5C,Wood=19.0C"},
		{"missing equals", "KilnPrimary=20.5C,KilnSecondary,Wood=19.0C"},
		{"empty value", "KilnPrimary=20.5C,KilnSecondary=,Wood=19.0C"},
		{"too few fields", "KilnPrimary=20.5C,Wood=19.0C"},
		{"too many fields", "KilnPrimary=20.5C,KilnSecondary=21.0C,Wood=19.0C,Extra=1.0C"},
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

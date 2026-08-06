package main

import "testing"

func TestFormatTemperatureName(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"kiln", "Kiln"},
		{"material", "Material/Wood"},
		{"humidifier", "Humidifier"},
		{"HEATER", "Heater"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run("name "+tt.name, func(t *testing.T) {
			if got := formatTemperatureName(tt.name); got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

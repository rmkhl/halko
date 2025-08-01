package tests

import (
	"testing"

	"github.com/rmkhl/halko/types"
)

func TestConfigReading(t *testing.T) {
	// Test reading the template configuration file
	config, err := types.ReadHalkoConfig("../templates/halko.cfg")
	if err != nil {
		t.Fatalf("Failed to read template config: %v", err)
	}

	// Validate executor config
	if config.ExecutorConfig == nil {
		t.Fatal("ExecutorConfig is nil")
	}

	// Validate basic executor fields
	if config.ExecutorConfig.BasePath == "" {
		t.Error("BasePath should not be empty")
	}
	if config.ExecutorConfig.Port == 0 {
		t.Error("Port should not be zero")
	}
	if config.ExecutorConfig.TickLength == 0 {
		t.Error("TickLength should not be zero")
	}

	// Validate URLs
	if config.ExecutorConfig.SensorUnitURL == "" {
		t.Error("SensorUnitURL should not be empty")
	}
	if config.ExecutorConfig.PowerUnitURL == "" {
		t.Error("PowerUnitURL should not be empty")
	}
	if config.ExecutorConfig.StatusMessageURL == "" {
		t.Error("StatusMessageURL should not be empty")
	}

	// Validate network interface
	if config.ExecutorConfig.NetworkInterface == "" {
		t.Error("NetworkInterface should not be empty")
	}

	// Validate defaults section
	if config.ExecutorConfig.Defaults == nil {
		t.Fatal("Defaults section is nil")
	}

	// Validate PID settings
	if config.ExecutorConfig.Defaults.PidSettings == nil {
		t.Fatal("PidSettings is nil")
	}

	// Check that acclimate PID settings exist
	acclimate, exists := config.ExecutorConfig.Defaults.PidSettings[types.StepTypeAcclimate]
	if !exists {
		t.Error("Acclimate PID settings should exist")
	} else if acclimate == nil {
		t.Error("Acclimate PID settings should not be nil")
	} else {
		if acclimate.Kp == 0 {
			t.Error("Acclimate Kp should not be zero")
		}
		if acclimate.Ki == 0 {
			t.Error("Acclimate Ki should not be zero")
		}
		if acclimate.Kd == 0 {
			t.Error("Acclimate Kd should not be zero")
		}
	}

	// Check that cooling and heating settings may be nil (they're optional)
	// But the keys should exist in the map structure if they're defined in config
	// Since they're not in the template config, we don't test for their existence

	// Validate delta heating values
	if config.ExecutorConfig.Defaults.MaxDeltaHeating == 0 {
		t.Error("MaxDeltaHeating should not be zero")
	}
	if config.ExecutorConfig.Defaults.MinDeltaHeating == 0 {
		t.Error("MinDeltaHeating should not be zero")
	}
	if config.ExecutorConfig.Defaults.MaxDeltaHeating <= config.ExecutorConfig.Defaults.MinDeltaHeating {
		t.Error("MaxDeltaHeating should be greater than MinDeltaHeating")
	}

	// Validate power unit config
	if config.PowerUnit == nil {
		t.Fatal("PowerUnit is nil")
	}
	if config.PowerUnit.ShellyAddress == "" {
		t.Error("ShellyAddress should not be empty")
	}
	if config.PowerUnit.CycleLength == 0 {
		t.Error("CycleLength should not be zero")
	}
	if config.PowerUnit.MaxIdleTime == 0 {
		t.Error("MaxIdleTime should not be zero")
	}
	if config.PowerUnit.PowerMapping == nil {
		t.Error("PowerMapping should not be nil")
	}

	// Validate sensor unit config
	if config.SensorUnit == nil {
		t.Fatal("SensorUnit is nil")
	}
	if config.SensorUnit.SerialDevice == "" {
		t.Error("SerialDevice should not be empty")
	}
	if config.SensorUnit.BaudRate == 0 {
		t.Error("BaudRate should not be zero")
	}
}

func TestConfigStructure(t *testing.T) {
	// Test that the new structure correctly groups temperature control settings
	config, err := types.ReadHalkoConfig("../templates/halko.cfg")
	if err != nil {
		t.Fatalf("Failed to read template config: %v", err)
	}

	// Verify that the defaults structure contains the expected fields
	defaults := config.ExecutorConfig.Defaults
	if defaults == nil {
		t.Fatal("Defaults should not be nil")
	}

	// Check that the expected PID settings are present in the template config
	// Only acclimate is defined in the template, cooling and heating are optional
	acclimate, exists := defaults.PidSettings[types.StepTypeAcclimate]
	if !exists {
		t.Error("Acclimate PID settings should exist in template config")
	}
	if acclimate == nil {
		t.Error("Acclimate PID settings should not be nil in template config")
	}

	// Cooling and heating may or may not be present depending on the config
	// They are not required to be in the template

	// Verify that delta heating values are reasonable
	if defaults.MaxDeltaHeating < 1.0 || defaults.MaxDeltaHeating > 100.0 {
		t.Errorf("MaxDeltaHeating value %f seems unreasonable", defaults.MaxDeltaHeating)
	}
	if defaults.MinDeltaHeating < 0.1 || defaults.MinDeltaHeating > 50.0 {
		t.Errorf("MinDeltaHeating value %f seems unreasonable", defaults.MinDeltaHeating)
	}
}

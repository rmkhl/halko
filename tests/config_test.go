package tests

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/rmkhl/halko/types"
)

func TestConfigReading(t *testing.T) {
	// Test business logic validation using a valid config
	configPath := createTestConfigFile(t)
	config, err := types.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test specific business logic that's not covered by basic validation
	// Verify that delta heating values are sensible
	if config.ControlUnitConfig.Defaults.MaxDeltaHeating <= config.ControlUnitConfig.Defaults.MinDeltaHeating {
		t.Error("MaxDeltaHeating should be greater than MinDeltaHeating")
	}

	// Check that acclimate PID settings exist and have reasonable values
	acclimate, exists := config.ControlUnitConfig.Defaults.PidSettings[types.StepTypeAcclimate]
	if exists && acclimate != nil {
		if acclimate.Kp <= 0 || acclimate.Ki <= 0 || acclimate.Kd < 0 {
			t.Error("PID values should be positive (Kp, Ki > 0, Kd >= 0)")
		}
	}
}

func TestConfigStructure(t *testing.T) {
	// Test that the configuration structure groups temperature control settings correctly
	configPath := createTestConfigFile(t)
	config, err := types.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test business logic constraints that go beyond basic validation
	defaults := config.ControlUnitConfig.Defaults

	// Verify that delta heating values are reasonable for the use case
	if defaults.MaxDeltaHeating < 1.0 || defaults.MaxDeltaHeating > 100.0 {
		t.Errorf("MaxDeltaHeating value %f seems unreasonable for temperature control", defaults.MaxDeltaHeating)
	}
	if defaults.MinDeltaHeating < 0.1 || defaults.MinDeltaHeating > 50.0 {
		t.Errorf("MinDeltaHeating value %f seems unreasonable for temperature control", defaults.MinDeltaHeating)
	}

	// Verify PID settings structure contains expected step types
	if len(defaults.PidSettings) == 0 {
		t.Error("PID settings should contain at least one step type configuration")
	}
}

// Test configuration data - represents a complete valid Halko configuration
// This configuration matches the new endpoint structure
var testConfigData = `{
  "executor": {
    "port": 8090,
    "base_path": "/tmp/test/halko",
    "tick_length": 6000,
    "network_interface": "enp4s0",
    "defaults": {
      "pid_settings": {
        "acclimate": {
          "kp": 2.0,
          "ki": 1.0,
          "kd": 0.5
        }
      },
      "max_delta_heating": 10.0,
      "min_delta_heating": 5.0
    }
  },
  "power_unit": {
    "port": 8092,
    "shelly_address": "http://localhost:8088",
    "cycle_length": 60,
    "max_idle_time": 70,
    "power_mapping": {
      "heater": 0,
      "humidifier": 1,
      "fan": 2
    }
  },
  "storage": {
    "port": 8091,
    "base_path": "/tmp/test/halko"
  },
  "sensorunit": {
    "port": 8093,
    "serial_device": "/dev/ttyUSB0",
    "baud_rate": 9600
  },
  "api_endpoints": {
    "executor": {
      "url": "http://localhost:8090",
      "status": "/status",
      "programs": "/programs",
      "running": "/running"
    },
    "sensorunit": {
      "url": "http://localhost:8088",
      "status": "/status",
      "temperatures": "/temperatures",
      "display": "/display"
    },
    "powerunit": {
      "url": "http://localhost:8092",
      "status": "/status",
      "power": "/power"
    },
    "storage": {
      "url": "http://localhost:8091",
      "status": "/status",
      "programs": "/programs",
      "execution_log": "/execution_log"
    }
  }
}`

func createTestConfigFile(t *testing.T) string {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test_halko.cfg")

	err := os.WriteFile(configPath, []byte(testConfigData), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	return configPath
}

func TestServiceEndpointEmbedding(t *testing.T) {
	configPath := createTestConfigFile(t)

	config, err := types.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test that the embedded Endpoint struct methods work correctly
	// This tests the actual functionality, not just configuration presence
	expectedExecutorStatusURL := "http://localhost:8090/status"
	if config.APIEndpoints.ControlUnit.GetStatusURL() != expectedExecutorStatusURL {
		t.Errorf("Expected Executor status URL %s, got %s", expectedExecutorStatusURL, config.APIEndpoints.ControlUnit.GetStatusURL())
	}

	expectedPowerUnitStatusURL := "http://localhost:8092/status"
	if config.APIEndpoints.PowerUnit.GetStatusURL() != expectedPowerUnitStatusURL {
		t.Errorf("Expected PowerUnit status URL %s, got %s", expectedPowerUnitStatusURL, config.APIEndpoints.PowerUnit.GetStatusURL())
	}
}

func TestGetServiceURLs(t *testing.T) {
	configPath := createTestConfigFile(t)

	config, err := types.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test endpoint URL methods
	if config.APIEndpoints.ControlUnit.GetURL() != "http://localhost:8090" {
		t.Errorf("Expected Executor URL http://localhost:8090, got %s", config.APIEndpoints.ControlUnit.GetURL())
	}

	if config.APIEndpoints.PowerUnit.GetURL() != "http://localhost:8092" {
		t.Errorf("Expected PowerUnit URL http://localhost:8092, got %s", config.APIEndpoints.PowerUnit.GetURL())
	}

	if config.APIEndpoints.SensorUnit.GetURL() != "http://localhost:8088" {
		t.Errorf("Expected SensorUnit URL http://localhost:8088, got %s", config.APIEndpoints.SensorUnit.GetURL())
	}

	// Test specific endpoint methods
	expectedProgramsURL := "http://localhost:8090/programs"
	if config.APIEndpoints.ControlUnit.GetProgramsURL() != expectedProgramsURL {
		t.Errorf("Expected Programs URL %s, got %s", expectedProgramsURL, config.APIEndpoints.ControlUnit.GetProgramsURL())
	}

	expectedEngineURL := "http://localhost:8090/engine"
	if config.APIEndpoints.ControlUnit.GetEngineURL() != expectedEngineURL {
		t.Errorf("Expected Engine URL %s, got %s", expectedEngineURL, config.APIEndpoints.ControlUnit.GetEngineURL())
	}

	expectedTemperaturesURL := "http://localhost:8088/temperatures"
	if config.APIEndpoints.SensorUnit.GetTemperaturesURL() != expectedTemperaturesURL {
		t.Errorf("Expected Temperatures URL %s, got %s", expectedTemperaturesURL, config.APIEndpoints.SensorUnit.GetTemperaturesURL())
	}
}

func TestGetPortMethod(t *testing.T) {
	configPath := createTestConfigFile(t)

	config, err := types.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test GetPort method with explicit ports
	tests := []struct {
		name         string
		endpoint     *types.Endpoint
		expectedPort string
	}{
		{"Executor", &config.APIEndpoints.ControlUnit.Endpoint, "8090"},
		{"PowerUnit", &config.APIEndpoints.PowerUnit.Endpoint, "8092"},
		{"SensorUnit", &config.APIEndpoints.SensorUnit.Endpoint, "8088"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			port, err := tt.endpoint.GetPort()
			if err != nil {
				t.Errorf("Unexpected error getting port for %s: %v", tt.name, err)
			}
			if port != tt.expectedPort {
				t.Errorf("Expected port %s for %s, got %s", tt.expectedPort, tt.name, port)
			}
		})
	}
}

func TestJSONMarshaling(t *testing.T) {
	configPath := createTestConfigFile(t)

	config, err := types.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test that the configuration can be marshaled back to JSON
	jsonData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal config to JSON: %v", err)
	}

	// Test that we can unmarshal it back
	var newConfig types.HalkoConfig
	err = json.Unmarshal(jsonData, &newConfig)
	if err != nil {
		t.Fatalf("Failed to unmarshal config from JSON: %v", err)
	}

	// Verify that endpoint URLs are preserved
	if newConfig.APIEndpoints.ControlUnit.GetURL() != "http://localhost:8090" {
		t.Error("Executor endpoint URL not preserved after JSON round-trip")
	}
	if newConfig.APIEndpoints.PowerUnit.GetURL() != "http://localhost:8092" {
		t.Error("PowerUnit endpoint URL not preserved after JSON round-trip")
	}
	if newConfig.APIEndpoints.SensorUnit.GetURL() != "http://localhost:8088" {
		t.Error("SensorUnit endpoint URL not preserved after JSON round-trip")
	}

	// Verify that endpoint paths are preserved
	if newConfig.APIEndpoints.ControlUnit.Programs != "/programs" {
		t.Error("ControlUnit programs path not preserved after JSON round-trip")
	}
	if newConfig.APIEndpoints.ControlUnit.Engine != "/engine" {
		t.Error("ControlUnit engine path not preserved after JSON round-trip")
	}
	if newConfig.APIEndpoints.SensorUnit.Temperatures != "/temperatures" {
		t.Error("SensorUnit temperatures path not preserved after JSON round-trip")
	}
}

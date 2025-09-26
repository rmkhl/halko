package tests

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/rmkhl/halko/types"
)

const testHost = "localhost"

func TestConfigReading(t *testing.T) {
	// Test reading the template configuration file
	// LoadConfig already validates the configuration, so we just need to ensure it loads successfully
	config, err := types.LoadConfig("../templates/halko.cfg")
	if err != nil {
		t.Fatalf("Failed to read template config: %v", err)
	}

	// Test specific business logic that's not covered by basic validation
	// Verify that delta heating values are sensible
	if config.ExecutorConfig.Defaults.MaxDeltaHeating <= config.ExecutorConfig.Defaults.MinDeltaHeating {
		t.Error("MaxDeltaHeating should be greater than MinDeltaHeating")
	}

	// Check that acclimate PID settings exist and have reasonable values
	acclimate, exists := config.ExecutorConfig.Defaults.PidSettings[types.StepTypeAcclimate]
	if exists && acclimate != nil {
		if acclimate.Kp <= 0 || acclimate.Ki <= 0 || acclimate.Kd < 0 {
			t.Error("PID values should be positive (Kp, Ki > 0, Kd >= 0)")
		}
	}
}

func TestConfigStructure(t *testing.T) {
	// Test that the configuration structure groups temperature control settings correctly
	config, err := types.LoadConfig("../templates/halko.cfg")
	if err != nil {
		t.Fatalf("Failed to read template config: %v", err)
	}

	// Test business logic constraints that go beyond basic validation
	defaults := config.ExecutorConfig.Defaults

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
// This configuration includes host attributes for all services to test the ServiceEndpoint refactoring
var testConfigData = `{
  "executor": {
    "host": "localhost",
    "port": 8090,
    "base_path": "/tmp/test/halko",
    "tick_length": 6000,
    "sensor_unit_host": "localhost:8088",
    "power_unit_host": "localhost:8092",
    "status_message_url": "http://localhost:8088/sensors/api/v1/status",
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
    "host": "localhost",
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
    "host": "localhost",
    "port": 8091,
    "base_path": "/tmp/test/halko"
  },
  "sensorunit": {
    "host": "localhost",
    "port": 8093,
    "serial_device": "/dev/ttyUSB0",
    "baud_rate": 9600
  },
  "api_endpoints": {
    "programs": "/programs",
    "running": "/running",
    "temperatures": "/temperatures",
    "status": "/status",
    "root": "/"
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

	// Test that embedded ServiceEndpoint fields are properly accessible
	tests := []struct {
		name         string
		expectedHost string
		expectedPort int
		actualHost   string
		actualPort   int
	}{
		{"ExecutorConfig", testHost, 8090, config.ExecutorConfig.Host, config.ExecutorConfig.Port},
		{"PowerUnit", testHost, 8092, config.PowerUnit.Host, config.PowerUnit.Port},
		{"SensorUnit", testHost, 8093, config.SensorUnit.Host, config.SensorUnit.Port},
		{"StorageConfig", testHost, 8091, config.StorageConfig.Host, config.StorageConfig.Port},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.actualHost != tt.expectedHost {
				t.Errorf("Expected %s host to be %s, got %s", tt.name, tt.expectedHost, tt.actualHost)
			}
			if tt.actualPort != tt.expectedPort {
				t.Errorf("Expected %s port to be %d, got %d", tt.name, tt.expectedPort, tt.actualPort)
			}
		})
	}
}

func TestGetBaseURL(t *testing.T) {
	configPath := createTestConfigFile(t)

	config, err := types.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	tests := []struct {
		service     types.ServiceType
		expectedURL string
		expectError bool
	}{
		{types.ServiceExecutor, "http://localhost:8090", false},
		{types.ServicePowerUnit, "http://localhost:8092", false},
		{types.ServiceSensorUnit, "http://localhost:8093", false},
		{types.ServiceStorage, "http://localhost:8091", false},
		{"unknown", "", true},
	}

	for _, tt := range tests {
		t.Run(string(tt.service), func(t *testing.T) {
			url, err := config.GetBaseURL(tt.service)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for service %s, but got none", tt.service)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for service %s: %v", tt.service, err)
				}
				if url != tt.expectedURL {
					t.Errorf("Expected URL %s for service %s, got %s", tt.expectedURL, tt.service, url)
				}
			}
		})
	}
}

func TestGetBaseURLHelperMethods(t *testing.T) {
	configPath := createTestConfigFile(t)

	config, err := types.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	tests := []struct {
		name        string
		method      func() (string, error)
		expectedURL string
	}{
		{"GetExecutorURL", config.GetExecutorURL, "http://localhost:8090"},
		{"GetPowerUnitURL", config.GetPowerUnitURL, "http://localhost:8092"},
		{"GetSensorUnitURL", config.GetSensorUnitURL, "http://localhost:8093"},
		{"GetStorageURL", config.GetStorageURL, "http://localhost:8091"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url, err := tt.method()
			if err != nil {
				t.Errorf("Unexpected error in %s: %v", tt.name, err)
			}
			if url != tt.expectedURL {
				t.Errorf("Expected URL %s from %s, got %s", tt.expectedURL, tt.name, url)
			}
		})
	}
}

func TestServiceEndpointConfigurationValidation(t *testing.T) {
	configPath := createTestConfigFile(t)

	// LoadConfig already calls ValidateRequired(), so successful loading means validation passed
	config, err := types.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test that the ServiceEndpoint embedding works correctly with validation
	// This is testing the integration between the refactored structure and validation
	if config.ExecutorConfig.Host == "" || config.ExecutorConfig.Port <= 0 {
		t.Error("ServiceEndpoint validation should ensure host and port are set")
	}
	if config.PowerUnit.Host == "" || config.PowerUnit.Port <= 0 {
		t.Error("ServiceEndpoint validation should ensure host and port are set")
	}
	if config.SensorUnit.Host == "" || config.SensorUnit.Port <= 0 {
		t.Error("ServiceEndpoint validation should ensure host and port are set")
	}
	if config.StorageConfig.Host == "" || config.StorageConfig.Port <= 0 {
		t.Error("ServiceEndpoint validation should ensure host and port are set")
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

	// Verify that the embedded ServiceEndpoint fields are preserved
	if newConfig.ExecutorConfig.Host != testHost || newConfig.ExecutorConfig.Port != 8090 {
		t.Error("ExecutorConfig ServiceEndpoint fields not preserved after JSON round-trip")
	}
	if newConfig.PowerUnit.Host != testHost || newConfig.PowerUnit.Port != 8092 {
		t.Error("PowerUnit ServiceEndpoint fields not preserved after JSON round-trip")
	}
	if newConfig.SensorUnit.Host != "localhost" || newConfig.SensorUnit.Port != 8093 {
		t.Error("SensorUnit ServiceEndpoint fields not preserved after JSON round-trip")
	}
	if newConfig.StorageConfig.Host != "localhost" || newConfig.StorageConfig.Port != 8091 {
		t.Error("StorageConfig ServiceEndpoint fields not preserved after JSON round-trip")
	}
}

func TestServiceTypeConstants(t *testing.T) {
	// Test that service type constants are defined correctly
	expectedServices := map[types.ServiceType]string{
		types.ServiceExecutor:   "executor",
		types.ServicePowerUnit:  "power_unit",
		types.ServiceSensorUnit: "sensor_unit",
		types.ServiceStorage:    "storage",
	}

	for service, expected := range expectedServices {
		if string(service) != expected {
			t.Errorf("Expected service type %s to equal %s, got %s", service, expected, string(service))
		}
	}
}

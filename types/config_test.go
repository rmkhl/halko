package types

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigReading(t *testing.T) {
	// Test business logic validation using a valid config
	configPath := createTestConfigFile(t)
	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test specific business logic that's not covered by basic validation
	heating := config.ControlUnitConfig.Defaults.Deltas[StepTypeHeating]
	if heating == nil || heating.MaxDelta <= heating.MinDelta {
		t.Error("heating defaults should have max_delta greater than min_delta")
	}
}

func TestConfigStructure(t *testing.T) {
	// Test that the configuration structure groups temperature control settings correctly
	configPath := createTestConfigFile(t)
	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test business logic constraints that go beyond basic validation
	defaults := config.ControlUnitConfig.Defaults

	// Verify that delta heating values are reasonable for the use case
	heating := defaults.Deltas[StepTypeHeating]
	if heating.MaxDelta < 1.0 || heating.MaxDelta > 100.0 {
		t.Errorf("heating max_delta %f seems unreasonable for temperature control", heating.MaxDelta)
	}
	if heating.MinDelta < 0.1 || heating.MinDelta > 50.0 {
		t.Errorf("heating min_delta %f seems unreasonable for temperature control", heating.MinDelta)
	}
}

// Test configuration data - represents a complete valid Halko configuration
// This configuration matches the new endpoint structure
var testConfigData = `{
  "controlunit": {
    "base_path": "/tmp/test/halko",
    "tick_length": "6s",
    "network_interface": "enp4s0",
    "defaults": {
      "deltas": {
        "heating": {"min_delta": 5.0, "max_delta": 10.0},
        "acclimate": {"min_delta": -1.0, "max_delta": 3.0}
      },
      "fan_power": 0,
      "steam_power": 0,
      "max_target_temperature": 200,
      "steam_ceiling": 100,
      "sensor_timeout": "120s",
      "execution_log_interval": "60s",
      "equalize": {
        "delta": 2.0,
        "steam_prewarm": false,
        "steam_prewarm_timeout": "20m"
      }
    }
  },
  "power_unit": {
    "shelly_address": "http://localhost:8088",
    "cycle_length": "60s",
    "max_idle_time": "70s",
    "power_mapping": {
      "heater": 0,
      "steam": 1,
      "fan": 2
    }
  },
  "sensorunit": {
    "serial_device": "/dev/ttyUSB0",
    "baud_rate": 9600
  },
  "api_endpoints": {
    "controlunit": {
      "url": "http://localhost:8090",
      "status": "/status",
      "programs": "/programs",
      "engine": "/engine"
    },
    "sensorunit": {
      "url": "http://localhost:8093",
      "status": "/status",
      "temperatures": "/temperatures",
      "die_temperatures": "/temperatures/die",
      "display": "/display"
    },
    "powerunit": {
      "url": "http://localhost:8092",
      "status": "/status",
      "power": "/power"
    }
  }
}`

func createTestConfigFile(t *testing.T) string {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test_halko.cfg")

	// The simulator creates the serial device it emulates, so the test
	// config must name a path it may create rather than real hardware.
	configData := strings.Replace(testConfigData, "/dev/ttyUSB0",
		filepath.Join(tempDir, "esp32"), 1)

	err := os.WriteFile(configPath, []byte(configData), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	return configPath
}

func TestServiceEndpointEmbedding(t *testing.T) {
	configPath := createTestConfigFile(t)

	config, err := LoadConfig(configPath)
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

	config, err := LoadConfig(configPath)
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

	if config.APIEndpoints.SensorUnit.GetURL() != "http://localhost:8093" {
		t.Errorf("Expected SensorUnit URL http://localhost:8093, got %s", config.APIEndpoints.SensorUnit.GetURL())
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

	expectedTemperaturesURL := "http://localhost:8093/temperatures"
	if config.APIEndpoints.SensorUnit.GetTemperaturesURL() != expectedTemperaturesURL {
		t.Errorf("Expected Temperatures URL %s, got %s", expectedTemperaturesURL, config.APIEndpoints.SensorUnit.GetTemperaturesURL())
	}
}

func TestGetPortMethod(t *testing.T) {
	configPath := createTestConfigFile(t)

	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test GetPort method with explicit ports
	tests := []struct {
		name         string
		endpoint     *Endpoint
		expectedPort string
	}{
		{"Executor", &config.APIEndpoints.ControlUnit.Endpoint, "8090"},
		{"PowerUnit", &config.APIEndpoints.PowerUnit.Endpoint, "8092"},
		{"SensorUnit", &config.APIEndpoints.SensorUnit.Endpoint, "8093"},
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

func TestDBusUnitConfigOptional(t *testing.T) {
	// testConfigData has no dbusunit section; config must still load and validate
	configPath := createTestConfigFile(t)
	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config without dbusunit section: %v", err)
	}
	if config.DBusUnit != nil {
		t.Error("Expected DBusUnit to be nil when the dbusunit section is missing")
	}
}

func TestDBusUnitConfigParsing(t *testing.T) {
	data := `{"dbusunit": {"system_bus_socket": "/run/host/run/dbus/system_bus_socket"}}`
	var config HalkoConfig
	if err := json.Unmarshal([]byte(data), &config); err != nil {
		t.Fatalf("Failed to unmarshal dbusunit section: %v", err)
	}
	if config.DBusUnit == nil {
		t.Fatal("Expected DBusUnit to be set when the dbusunit section is present")
	}
	if config.DBusUnit.SystemBusSocket != "/run/host/run/dbus/system_bus_socket" {
		t.Errorf("Expected system_bus_socket /run/host/run/dbus/system_bus_socket, got %s", config.DBusUnit.SystemBusSocket)
	}
}

func TestJSONMarshaling(t *testing.T) {
	configPath := createTestConfigFile(t)

	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test that the configuration can be marshaled back to JSON
	jsonData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal config to JSON: %v", err)
	}

	// Test that we can unmarshal it back
	var newConfig HalkoConfig
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
	if newConfig.APIEndpoints.SensorUnit.GetURL() != "http://localhost:8093" {
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

// writeConfigWithDefaults writes the standard test config with its
// controlunit.defaults block replaced by the given JSON.
func writeConfigWithDefaults(t *testing.T, defaults string) string {
	t.Helper()
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test_halko.cfg")

	data := strings.Replace(testConfigData, "/dev/ttyUSB0", filepath.Join(tempDir, "esp32"), 1)
	start := strings.Index(data, `"defaults": {`)
	if start < 0 {
		t.Fatal("test config has no defaults block")
	}
	// The block nests, so find its close by balancing braces rather than
	// taking the first one.
	depth, end := 0, -1
	for i := strings.Index(data[start:], "{") + start; i < len(data); i++ {
		switch data[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				end = i + 1
			}
		}
		if end >= 0 {
			break
		}
	}
	if end < 0 {
		t.Fatal("unbalanced defaults block in test config")
	}
	data = data[:start] + `"defaults": ` + defaults + data[end:]

	if err := os.WriteFile(configPath, []byte(data), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return configPath
}

// A delta-controlled step type whose defaults are missing or unusable used to
// surface much later: ApplyDefaults handed the step a 0/0 band and the program
// failed validation when someone started it, naming the program rather than the
// config. Catch it while loading instead.
func TestLoadConfigRejectsUnusableDeltaDefaults(t *testing.T) {
	tests := []struct {
		name     string
		defaults string
	}{
		{
			"acclimate entry missing",
			`{"deltas": {"heating": {"min_delta": 5.0, "max_delta": 10.0}}, "fan_power": 0, "steam_power": 0, "max_target_temperature": 200, "steam_ceiling": 100, "sensor_timeout": "120s", "execution_log_interval": "60s", "equalize": {"delta": 2.0, "steam_prewarm": false, "steam_prewarm_timeout": "20m"}}`,
		},
		{
			"heating entry missing",
			`{"deltas": {"acclimate": {"min_delta": -1.0, "max_delta": 3.0}}, "fan_power": 0, "steam_power": 0, "max_target_temperature": 200, "steam_ceiling": 100, "sensor_timeout": "120s", "execution_log_interval": "60s", "equalize": {"delta": 2.0, "steam_prewarm": false, "steam_prewarm_timeout": "20m"}}`,
		},
		{
			"collapsed band",
			`{"deltas": {"heating": {"min_delta": 5.0, "max_delta": 5.0}, "acclimate": {"min_delta": -1.0, "max_delta": 3.0}}, "fan_power": 0, "steam_power": 0, "max_target_temperature": 200, "steam_ceiling": 100, "sensor_timeout": "120s", "execution_log_interval": "60s", "equalize": {"delta": 2.0, "steam_prewarm": false, "steam_prewarm_timeout": "20m"}}`,
		},
		{
			"reversed band",
			`{"deltas": {"heating": {"min_delta": 5.0, "max_delta": 10.0}, "acclimate": {"min_delta": 3.0, "max_delta": -1.0}}, "fan_power": 0, "steam_power": 0, "max_target_temperature": 200, "steam_ceiling": 100, "sensor_timeout": "120s", "execution_log_interval": "60s", "equalize": {"delta": 2.0, "steam_prewarm": false, "steam_prewarm_timeout": "20m"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := LoadConfig(writeConfigWithDefaults(t, tt.defaults)); err == nil {
				t.Fatal("expected LoadConfig to fail, got nil")
			}
		})
	}
}

func TestLoadConfigAcceptsNestedDeltaDefaults(t *testing.T) {
	path := writeConfigWithDefaults(t,
		`{"deltas": {"heating": {"min_delta": 5.0, "max_delta": 10.0}, "acclimate": {"min_delta": -1.0, "max_delta": 3.0}}, "fan_power": 0, "steam_power": 0, "max_target_temperature": 200, "steam_ceiling": 100, "sensor_timeout": "120s", "execution_log_interval": "60s", "equalize": {"delta": 2.0, "steam_prewarm": false, "steam_prewarm_timeout": "20m"}}`)

	config, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}

	acclimate := config.ControlUnitConfig.Defaults.Deltas[StepTypeAcclimate]
	if acclimate == nil {
		t.Fatal("acclimate defaults missing after load")
	}
	if acclimate.MinDelta != -1.0 || acclimate.MaxDelta != 3.0 {
		t.Fatalf("acclimate defaults = %v/%v, want -1/3", acclimate.MinDelta, acclimate.MaxDelta)
	}
}

// The startup steps' band and the steam warm-up's limit are properties of the
// installation, so both come from the configured defaults and both are
// required.
func TestEqualizeDefaultsLoad(t *testing.T) {
	path := writeConfigWithDefaults(t,
		`{"deltas": {"heating": {"min_delta": 5.0, "max_delta": 10.0}, "acclimate": {"min_delta": -1.0, "max_delta": 3.0}}, "fan_power": 0, "steam_power": 0, "max_target_temperature": 200, "steam_ceiling": 100, "sensor_timeout": "120s", "execution_log_interval": "60s", "equalize": {"delta": 2.0, "steam_prewarm": false, "steam_prewarm_timeout": "20m"}}`)

	config, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	defaults := config.ControlUnitConfig.Defaults

	if defaults.Equalize.Delta == nil || *defaults.Equalize.Delta != 2.0 {
		t.Errorf("equalize.delta = %v, want 2.0", defaults.Equalize.Delta)
	}
	if defaults.Equalize.SteamPrewarmTimeoutSeconds != 1200 {
		t.Errorf("equalize.steam_prewarm_timeout resolved to %ds, want 1200", defaults.Equalize.SteamPrewarmTimeoutSeconds)
	}
}

func TestLoadConfigRejectsUnusableEqualizeDefaults(t *testing.T) {
	const base = `{"deltas": {"heating": {"min_delta": 5.0, "max_delta": 10.0}, "acclimate": {"min_delta": -1.0, "max_delta": 3.0}}, "fan_power": 0, "steam_power": 0, "max_target_temperature": 200, "steam_ceiling": 100, "sensor_timeout": "120s", "execution_log_interval": "60s"`

	tests := []struct {
		name     string
		defaults string
	}{
		{"equalize block missing", base + `}`},
		{"delta missing", base + `, "equalize": {"steam_prewarm": false, "steam_prewarm_timeout": "20m"}}`},
		{"delta zero", base + `, "equalize": {"delta": 0, "steam_prewarm": false, "steam_prewarm_timeout": "20m"}}`},
		{"delta negative", base + `, "equalize": {"delta": -1.0, "steam_prewarm": false, "steam_prewarm_timeout": "20m"}}`},
		{"steam_prewarm missing", base + `, "equalize": {"delta": 2.0, "steam_prewarm_timeout": "20m"}}`},
		{"steam_prewarm_timeout missing", base + `, "equalize": {"delta": 2.0, "steam_prewarm": false}}`},
		{"steam_prewarm_timeout unparseable", base + `, "equalize": {"delta": 2.0, "steam_prewarm": false, "steam_prewarm_timeout": "soon"}}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := LoadConfig(writeConfigWithDefaults(t, tt.defaults)); err == nil {
				t.Fatal("expected LoadConfig to fail, got nil")
			}
		})
	}
}

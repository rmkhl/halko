package types

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type (
	Defaults struct {
		PidSettings     map[StepType]*PidSettings `json:"pid_settings"`
		MaxDeltaHeating float32                   `json:"max_delta_heating"`
		MinDeltaHeating float32                   `json:"min_delta_heating"`
	}

	ExecutorConfig struct {
		BasePath         string    `json:"base_path"`
		Port             int       `json:"port"`
		TickLength       int       `json:"tick_length"`
		PowerUnitURL     string    `json:"power_unit_url"`
		SensorUnitURL    string    `json:"sensor_unit_url"`
		StatusMessageURL string    `json:"status_message_url"`
		NetworkInterface string    `json:"network_interface"`
		Defaults         *Defaults `json:"defaults"`
	}

	PowerUnit struct {
		ShellyAddress string         `json:"shelly_address"`
		CycleLength   int            `json:"cycle_length"` // Duration of a power cycle in seconds
		PowerMapping  map[string]int `json:"power_mapping"`
		MaxIdleTime   int            `json:"max_idle_time"` // Maximum idle time in seconds before a executor is considered idle
	}

	SensorUnitConfig struct {
		SerialDevice string `json:"serial_device"`
		BaudRate     int    `json:"baud_rate"`
	}

	HalkoConfig struct {
		ExecutorConfig *ExecutorConfig   `json:"executor"`
		PowerUnit      *PowerUnit        `json:"power_unit"`
		SensorUnit     *SensorUnitConfig `json:"sensorunit"`
	}
)

// LoadConfig loads the halko configuration from the specified path or finds it in default locations
func LoadConfig(configPath string) (*HalkoConfig, error) {
	// If no config path provided, try to find default location
	if configPath == "" {
		configPath = findDefaultConfigPath()
		if configPath == "" {
			return nil, errors.New("no config file specified and none found in default locations")
		}
	}

	config, err := readHalkoConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	return config, nil
}

// findDefaultConfigPath searches for halko.cfg in common locations
func findDefaultConfigPath() string {
	possiblePaths := []string{
		"halko.cfg",
		"/etc/halko/halko.cfg",
		"/var/opt/halko/halko.cfg",
	}

	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// Try to find halko.cfg relative to the executable
	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		configPath := filepath.Join(exeDir, "halko.cfg")
		if _, err := os.Stat(configPath); err == nil {
			return configPath
		}
	}

	return ""
}

// readHalkoConfig reads the halko configuration from the specified path (private function)
func readHalkoConfig(path string) (*HalkoConfig, error) {
	jsonFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer jsonFile.Close()

	content, err := io.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}

	var config HalkoConfig
	err = json.Unmarshal(content, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

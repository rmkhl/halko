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
		PowerUnitHost    string    `json:"power_unit_host"`
		SensorUnitHost   string    `json:"sensor_unit_host"`
		StatusMessageURL string    `json:"status_message_url"`
		NetworkInterface string    `json:"network_interface"`
		Defaults         *Defaults `json:"defaults"`
	}

	PowerUnit struct {
		ShellyAddress string         `json:"shelly_address"`
		CycleLength   int            `json:"cycle_length"`
		PowerMapping  map[string]int `json:"power_mapping"`
		MaxIdleTime   int            `json:"max_idle_time"`
		Port          int            `json:"port"`
	}

	SensorUnitConfig struct {
		SerialDevice string `json:"serial_device"`
		BaudRate     int    `json:"baud_rate"`
		Port         int    `json:"port"`
	}

	StorageConfig struct {
		BasePath string `json:"base_path"`
		Port     int    `json:"port"`
	}

	HalkoConfig struct {
		ExecutorConfig *ExecutorConfig   `json:"executor"`
		PowerUnit      *PowerUnit        `json:"power_unit"`
		SensorUnit     *SensorUnitConfig `json:"sensorunit"`
		StorageConfig  *StorageConfig    `json:"storage"`
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

	// Validate required configuration values
	if err := config.ValidateRequired(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
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

// ValidateRequired validates that all required configuration values are present
func (c *HalkoConfig) ValidateRequired() error {
	if c.ExecutorConfig == nil {
		return errors.New("executor configuration is required")
	}
	if c.ExecutorConfig.Port <= 0 {
		return errors.New("executor port is required and must be positive")
	}
	if c.ExecutorConfig.SensorUnitHost == "" {
		return errors.New("sensor unit host is required")
	}
	if c.ExecutorConfig.PowerUnitHost == "" {
		return errors.New("power unit host is required")
	}
	if c.ExecutorConfig.BasePath == "" {
		return errors.New("executor base path is required")
	}
	if c.ExecutorConfig.TickLength <= 0 {
		return errors.New("executor tick length is required and must be positive")
	}

	if c.SensorUnit == nil {
		return errors.New("sensor unit configuration is required")
	}
	if c.SensorUnit.SerialDevice == "" {
		return errors.New("sensor unit serial device is required")
	}
	if c.SensorUnit.BaudRate <= 0 {
		return errors.New("sensor unit baud rate is required and must be positive")
	}
	if c.SensorUnit.Port <= 0 {
		return errors.New("sensor unit port is required and must be positive")
	}

	if c.PowerUnit == nil {
		return errors.New("power unit configuration is required")
	}
	if c.PowerUnit.ShellyAddress == "" {
		return errors.New("power unit shelly address is required")
	}
	if c.PowerUnit.CycleLength <= 0 {
		return errors.New("power unit cycle length is required and must be positive")
	}
	if c.PowerUnit.MaxIdleTime <= 0 {
		return errors.New("power unit max idle time is required and must be positive")
	}
	if c.PowerUnit.Port <= 0 {
		return errors.New("power unit port is required and must be positive")
	}
	if c.PowerUnit.PowerMapping == nil || len(c.PowerUnit.PowerMapping) == 0 {
		return errors.New("power unit power mapping is required")
	}

	return nil
}

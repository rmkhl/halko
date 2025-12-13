package types

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"

	"github.com/rmkhl/halko/types/log"
)

type (
	EndpointWithStatus interface {
		GetStatusURL() string
	}

	Endpoint struct {
		URL    string `json:"url"`
		Status string `json:"status"`
	}

	Defaults struct {
		PidSettings     map[StepType]*PidSettings `json:"pid_settings"`
		MaxDeltaHeating float32                   `json:"max_delta_heating"`
		MinDeltaHeating float32                   `json:"min_delta_heating"`
	}

	ControlUnitConfig struct {
		BasePath         string    `json:"base_path"`
		TickLength       int       `json:"tick_length"`
		NetworkInterface string    `json:"network_interface"`
		Defaults         *Defaults `json:"defaults"`
	}

	PowerUnit struct {
		ShellyAddress string         `json:"shelly_address"`
		CycleLength   int            `json:"cycle_length"`
		PowerMapping  map[string]int `json:"power_mapping"`
		MaxIdleTime   int            `json:"max_idle_time"`
	}

	SensorUnitConfig struct {
		SerialDevice string `json:"serial_device"`
		BaudRate     int    `json:"baud_rate"`
	}

	StorageConfig struct {
		BasePath string `json:"base_path"`
	}

	ControlUnitEndpoints struct {
		Endpoint `json:",inline"`
		Engine   string `json:"engine"`
		Programs string `json:"programs"`
		Status   string `json:"status"`
	}

	SensorUnitEndpoints struct {
		Endpoint     `json:",inline"`
		Temperatures string `json:"temperatures"`
		Display      string `json:"display"`
		Status       string `json:"status"`
	}

	PowerUnitEndpoints struct {
		Endpoint `json:",inline"`
		Power    string `json:"power"`
		Status   string `json:"status"`
	}

	APIEndpoints struct {
		ControlUnit ControlUnitEndpoints `json:"controlunit"`
		SensorUnit  SensorUnitEndpoints  `json:"sensorunit"`
		PowerUnit   PowerUnitEndpoints   `json:"powerunit"`
	}

	HalkoConfig struct {
		ControlUnitConfig *ControlUnitConfig `json:"controlunit"`
		PowerUnit         *PowerUnit         `json:"power_unit"`
		SensorUnit        *SensorUnitConfig  `json:"sensorunit"`
		APIEndpoints      *APIEndpoints      `json:"api_endpoints"`
	}
)

// GetURL returns the base URL for this endpoint
func (e *Endpoint) GetURL() string {
	return e.URL
}

// GetStatusURL returns the full status endpoint URL
func (e *Endpoint) GetStatusURL() string {
	return e.URL + e.Status
}

// GetStatusURL returns the full status endpoint URL for ControlUnitEndpoints
func (e *ControlUnitEndpoints) GetStatusURL() string {
	return e.URL + e.Status
}

// GetStatusURL returns the full status endpoint URL for SensorUnitEndpoints
func (e *SensorUnitEndpoints) GetStatusURL() string {
	return e.URL + e.Status
}

// GetStatusURL returns the full status endpoint URL for PowerUnitEndpoints
func (e *PowerUnitEndpoints) GetStatusURL() string {
	return e.URL + e.Status
}

// GetPort extracts the port as a string from the URL or provided urlStr
// Returns "80" for HTTP and "443" for HTTPS if port is not explicitly specified
func (e *Endpoint) GetPort(urlStr ...string) (string, error) {
	var targetURL string
	if len(urlStr) > 0 && urlStr[0] != "" {
		targetURL = urlStr[0]
	} else {
		targetURL = e.URL
	}

	if targetURL == "" {
		return "", errors.New("empty URL")
	}

	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse URL: %w", err)
	}

	port := parsedURL.Port()
	if port != "" {
		return port, nil
	}

	// Use standard ports based on scheme
	switch parsedURL.Scheme {
	case "http":
		return "80", nil
	case "https":
		return "443", nil
	default:
		return "", fmt.Errorf("unsupported scheme: %s", parsedURL.Scheme)
	}
}

// ControlUnitEndpoints methods
func (e *ControlUnitEndpoints) GetProgramsURL() string {
	return e.URL + e.Programs
}

func (e *ControlUnitEndpoints) GetEngineURL() string {
	return e.URL + e.Engine
}

// SensorUnitEndpoints methods
func (e *SensorUnitEndpoints) GetTemperaturesURL() string {
	return e.URL + e.Temperatures
}

func (e *SensorUnitEndpoints) GetDisplayURL() string {
	return e.URL + e.Display
}

// PowerUnitEndpoints methods
func (e *PowerUnitEndpoints) GetPowerURL() string {
	return e.URL + e.Power
}

func LoadConfig(configPath string) (*HalkoConfig, error) {
	if configPath == "" {
		configPath = findDefaultConfigPath()
		if configPath == "" {
			return nil, errors.New("no config file specified and none found in default locations")
		}
	}

	log.Info("Loading configuration from: %s", configPath)

	config, err := readHalkoConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	if err := config.ValidateRequired(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	log.Info("Configuration loaded successfully from: %s", configPath)
	return config, nil
}

func findDefaultConfigPath() string {
	// Check environment variable first
	if configPath := os.Getenv("HALKO_CONFIG"); configPath != "" {
		if _, err := os.Stat(configPath); err == nil {
			return configPath
		}
	}

	// Define search paths in priority order
	searchPaths := []string{
		"halko.cfg", // Current directory
	}

	if homeDir, err := os.UserHomeDir(); err == nil {
		searchPaths = append(searchPaths,
			filepath.Join(homeDir, ".halko.cfg"),           // ~/.halko.cfg
			filepath.Join(homeDir, ".config", "halko.cfg"), // ~/.config/halko.cfg
		)
	}

	searchPaths = append(searchPaths,
		"/etc/halko/halko.cfg",     // System config directory
		"/etc/opt/halko/halko.cfg", // Optional system config directory
	)

	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		executablePath := filepath.Join(exeDir, "halko.cfg")
		searchPaths = append(searchPaths, executablePath)
	}

	for _, path := range searchPaths {
		if _, err := os.Stat(path); err == nil {
			return path
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

func (c *HalkoConfig) ValidateRequired() error {
	if c.ControlUnitConfig == nil {
		return errors.New("controlunit configuration is required")
	}
	if c.ControlUnitConfig.BasePath == "" {
		return errors.New("controlunit base path is required")
	}
	if c.ControlUnitConfig.TickLength <= 0 {
		return errors.New("controlunit tick length is required and must be positive")
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
	if len(c.PowerUnit.PowerMapping) == 0 {
		return errors.New("power unit power mapping is required")
	}

	if c.APIEndpoints == nil {
		return errors.New("API endpoints configuration is required")
	}

	// Validate controlunit endpoints
	if c.APIEndpoints.ControlUnit.URL == "" {
		return errors.New("controlunit endpoints URL is required")
	}
	if c.APIEndpoints.ControlUnit.Engine == "" {
		return errors.New("controlunit endpoints engine path is required")
	}
	if c.APIEndpoints.ControlUnit.Programs == "" {
		return errors.New("controlunit endpoints programs path is required")
	}
	if c.APIEndpoints.ControlUnit.Status == "" {
		return errors.New("controlunit endpoints status path is required")
	}

	// Validate sensorunit endpoints
	if c.APIEndpoints.SensorUnit.URL == "" {
		return errors.New("sensorunit endpoints URL is required")
	}
	if c.APIEndpoints.SensorUnit.Temperatures == "" {
		return errors.New("sensorunit endpoints temperatures path is required")
	}
	if c.APIEndpoints.SensorUnit.Display == "" {
		return errors.New("sensorunit endpoints display path is required")
	}
	if c.APIEndpoints.SensorUnit.Status == "" {
		return errors.New("sensorunit endpoints status path is required")
	}

	// Validate powerunit endpoints
	if c.APIEndpoints.PowerUnit.URL == "" {
		return errors.New("powerunit endpoints URL is required")
	}
	if c.APIEndpoints.PowerUnit.Status == "" {
		return errors.New("powerunit endpoints status path is required")
	}
	if c.APIEndpoints.PowerUnit.Power == "" {
		return errors.New("powerunit endpoints power path is required")
	}

	return nil
}

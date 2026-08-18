package types

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"time"

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

	// DeltaSettings is the band a step type's delta controller starts from when
	// a program names no heater of its own. What the band is measured against
	// depends on the step type, so the numbers are not comparable across them:
	// a heating step bands the kiln around the material and both values are
	// positive, an acclimate bands it around the target and straddles zero.
	DeltaSettings struct {
		MinDelta float32 `json:"min_delta"`
		MaxDelta float32 `json:"max_delta"`
	}

	// Defaults carries every value the control unit would otherwise have to
	// invent: the delta band each step type falls back to, the power a program
	// gets for components it does not name, and the kiln's own ceiling. All of
	// it is required, so a missing entry fails at load rather than turning into
	// a zero somewhere downstream.
	Defaults struct {
		Deltas map[StepType]*DeltaSettings `json:"deltas"`
		// Power applied to a component a step leaves unspecified. Pointers so
		// that "absent" is distinguishable from a deliberate 0.
		FanPower   *uint8 `json:"fan_power"`
		SteamPower *uint8 `json:"steam_power"`
		// Power the fan runs at while preheating, before the first step.
		PreheatFanPower *uint8 `json:"preheat_fan_power"`
		// Highest target any step may ask for.
		MaxTargetTemperature *uint8 `json:"max_target_temperature"`
		// Temperature steam cannot heat the kiln past. Above it steam is
		// thermally neutral; below it steam outruns the heater.
		SteamCeiling *uint8 `json:"steam_ceiling"`
		// How long a sensor may go without a valid reading before the running
		// program is failed and all power switched off.
		SensorTimeout string `json:"sensor_timeout"`
		// How often a running program appends a line to its execution log.
		ExecutionLogInterval string `json:"execution_log_interval"`
	}

	ControlUnitConfig struct {
		BasePath         string    `json:"base_path"`
		TickLength       string    `json:"tick_length"`
		NetworkInterface string    `json:"network_interface"`
		Defaults         *Defaults `json:"defaults"`
	}

	PowerUnit struct {
		ShellyAddress string         `json:"shelly_address"`
		CycleLength   string         `json:"cycle_length"`
		PowerMapping  map[string]int `json:"power_mapping"`
		MaxIdleTime   string         `json:"max_idle_time"`
	}

	SensorUnitConfig struct {
		SerialDevice string `json:"serial_device"`
		BaudRate     int    `json:"baud_rate"`
	}

	DBusUnitConfig struct {
		SystemBusSocket string `json:"system_bus_socket"`
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

	DBusUnitEndpoints struct {
		Endpoint `json:",inline"`
		VPN      string `json:"vpn"`
		Power    string `json:"power"`
		Status   string `json:"status"`
	}

	APIEndpoints struct {
		ControlUnit ControlUnitEndpoints `json:"controlunit"`
		SensorUnit  SensorUnitEndpoints  `json:"sensorunit"`
		PowerUnit   PowerUnitEndpoints   `json:"powerunit"`
		DBusUnit    DBusUnitEndpoints    `json:"dbusunit"`
	}

	HalkoConfig struct {
		ControlUnitConfig *ControlUnitConfig `json:"controlunit"`
		PowerUnit         *PowerUnit         `json:"power_unit"`
		SensorUnit        *SensorUnitConfig  `json:"sensorunit"`
		DBusUnit          *DBusUnitConfig    `json:"dbusunit"`
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

// GetStatusURL returns the full status endpoint URL for DBusUnitEndpoints
func (e *DBusUnitEndpoints) GetStatusURL() string {
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

// FindConfigFile searches for a configuration file in standard locations
// filename: name of config file to search for (e.g., "halko.cfg", "simulator.conf")
// envVar: environment variable name to check (e.g., "HALKO_CONFIG", "SIMULATOR_CONFIG")
// Returns empty string if not found
func FindConfigFile(filename string, envVar string) string {
	// 1. Check environment variable first
	if envVar != "" {
		if configPath := os.Getenv(envVar); configPath != "" {
			if _, err := os.Stat(configPath); err == nil {
				return configPath
			}
		}
	}

	// 2. Current directory
	if _, err := os.Stat(filename); err == nil {
		return filename
	}

	// 3. Home directory variations
	if homeDir, err := os.UserHomeDir(); err == nil {
		paths := []string{
			filepath.Join(homeDir, "."+filename),
			filepath.Join(homeDir, ".config", filename),
		}
		for _, path := range paths {
			if _, err := os.Stat(path); err == nil {
				return path
			}
		}
	}

	// 4. System directories. /etc/opt is last but is the one that matters in
	// practice: `make install` writes the config straight to /etc/opt/halko.cfg
	// and the systemd units name it explicitly, so without this entry the
	// production config is invisible to anything relying on auto-discovery
	// (halkoctl in particular).
	systemPaths := []string{
		filepath.Join("/etc/halko", filename),
		filepath.Join("/etc/opt/halko", filename),
		filepath.Join("/etc/opt", filename),
	}
	for _, path := range systemPaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// 5. Executable directory
	if exePath, err := os.Executable(); err == nil {
		path := filepath.Join(filepath.Dir(exePath), filename)
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}

func findDefaultConfigPath() string {
	return FindConfigFile("halko.cfg", "HALKO_CONFIG")
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
	if c.ControlUnitConfig.TickLength == "" {
		return errors.New("controlunit tick_length is required")
	}
	if _, err := time.ParseDuration(c.ControlUnitConfig.TickLength); err != nil {
		return fmt.Errorf("controlunit tick_length must be a valid duration (e.g., '6s', '100ms'): %w", err)
	}

	// Everything the control unit falls back to has to be present and usable.
	// Without this a missing entry arrives as a zero and the failure only
	// surfaces when someone starts a program, naming the program rather than
	// the config.
	if c.ControlUnitConfig.Defaults == nil {
		return errors.New("controlunit defaults are required")
	}
	defaults := c.ControlUnitConfig.Defaults
	if defaults.FanPower == nil {
		return errors.New("controlunit defaults: fan_power is required")
	}
	if defaults.SteamPower == nil {
		return errors.New("controlunit defaults: steam_power is required")
	}
	if *defaults.FanPower > 100 || *defaults.SteamPower > 100 {
		return errors.New("controlunit defaults: fan_power and steam_power must be percentages")
	}
	if defaults.MaxTargetTemperature == nil || *defaults.MaxTargetTemperature == 0 {
		return errors.New("controlunit defaults: max_target_temperature is required")
	}
	if defaults.SteamCeiling == nil || *defaults.SteamCeiling == 0 {
		return errors.New("controlunit defaults: steam_ceiling is required")
	}
	if defaults.PreheatFanPower == nil {
		return errors.New("controlunit defaults: preheat_fan_power is required")
	}
	if *defaults.PreheatFanPower > 100 {
		return errors.New("controlunit defaults: preheat_fan_power must be a percentage")
	}
	for _, d := range []struct {
		name, value string
	}{
		{"sensor_timeout", defaults.SensorTimeout},
		{"execution_log_interval", defaults.ExecutionLogInterval},
	} {
		if d.value == "" {
			return fmt.Errorf("controlunit defaults: %s is required", d.name)
		}
		if _, err := time.ParseDuration(d.value); err != nil {
			return fmt.Errorf("controlunit defaults: %s must be a valid duration: %w", d.name, err)
		}
	}
	for _, stepType := range []StepType{StepTypeHeating, StepTypeAcclimate} {
		band := defaults.Deltas[stepType]
		if band == nil {
			return fmt.Errorf("controlunit defaults are missing a %q entry", stepType)
		}
		if band.MinDelta >= band.MaxDelta {
			return fmt.Errorf("controlunit %q defaults: min_delta must be below max_delta", stepType)
		}
	}
	// An acclimate bands the kiln around the target, so its floor must sit
	// below the target and its ceiling above it.
	if acclimate := defaults.Deltas[StepTypeAcclimate]; acclimate.MinDelta >= 0 || acclimate.MaxDelta <= 0 {
		return errors.New(`controlunit "acclimate" defaults: min_delta must be negative and max_delta positive`)
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
	if c.PowerUnit.CycleLength == "" {
		return errors.New("power unit cycle_length is required")
	}
	if _, err := time.ParseDuration(c.PowerUnit.CycleLength); err != nil {
		return fmt.Errorf("power unit cycle_length must be a valid duration (e.g., '60s', '1m'): %w", err)
	}
	if c.PowerUnit.MaxIdleTime == "" {
		return errors.New("power unit max_idle_time is required")
	}
	if _, err := time.ParseDuration(c.PowerUnit.MaxIdleTime); err != nil {
		return fmt.Errorf("power unit max_idle_time must be a valid duration (e.g., '70s', '1m10s'): %w", err)
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

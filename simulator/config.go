package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/rmkhl/halko/types"
	"github.com/rmkhl/halko/types/log"
)

type SimulatorConfig struct {
	TickDuration        string  `json:"tick_duration"`
	StatusInterval      int     `json:"status_interval"`
	InitialOvenTemp     float64 `json:"initial_oven_temp"`
	InitialMaterialTemp float64 `json:"initial_material_temp"`
	EnvironmentTemp     float64 `json:"environment_temp"`
}

func LoadSimulatorConfig(configPath string) (*SimulatorConfig, error) {
	if configPath == "" {
		configPath = types.FindConfigFile("simulator.conf", "SIMULATOR_CONFIG")
		if configPath == "" {
			return nil, errors.New("no simulator configuration file specified and none found in default locations")
		}
	}

	log.Info("Loading simulator configuration from: %s", configPath)

	file, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open simulator config file %s: %w", configPath, err)
	}
	defer file.Close()

	var config SimulatorConfig
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to parse simulator config file: %w", err)
	}

	// Validate required fields and tick duration format
	if config.TickDuration == "" {
		return nil, errors.New("tick_duration is required in simulator config")
	}
	if _, err := time.ParseDuration(config.TickDuration); err != nil {
		return nil, fmt.Errorf("invalid tick_duration '%s': %w", config.TickDuration, err)
	}

	log.Info("Simulator configuration loaded successfully from: %s", configPath)
	return &config, nil
}

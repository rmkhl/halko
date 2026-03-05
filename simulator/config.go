package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/rmkhl/halko/types"
	"github.com/rmkhl/halko/types/log"
)

type SimulatorConfig struct {
	StatusInterval      int                    `json:"status_interval"`
	InitialOvenTemp     float64                `json:"initial_oven_temp"`
	InitialMaterialTemp float64                `json:"initial_material_temp"`
	EnvironmentTemp     float64                `json:"environment_temp"`
	SimulationEngine    string                 `json:"simulation_engine"`
	EngineConfig        map[string]interface{} `json:"engine_config"`
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

	// Validate simulation engine is specified
	if config.SimulationEngine == "" {
		return nil, errors.New("simulation_engine is required in simulator config")
	}

	// Validate engine config is present
	if config.EngineConfig == nil {
		return nil, fmt.Errorf("engine_config is required for simulation engine '%s'", config.SimulationEngine)
	}

	log.Info("Simulator configuration loaded successfully from: %s", configPath)
	return &config, nil
}

package main

import (
	"github.com/rmkhl/halko/types"
)

var globalConfig *types.HalkoConfig

// getExecutorAPIURL returns the base API URL for executor
func getExecutorAPIURL(config *types.HalkoConfig) string {
	if config == nil || config.APIEndpoints == nil {
		return "http://localhost:8080"
	}

	return config.APIEndpoints.Executor.GetURL()
}

// getSensorUnitAPIURL returns the base API URL for sensorunit
func getSensorUnitAPIURL(config *types.HalkoConfig) string {
	if config == nil || config.APIEndpoints == nil {
		return "http://localhost:8081"
	}

	return config.APIEndpoints.SensorUnit.GetURL()
}

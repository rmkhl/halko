package main

import (
	"github.com/rmkhl/halko/types"
)

var globalConfig *types.HalkoConfig

// getControlUnitAPIURL returns the base API URL for controlunit
func getControlUnitAPIURL(config *types.HalkoConfig) string {
	if config == nil || config.APIEndpoints == nil {
		return "http://localhost:8080"
	}

	return config.APIEndpoints.ControlUnit.GetURL()
}

// getStorageAPIURL returns the base API URL for storage
func getStorageAPIURL(config *types.HalkoConfig) string {
	if config == nil || config.APIEndpoints == nil {
		return "http://localhost:8091"
	}

	return config.APIEndpoints.ControlUnit.GetURL()
}

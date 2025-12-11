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

// getStorageAPIURL returns the base API URL for storage
func getStorageAPIURL(config *types.HalkoConfig) string {
	if config == nil || config.APIEndpoints == nil {
		return "http://localhost:8091"
	}

	return config.APIEndpoints.Storage.GetURL()
}

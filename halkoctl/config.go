package main

import (
	"github.com/rmkhl/halko/types"
)

var globalConfig *types.HalkoConfig

// getExecutorAPIURL returns the base API URL for executor
func getExecutorAPIURL(config *types.HalkoConfig) string {
	if config == nil {
		return "http://localhost:8080"
	}

	url, err := config.GetExecutorURL()
	if err != nil {
		// Fallback to default if there's an error
		return "http://localhost:8080"
	}

	return url
}

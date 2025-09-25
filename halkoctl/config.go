package main

import (
	"fmt"

	"github.com/rmkhl/halko/types"
)

var globalConfig *types.HalkoConfig

// getExecutorAPIURL returns the base API URL for executor
func getExecutorAPIURL(config *types.HalkoConfig) string {
	if config == nil || config.ExecutorConfig == nil {
		return "http://localhost:8080"
	}

	port := config.ExecutorConfig.Port
	if port == 0 {
		port = 8080
	}

	return fmt.Sprintf("http://localhost:%d", port)
}

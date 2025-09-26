package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/rmkhl/halko/types"
)

func handleTemperaturesCommand() {
	opts, err := ParseTemperaturesOptions()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		os.Exit(exitError)
	}

	if opts.Help {
		showTemperaturesHelp()
		os.Exit(exitSuccess)
	}

	sensorURL := getSensorunitAPIURL(globalConfig)
	if sensorURL == "" {
		fmt.Fprintf(os.Stderr, "Error: Could not determine sensor unit URL from config\n")
		os.Exit(exitError)
	}

	if globalOpts.Verbose {
		fmt.Printf("Querying temperatures from sensor unit at: %s\n", sensorURL)
		fmt.Println()
	}

	err = getTemperatures(sensorURL, globalOpts.Verbose)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get temperatures: %v\n", err)
		os.Exit(exitError)
	}

	os.Exit(exitSuccess)
}

func showTemperaturesHelp() {
	fmt.Println("halkoctl temperatures - Get current temperatures")
	fmt.Println()
	fmt.Println("Gets the current temperature readings from the Halko sensor unit.")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Printf("  %s [global-options] temperatures [options]\n", os.Args[0])
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -h, --help")
	fmt.Println("        Show this help message")
	fmt.Println()
	fmt.Println("Global Options:")
	fmt.Println("  -c, --config string")
	fmt.Println("        Path to the halko.cfg configuration file")
	fmt.Println("  -v, --verbose")
	fmt.Println("        Enable verbose output")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Printf("  %s temperatures\n", os.Args[0])
	fmt.Printf("  %s --config /path/to/halko.cfg temperatures\n", os.Args[0])
	fmt.Printf("  %s --verbose temperatures\n", os.Args[0])
	fmt.Println()
	fmt.Println("The temperatures will be retrieved from the sensor unit's GET /temperatures endpoint.")
}

func getTemperatures(sensorunitURL string, verbose bool) error {
	if verbose {
		fmt.Printf("Querying temperatures from: %s\n", sensorunitURL)
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Get temperatures endpoint from config
	var url string
	if globalConfig != nil && globalConfig.APIEndpoints != nil {
		url = globalConfig.APIEndpoints.SensorUnit.GetTemperaturesURL()
	} else {
		url = sensorunitURL + "/temperatures"
	}

	if verbose {
		fmt.Printf("GET %s\n", url)
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if verbose {
		fmt.Printf("HTTP Status: %d %s\n", resp.StatusCode, resp.Status)
		if len(respBody) > 0 {
			fmt.Printf("Raw Response: %s\n", string(respBody))
		}
		fmt.Println()
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		errorMsg := fmt.Sprintf("HTTP request failed with status %d", resp.StatusCode)
		if len(respBody) > 0 {
			errorMsg += ": " + strings.TrimSpace(string(respBody))
		}
		return fmt.Errorf("%s", errorMsg)
	}

	var response types.APIResponse[types.TemperatureResponse]
	if err := json.Unmarshal(respBody, &response); err != nil {
		return fmt.Errorf("failed to parse response JSON: %w", err)
	}

	fmt.Println("Current Temperatures:")
	if len(response.Data) == 0 {
		fmt.Println("  No temperature data available")
		return nil
	}

	// Display temperatures in a formatted way
	for name, temp := range response.Data {
		if temp == types.InvalidTemperatureReading {
			fmt.Printf("  %s: Invalid reading\n", formatTemperatureName(name))
		} else {
			fmt.Printf("  %s: %.2f°C\n", formatTemperatureName(name), temp)
		}
	}

	return nil
}

// formatTemperatureName converts internal temperature names to user-friendly names
func formatTemperatureName(name string) string {
	switch name {
	case "oven":
		return "Oven"
	case "material":
		return "Material/Wood"
	default:
		// Simple title case - capitalize first letter
		if len(name) == 0 {
			return name
		}
		return strings.ToUpper(string(name[0])) + strings.ToLower(name[1:])
	}
}

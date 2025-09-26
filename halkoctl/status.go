package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

func handleStatusCommand() {
	opts, err := ParseStatusOptions()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing arguments: %v\n", err)
		os.Exit(exitError)
	}

	if opts.Help {
		showStatusHelp()
		os.Exit(exitSuccess)
	}

	// Check status for each requested service
	for _, service := range opts.Services {
		switch service {
		case "executor":
			queryExecutorStatus(globalOpts.Verbose)
		case "sensorunit":
			querySensorUnitStatus(globalOpts.Verbose)
		default:
			fmt.Fprintf(os.Stderr, "Unknown service: %s\n", service)
		}

		// Add spacing between services if checking multiple
		if len(opts.Services) > 1 {
			fmt.Println()
		}
	}

	os.Exit(exitSuccess)
}

func showStatusHelp() {
	fmt.Println("halkoctl status - Get service status")
	fmt.Println()
	fmt.Println("Gets the status of Halko services. If no services are specified, checks all available services.")
	fmt.Println("Connection failures are reported as 'unavailable' status.")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Printf("  %s [global-options] status [service...]\n", os.Args[0])
	fmt.Println()
	fmt.Println("Arguments:")
	fmt.Println("  service")
	fmt.Println("        Service name to check (executor, sensorunit)")
	fmt.Println("        If no services specified, checks all available services")
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
	fmt.Printf("  %s status                    # Check all services\n", os.Args[0])
	fmt.Printf("  %s status executor           # Check executor service only\n", os.Args[0])
	fmt.Printf("  %s status sensorunit         # Check sensorunit service only\n", os.Args[0])
	fmt.Printf("  %s status executor sensorunit # Check both services\n", os.Args[0])
	fmt.Printf("  %s --verbose status          # Verbose output for all services\n", os.Args[0])
	fmt.Println()
}

func queryExecutorStatus(verbose bool) {
	executorURL := getExecutorAPIURL(globalConfig)
	if executorURL == "" {
		fmt.Printf("Executor Status: unavailable (could not determine URL from config)\n")
		return
	}

	if verbose {
		fmt.Printf("Querying executor status at: %s/engine/running\n", executorURL)
		fmt.Println()
	}

	getExecutorStatus(executorURL, verbose)
}

func getExecutorStatus(executorURL string, verbose bool) {
	if verbose {
		fmt.Printf("Querying status from: %s\n", executorURL)
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	url := executorURL + "/engine/running"

	if verbose {
		fmt.Printf("GET %s\n", url)
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("Executor Status: unavailable (failed to create HTTP request: %v)\n", err)
		return
	}

	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Executor Status: unavailable (connection failed)\n")
		if verbose {
			fmt.Printf("  Error: %v\n", err)
		}
		return
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Executor Status: unavailable (failed to read response: %v)\n", err)
		return
	}

	if verbose {
		fmt.Printf("HTTP Status: %d %s\n", resp.StatusCode, resp.Status)
		if len(respBody) > 0 {
			fmt.Printf("Raw Response: %s\n", string(respBody))
		}
		fmt.Println()
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		fmt.Printf("Executor Status: unavailable (HTTP %d)\n", resp.StatusCode)
		if verbose && len(respBody) > 0 {
			fmt.Printf("  Error: %s\n", strings.TrimSpace(string(respBody)))
		}
		return
	}

	var response map[string]interface{}
	if err := json.Unmarshal(respBody, &response); err != nil {
		fmt.Printf("Executor Status: unavailable (failed to parse response: %v)\n", err)
		return
	}

	data, ok := response["data"]
	if !ok {
		fmt.Printf("Executor Status: unavailable (no data field in response)\n")
		return
	}

	dataMap, ok := data.(map[string]interface{})
	if !ok {
		fmt.Printf("Executor Status: unavailable (invalid data format in response)\n")
		return
	}

	status, ok := dataMap["status"]
	if !ok {
		fmt.Printf("Executor Status: unavailable (no status field in response)\n")
		return
	}

	fmt.Printf("Executor Status: %v\n", status)

	displayProgramDetails(dataMap)
}

func displayProgramDetails(dataMap map[string]interface{}) {
	program, ok := dataMap["program"]
	if !ok {
		return
	}

	programMap, ok := program.(map[string]interface{})
	if !ok {
		return
	}

	fmt.Println()
	fmt.Println("Running Program Details:")

	if name, ok := programMap["name"]; ok {
		fmt.Printf("  Program Name: %v\n", name)
	}
	if currentPhase, ok := programMap["currentPhase"]; ok {
		fmt.Printf("  Current Phase: %v\n", currentPhase)
	}
	if elapsedTime, ok := programMap["elapsedTime"]; ok {
		fmt.Printf("  Elapsed Time: %v seconds\n", elapsedTime)
	}
	if currentTemp, ok := programMap["currentTemperature"]; ok {
		fmt.Printf("  Current Temperature: %v°C\n", currentTemp)
	}
	if targetTemp, ok := programMap["targetTemperature"]; ok {
		fmt.Printf("  Target Temperature: %v°C\n", targetTemp)
	}
	if remainingTime, ok := programMap["remainingTime"]; ok {
		fmt.Printf("  Remaining Time: %v seconds\n", remainingTime)
	}
}

func querySensorUnitStatus(verbose bool) {
	sensorUnitURL := getSensorUnitAPIURL(globalConfig)
	if sensorUnitURL == "" {
		fmt.Printf("SensorUnit Status: unavailable (could not determine URL from config)\n")
		return
	}

	if verbose {
		fmt.Printf("Querying sensorunit status at: %s/status\n", sensorUnitURL)
		fmt.Println()
	}

	getSensorUnitStatus(sensorUnitURL, verbose)
}

func getSensorUnitStatus(sensorUnitURL string, verbose bool) {
	if verbose {
		fmt.Printf("Querying status from: %s\n", sensorUnitURL)
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	url := sensorUnitURL + "/status"

	if verbose {
		fmt.Printf("GET %s\n", url)
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("SensorUnit Status: unavailable (failed to create HTTP request: %v)\n", err)
		return
	}

	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("SensorUnit Status: unavailable (connection failed)\n")
		if verbose {
			fmt.Printf("  Error: %v\n", err)
		}
		return
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("SensorUnit Status: unavailable (failed to read response: %v)\n", err)
		return
	}

	if verbose {
		fmt.Printf("HTTP Status: %d %s\n", resp.StatusCode, resp.Status)
		if len(respBody) > 0 {
			fmt.Printf("Raw Response: %s\n", string(respBody))
		}
		fmt.Println()
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		fmt.Printf("SensorUnit Status: unavailable (HTTP %d)\n", resp.StatusCode)
		if verbose && len(respBody) > 0 {
			fmt.Printf("  Error: %s\n", strings.TrimSpace(string(respBody)))
		}
		return
	}

	var response map[string]interface{}
	if err := json.Unmarshal(respBody, &response); err != nil {
		fmt.Printf("SensorUnit Status: unavailable (failed to parse response: %v)\n", err)
		return
	}

	data, ok := response["data"]
	if !ok {
		fmt.Printf("SensorUnit Status: unavailable (no data field in response)\n")
		return
	}

	dataMap, ok := data.(map[string]interface{})
	if !ok {
		fmt.Printf("SensorUnit Status: unavailable (invalid data format in response)\n")
		return
	}

	status, ok := dataMap["status"]
	if !ok {
		fmt.Printf("SensorUnit Status: unavailable (no status field in response)\n")
		return
	}

	fmt.Printf("SensorUnit Status: %v\n", status)
}

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
		case "powerunit":
			queryPowerUnitStatus(globalOpts.Verbose)
		case "storage":
			queryStorageStatus(globalOpts.Verbose)
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
	fmt.Println("        Service name to check (executor, sensorunit, powerunit, storage)")
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
	fmt.Printf("  %s status powerunit storage  # Check powerunit and storage services\n", os.Args[0])
	fmt.Printf("  %s --verbose status          # Verbose output for all services\n", os.Args[0])
	fmt.Println()
}

func queryExecutorStatus(verbose bool) {
	queryServiceStatus("Executor", &globalConfig.APIEndpoints.Executor, verbose)
}

func querySensorUnitStatus(verbose bool) {
	queryServiceStatus("SensorUnit", &globalConfig.APIEndpoints.SensorUnit, verbose)
}

func queryPowerUnitStatus(verbose bool) {
	queryServiceStatus("PowerUnit", &globalConfig.APIEndpoints.PowerUnit, verbose)
}

func queryStorageStatus(verbose bool) {
	queryServiceStatus("Storage", &globalConfig.APIEndpoints.Storage, verbose)
}

func queryServiceStatus(serviceName string, endpoint types.EndpointWithStatus, verbose bool) {
	url := endpoint.GetStatusURL()

	if verbose {
		fmt.Printf("Querying %s status at: %s\n", serviceName, url)
		fmt.Println()
	}

	getServiceStatus(serviceName, url, verbose)
}

func getServiceStatus(serviceName string, url string, verbose bool) {
	if verbose {
		fmt.Printf("Querying status from: %s\n", url)
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	if verbose {
		fmt.Printf("GET %s\n", url)
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("%s Status: unavailable (failed to create HTTP request: %v)\n", serviceName, err)
		return
	}

	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("%s Status: unavailable (connection failed)\n", serviceName)
		if verbose {
			fmt.Printf("  Error: %v\n", err)
		}
		return
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("%s Status: unavailable (failed to read response: %v)\n", serviceName, err)
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
		fmt.Printf("%s Status: unavailable (HTTP %d)\n", serviceName, resp.StatusCode)
		if verbose && len(respBody) > 0 {
			fmt.Printf("  Error: %s\n", strings.TrimSpace(string(respBody)))
		}
		return
	}

	// Only try to parse JSON if we have a body
	if len(respBody) == 0 {
		fmt.Printf("%s Status: unavailable (empty response body)\n", serviceName)
		return
	}

	var response map[string]interface{}
	if err := json.Unmarshal(respBody, &response); err != nil {
		fmt.Printf("%s Status: unavailable (failed to parse response: %v)\n", serviceName, err)
		return
	}

	data, ok := response["data"]
	if !ok {
		fmt.Printf("%s Status: unavailable (no data field in response)\n", serviceName)
		return
	}

	dataMap, ok := data.(map[string]interface{})
	if !ok {
		fmt.Printf("%s Status: unavailable (invalid data format in response)\n", serviceName)
		return
	}

	status, ok := dataMap["status"]
	if !ok {
		fmt.Printf("%s Status: unavailable (no status field in response)\n", serviceName)
		return
	}

	fmt.Printf("%s Status: %v\n", serviceName, status)

	// Display details if in verbose mode
	if verbose {
		if details, ok := dataMap["details"]; ok {
			if detailsMap, ok := details.(map[string]interface{}); ok && len(detailsMap) > 0 {
				fmt.Println("  Details:")
				for key, value := range detailsMap {
					fmt.Printf("    %s: %v\n", key, value)
				}
			}
		}
	}
}

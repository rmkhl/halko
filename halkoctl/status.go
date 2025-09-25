package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

func handleStatusCommand() {
	// Create a new FlagSet for the status command
	statusFlags := flag.NewFlagSet("status", flag.ExitOnError)

	var (
		host    = statusFlags.String("host", "localhost", "Executor host (default: localhost)")
		port    = statusFlags.String("port", "8080", "Executor port (default: 8080)")
		verbose = statusFlags.Bool("verbose", false, "Enable verbose output")
		help    = statusFlags.Bool("help", false, "Show help for status command")
	)

	// Parse the arguments starting from os.Args[2] (after "status")
	if err := statusFlags.Parse(os.Args[2:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		os.Exit(exitError)
	}

	if *help {
		showStatusHelp()
		os.Exit(exitSuccess)
	}

	if *verbose {
		fmt.Printf("Querying executor status at: http://%s:%s/engine/api/v1/running\n", *host, *port)
		fmt.Println()
	}

	// Get the status
	err := getStatus(*host, *port, *verbose)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get status: %v\n", err)
		os.Exit(exitError)
	}

	os.Exit(exitSuccess)
}

func showStatusHelp() {
	fmt.Println("halkoctl status - Get program status")
	fmt.Println()
	fmt.Println("Gets the status of the currently running program from the Halko executor.")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Printf("  %s status [options]\n", os.Args[0])
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -host string")
	fmt.Println("        Executor host (default: localhost)")
	fmt.Println("  -port string")
	fmt.Println("        Executor port (default: 8080)")
	fmt.Println("  -verbose")
	fmt.Println("        Enable verbose output")
	fmt.Println("  -help")
	fmt.Println("        Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Printf("  %s status\n", os.Args[0])
	fmt.Printf("  %s status -host 192.168.1.100 -port 8080 -verbose\n", os.Args[0])
	fmt.Println()
	fmt.Println("The status will be retrieved from the executor's GET /engine/api/v1/running endpoint.")
}

func getStatus(host, port string, verbose bool) error {
	if verbose {
		fmt.Printf("Querying status from: %s:%s\n", host, port)
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Construct the URL
	url := fmt.Sprintf("http://%s:%s/engine/api/v1/running", host, port)

	if verbose {
		fmt.Printf("GET %s\n", url)
	}

	// Create the HTTP request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers for consistency
	req.Header.Set("Accept", "application/json")

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
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

	// Check if the request was successful
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		errorMsg := fmt.Sprintf("HTTP request failed with status %d", resp.StatusCode)
		if len(respBody) > 0 {
			errorMsg += ": " + strings.TrimSpace(string(respBody))
		}
		return fmt.Errorf("%s", errorMsg)
	}

	// Parse and display the response
	var response map[string]interface{}
	if err := json.Unmarshal(respBody, &response); err != nil {
		return fmt.Errorf("failed to parse response JSON: %w", err)
	}

	// Extract the data field
	data, ok := response["data"]
	if !ok {
		fmt.Println("No data field in response")
		return nil
	}

	dataMap, ok := data.(map[string]interface{})
	if !ok {
		fmt.Println("Invalid data format in response")
		return nil
	}

	// Get the status
	status, ok := dataMap["status"]
	if !ok {
		fmt.Println("No status field in response")
		return nil
	}

	fmt.Printf("Executor Status: %v\n", status)

	// If there's a program running, show its details
	displayProgramDetails(dataMap)

	return nil
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

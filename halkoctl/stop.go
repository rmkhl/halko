package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/rmkhl/halko/types"
)

func handleStopCommand() {
	if len(os.Args) > 2 {
		arg := os.Args[2]
		if arg == "-h" || arg == "--help" {
			showStopHelp()
			os.Exit(exitSuccess)
		}
	}

	stopRunningProgram()
	os.Exit(exitSuccess)
}

func showStopHelp() {
	fmt.Println("halkoctl stop - Stop currently running program")
	fmt.Println()
	fmt.Println("Stops the program currently executing in the controlunit.")
	fmt.Println("The program will be canceled and marked as stopped in the execution history.")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Printf("  %s [global-options] stop\n", os.Args[0])
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -h, --help")
	fmt.Println("        Show this help message")
	fmt.Println()
	fmt.Println("Global Options:")
	fmt.Println("  -c, --config string")
	fmt.Println("        Path to the halko.cfg configuration file")
	fmt.Println("  -v, --verbose")
	fmt.Println("        Enable verbose output for HTTP requests")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Printf("  %s stop                       # Stop current program\n", os.Args[0])
	fmt.Printf("  %s --verbose stop             # Stop with verbose HTTP output\n", os.Args[0])
	fmt.Println()
}

func stopRunningProgram() {
	controlunitURL := globalConfig.APIEndpoints.ControlUnit.URL
	url := controlunitURL + "/engine/running"

	if globalOpts.Verbose {
		fmt.Printf("Stopping program at: %s\n", url)
		fmt.Println()
	}

	fmt.Println("Stopping program... (this may take a minute for long-running programs)")

	// Use longer timeout for stop operation as it involves:
	// - Stopping FSM engine
	// - Saving execution state
	// - Moving potentially large log files to history
	client := &http.Client{
		Timeout: 120 * time.Second,
	}

	if globalOpts.Verbose {
		fmt.Printf("DELETE %s\n", url)
	}

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating HTTP request: %v\n", err)
		os.Exit(exitError)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error connecting to controlunit: %v\n", err)
		os.Exit(exitError)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading response: %v\n", err)
		os.Exit(exitError)
	}

	if globalOpts.Verbose {
		fmt.Printf("HTTP Status: %d %s\n", resp.StatusCode, resp.Status)
		if len(respBody) > 0 {
			fmt.Printf("Raw Response: %s\n", string(respBody))
		}
		fmt.Println()
	}

	// Handle 404 Not Found as "no program running"
	if resp.StatusCode == http.StatusNotFound {
		fmt.Println("No program currently running")
		return
	}

	if resp.StatusCode != http.StatusOK {
		// Try to parse error response
		var errorResponse types.APIErrorResponse
		if err := json.Unmarshal(respBody, &errorResponse); err == nil && errorResponse.Err != "" {
			fmt.Fprintf(os.Stderr, "Error: %s\n", errorResponse.Err)
		} else {
			fmt.Fprintf(os.Stderr, "Error: HTTP %d - %s\n", resp.StatusCode, string(respBody))
		}
		os.Exit(exitError)
	}

	fmt.Println("✓ Program stopped successfully")
}

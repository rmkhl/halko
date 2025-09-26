package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/rmkhl/halko/types"
)

func handleDisplayCommand() {
	opts, err := ParseDisplayOptions()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		os.Exit(exitError)
	}

	if opts.Help {
		showDisplayHelp()
		os.Exit(exitSuccess)
	}

	if opts.Message == "" {
		fmt.Fprintf(os.Stderr, "Error: message text is required\n\n")
		showDisplayHelp()
		os.Exit(exitError)
	}

	url := getSensorunitAPIURL(globalConfig)
	if url == "" {
		fmt.Fprintf(os.Stderr, "Error: Could not determine sensorunit URL from config\n")
		os.Exit(exitError)
	}

	if globalOpts.Verbose {
		fmt.Printf("Sending display message: %s\n", opts.Message)
		fmt.Printf("Sensorunit endpoint: %s\n", url)
		fmt.Println()
	}

	err = sendDisplayMessage(opts.Message, url, globalOpts.Verbose)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to send display message: %v\n", err)
		os.Exit(exitError)
	}

	fmt.Println("✓ Display message sent successfully!")
	os.Exit(exitSuccess)
}

func showDisplayHelp() {
	fmt.Println("halkoctl display - Send text to sensor unit display")
	fmt.Println()
	fmt.Println("Sends a text message to the sensor unit to be displayed on its LCD.")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Printf("  %s [global-options] display <message> [options]\n", os.Args[0])
	fmt.Println()
	fmt.Println("Arguments:")
	fmt.Println("  message           Text message to display on the sensor unit LCD (required)")
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
	fmt.Printf("  %s display \"Hello World\"\n", os.Args[0])
	fmt.Printf("  %s --config /path/to/halko.cfg display \"Temperature: 25°C\"\n", os.Args[0])
	fmt.Printf("  %s --verbose display \"System Ready\"\n", os.Args[0])
	fmt.Println()
	fmt.Println("The message will be sent to the sensor unit's POST /display endpoint")
	fmt.Println("to update the LCD display text.")
}

func sendDisplayMessage(message, sensorunitURL string, verbose bool) error {
	if verbose {
		fmt.Printf("Preparing display message: %s\n", message)
	}

	payload := types.DisplayRequest{
		Message: message,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON payload: %w", err)
	}

	if verbose {
		fmt.Printf("JSON payload: %s\n", string(jsonData))
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Construct the full URL using the display endpoint
	displayEndpoint := "/display"
	if globalConfig != nil && globalConfig.APIEndpoints != nil && globalConfig.APIEndpoints.Display != "" {
		displayEndpoint = globalConfig.APIEndpoints.Display
	}

	url := sensorunitURL + displayEndpoint

	if verbose {
		fmt.Printf("POST %s\n", url)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if verbose {
		fmt.Printf("Response status: %d\n", resp.StatusCode)
		fmt.Printf("Response body: %s\n", string(respBody))
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned error status %d: %s", resp.StatusCode, string(respBody))
	}

	if verbose {
		fmt.Println("✓ Display message sent successfully")
	}

	return nil
}

// getSensorunitAPIURL returns the base API URL for sensorunit
func getSensorunitAPIURL(config *types.HalkoConfig) string {
	if config == nil {
		return "http://localhost:8081"
	}

	url, err := config.GetSensorUnitUrl()
	if err != nil {
		// Fallback to default if there's an error
		return "http://localhost:8081"
	}

	return url
}

package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/rmkhl/halko/types"
)

func handleSendCommand() {
	// Create a new FlagSet for the send command
	sendFlags := flag.NewFlagSet("send", flag.ExitOnError)

	var (
		verbose = sendFlags.Bool("v", false, "Enable verbose output")
		help    = sendFlags.Bool("h", false, "Show help for send command")
	)

	// Add long options
	sendFlags.BoolVar(verbose, "verbose", false, "Enable verbose output")
	sendFlags.BoolVar(help, "help", false, "Show help for send command")

	// Parse the arguments starting from os.Args[2] (after "send")
	if err := sendFlags.Parse(os.Args[2:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		os.Exit(exitError)
	}

	if *help {
		showSendHelp()
		os.Exit(exitSuccess)
	}

	// Get the program path from remaining arguments
	args := sendFlags.Args()
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Error: program file path is required\n\n")
		showSendHelp()
		os.Exit(exitError)
	}

	programPath := args[0]

	// Get the executor URL from config
	url := getExecutorAPIURL(globalConfig)

	if *verbose {
		fmt.Printf("Sending program: %s\n", programPath)
		fmt.Printf("Executor endpoint: %s\n", url)
		fmt.Println()
	}

	// Send the program
	err := sendProgram(programPath, url, *verbose)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to send program: %v\n", err)
		os.Exit(exitError)
	}

	fmt.Println("✓ Program sent successfully!")
	os.Exit(exitSuccess)
}

func showSendHelp() {
	fmt.Println("halkoctl send - Send program to executor")
	fmt.Println()
	fmt.Println("Sends a program.json file to the Halko executor to start execution.")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Printf("  %s send <program-file> [options]\n", os.Args[0])
	fmt.Println()
	fmt.Println("Arguments:")
	fmt.Println("  program-file      Path to the program.json file to send (required)")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -v, --verbose")
	fmt.Println("        Enable verbose output")
	fmt.Println("  -h, --help")
	fmt.Println("        Show this help message")
	fmt.Println()
	fmt.Println("Global Options:")
	fmt.Println("  -c, --config string")
	fmt.Println("        Path to the halko.cfg configuration file")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Printf("  %s send example/example-program-delta.json\n", os.Args[0])
	fmt.Printf("  %s --config /path/to/halko.cfg send my-program.json -v\n", os.Args[0])
	fmt.Printf("  %s send my-program.json --verbose\n", os.Args[0])
	fmt.Println()
	fmt.Println("The program will be sent to the executor's POST /engine/api/v1/running endpoint")
	fmt.Println("to start immediate execution. The executor will validate the program.")
}

func sendProgram(programPath, executorURL string, verbose bool) error {
	// Check if file exists
	if _, err := os.Stat(programPath); os.IsNotExist(err) {
		return fmt.Errorf("program file does not exist: %s", programPath)
	}

	// Load the program file
	if verbose {
		fmt.Printf("Loading program from: %s\n", programPath)
	}

	data, err := os.ReadFile(programPath)
	if err != nil {
		return fmt.Errorf("failed to read program file: %w", err)
	}

	if verbose {
		fmt.Println("✓ Program file loaded successfully")
	}

	// Parse the JSON to validate it's properly formatted
	if verbose {
		fmt.Println("Validating JSON format...")
	}

	var program types.Program
	err = json.Unmarshal(data, &program)
	if err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	if verbose {
		fmt.Printf("✓ JSON parsed successfully - Program: '%s' with %d steps\n",
			program.ProgramName, len(program.ProgramSteps))
	}

	// Prepare the request body
	// According to the API, we can send the program directly in a "program" field
	requestBody := map[string]interface{}{
		"program": program,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	if verbose {
		fmt.Println("Sending HTTP request...")
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Construct the URL
	url := executorURL + "/engine/api/v1/running"

	// Create the HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

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
			fmt.Printf("Response: %s\n", string(respBody))
		}
	}

	// Check if the request was successful
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		errorMsg := fmt.Sprintf("HTTP request failed with status %d", resp.StatusCode)
		if len(respBody) > 0 {
			errorMsg += ": " + strings.TrimSpace(string(respBody))
		}
		return fmt.Errorf("%s", errorMsg)
	}

	if verbose {
		fmt.Println("✓ Program sent and accepted by executor")
		displaySendResponse(respBody)
	}

	return nil
}

func displaySendResponse(respBody []byte) {
	var response map[string]interface{}
	if err := json.Unmarshal(respBody, &response); err != nil {
		return
	}

	data, ok := response["data"]
	if !ok {
		return
	}

	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return
	}

	if status, ok := dataMap["status"]; ok {
		fmt.Printf("Executor status: %v\n", status)
	}

	if program, ok := dataMap["program"]; ok {
		if programMap, ok := program.(map[string]interface{}); ok {
			if name, ok := programMap["name"]; ok {
				fmt.Printf("Started program: %v\n", name)
			}
		}
	}
}

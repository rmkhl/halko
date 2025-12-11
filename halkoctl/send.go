package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/rmkhl/halko/types"
)

func handleSendCommand() {
	opts, err := ParseSendOptions()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		os.Exit(exitError)
	}

	if opts.Help {
		showSendHelp()
		os.Exit(exitSuccess)
	}

	if opts.ProgramPath == "" {
		fmt.Fprintf(os.Stderr, "Error: program file path is required\n\n")
		showSendHelp()
		os.Exit(exitError)
	}

	url := getControlUnitAPIURL(globalConfig)

	if globalOpts.Verbose {
		fmt.Printf("Sending program: %s\n", opts.ProgramPath)
		fmt.Printf("ControlUnit endpoint: %s\n", url)
		fmt.Println()
	}

	err = sendProgram(opts.ProgramPath, url, globalOpts.Verbose)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to send program: %v\n", err)
		os.Exit(exitError)
	}

	fmt.Println("✓ Program sent successfully!")
	os.Exit(exitSuccess)
}

func showSendHelp() {
	fmt.Println("halkoctl send - Send program to controlunit")
	fmt.Println()
	fmt.Println("Sends a program.json file to the Halko controlunit to start execution.")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Printf("  %s [global-options] send <program-file> [options]\n", os.Args[0])
	fmt.Println()
	fmt.Println("Arguments:")
	fmt.Println("  program-file      Path to the program.json file to send (required)")
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
	fmt.Printf("  %s send example/example-program-delta.json\n", os.Args[0])
	fmt.Printf("  %s --config /path/to/halko.cfg send my-program.json\n", os.Args[0])
	fmt.Printf("  %s --verbose send my-program.json\n", os.Args[0])
	fmt.Println()
	fmt.Println("The program will be sent to the controlunit's POST /engine/running endpoint")
	fmt.Println("to start immediate execution. The controlunit will validate the program.")
}

func sendProgram(programPath, controlunitURL string, verbose bool) error {
	if _, err := os.Stat(programPath); os.IsNotExist(err) {
		return fmt.Errorf("program file does not exist: %s", programPath)
	}

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

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Construct the full URL using the running endpoint
	var url string
	if globalConfig != nil && globalConfig.APIEndpoints != nil {
		url = globalConfig.APIEndpoints.ControlUnit.GetEngineURL() + "/running"
	} else {
		url = controlunitURL + "/engine/running"
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

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
			fmt.Printf("Response: %s\n", string(respBody))
		}
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		errorMsg := fmt.Sprintf("HTTP request failed with status %d", resp.StatusCode)
		if len(respBody) > 0 {
			errorMsg += ": " + strings.TrimSpace(string(respBody))
		}
		return fmt.Errorf("%s", errorMsg)
	}

	if verbose {
		fmt.Println("✓ Program sent and accepted by controlunit")
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
		fmt.Printf("ControlUnit status: %v\n", status)
	}

	if program, ok := dataMap["program"]; ok {
		if programMap, ok := program.(map[string]interface{}); ok {
			if name, ok := programMap["name"]; ok {
				fmt.Printf("Started program: %v\n", name)
			}
		}
	}
}

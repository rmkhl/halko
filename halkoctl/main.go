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

const (
	exitSuccess = 0
	exitError   = 1
)

func main() {
	if len(os.Args) < 2 {
		showHelp()
		os.Exit(exitError)
	}

	command := os.Args[1]
	switch command {
	case "send":
		handleSendCommand()
	case "help", "-help", "--help":
		showHelp()
		os.Exit(exitSuccess)
	default:
		fmt.Fprintf(os.Stderr, "Error: Unknown command '%s'\n\n", command)
		showHelp()
		os.Exit(exitError)
	}
}

func handleSendCommand() {
	// Create a new FlagSet for the send command
	sendFlags := flag.NewFlagSet("send", flag.ExitOnError)

	var (
		programPath = sendFlags.String("program", "", "Path to the program.json file to send (required)")
		host        = sendFlags.String("host", "localhost", "Executor host (default: localhost)")
		port        = sendFlags.String("port", "8080", "Executor port (default: 8080)")
		verbose     = sendFlags.Bool("verbose", false, "Enable verbose output")
		help        = sendFlags.Bool("help", false, "Show help for send command")
	)

	// Parse the arguments starting from os.Args[2] (after "send")
	sendFlags.Parse(os.Args[2:])

	if *help {
		showSendHelp()
		os.Exit(exitSuccess)
	}

	if *programPath == "" {
		fmt.Fprintf(os.Stderr, "Error: -program flag is required\n\n")
		showSendHelp()
		os.Exit(exitError)
	}

	if *verbose {
		fmt.Printf("Sending program: %s\n", *programPath)
		fmt.Printf("Executor endpoint: http://%s:%s/engine/api/v1/running\n", *host, *port)
		fmt.Println()
	}

	// Send the program
	err := sendProgram(*programPath, *host, *port, *verbose)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to send program: %v\n", err)
		os.Exit(exitError)
	}

	fmt.Println("✓ Program sent successfully!")
	os.Exit(exitSuccess)
}

func showHelp() {
	fmt.Println("halkoctl - Halko Control Tool")
	fmt.Println()
	fmt.Println("A command-line tool for interacting with the Halko wood drying kiln control system.")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Printf("  %s <command> [options]\n", os.Args[0])
	fmt.Println()
	fmt.Println("Available Commands:")
	fmt.Println("  send        Send a program.json file to the executor to start execution")
	fmt.Println("  help        Show this help message")
	fmt.Println()
	fmt.Println("Use \"halkoctl <command> -help\" for more information about a command.")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Printf("  %s send -program example/example-program-delta.json\n", os.Args[0])
	fmt.Printf("  %s send -program my-program.json -host 192.168.1.100 -port 8080 -verbose\n", os.Args[0])
}

func showSendHelp() {
	fmt.Println("halkoctl send - Send program to executor")
	fmt.Println()
	fmt.Println("Sends a program.json file to the Halko executor to start execution.")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Printf("  %s send -program <path-to-program.json> [options]\n", os.Args[0])
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -program string")
	fmt.Println("        Path to the program.json file to send (required)")
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
	fmt.Printf("  %s send -program example/example-program-delta.json\n", os.Args[0])
	fmt.Printf("  %s send -program my-program.json -host 192.168.1.100 -port 8080 -verbose\n", os.Args[0])
	fmt.Println()
	fmt.Println("The program will be sent to the executor's POST /engine/api/v1/running endpoint")
	fmt.Println("to start immediate execution. The executor will validate the program.")
}

func sendProgram(programPath, host, port string, verbose bool) error {
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
	url := fmt.Sprintf("http://%s:%s/engine/api/v1/running", host, port)

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
			errorMsg += fmt.Sprintf(": %s", strings.TrimSpace(string(respBody)))
		}
		return fmt.Errorf("%s", errorMsg)
	}

	if verbose {
		fmt.Println("✓ Program sent and accepted by executor")

		// Try to parse and display the response
		var response map[string]interface{}
		if err := json.Unmarshal(respBody, &response); err == nil {
			if data, ok := response["data"]; ok {
				if dataMap, ok := data.(map[string]interface{}); ok {
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
			}
		}
	}

	return nil
}

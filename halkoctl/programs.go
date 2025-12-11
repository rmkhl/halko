package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/rmkhl/halko/types"
)

func handleProgramsCommand() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Error: programs command requires a subcommand\n\n")
		showProgramsHelp()
		os.Exit(exitError)
	}

	subcommand := os.Args[2]

	// Check for help flag
	for _, arg := range os.Args[3:] {
		if arg == "-h" || arg == "--help" {
			showProgramsHelp()
			os.Exit(exitSuccess)
		}
	}

	switch subcommand {
	case "list":
		handleProgramListCommand()
	case "get":
		handleProgramGetCommand()
	case "create":
		handleProgramCreateCommand()
	case "update":
		handleProgramUpdateCommand()
	case "delete":
		handleProgramDeleteCommand()
	default:
		fmt.Fprintf(os.Stderr, "Error: Unknown programs subcommand '%s'\n\n", subcommand)
		showProgramsHelp()
		os.Exit(exitError)
	}
}

func handleProgramListCommand() {
	baseURL := getStorageAPIURL(globalConfig)
	url := baseURL + globalConfig.APIEndpoints.ControlUnit.Programs

	if globalOpts.Verbose {
		fmt.Printf("Creating program at: %s\n", url)
		fmt.Println()
	}

	resp, err := http.Get(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to storage service: %v\n", err)
		os.Exit(exitError)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		resp.Body.Close()
		fmt.Fprintf(os.Stderr, "Failed to read response: %v\n", err)
		os.Exit(exitError)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		var errorResp types.APIErrorResponse
		if err := json.Unmarshal(body, &errorResp); err == nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", errorResp.Err)
		} else {
			fmt.Fprintf(os.Stderr, "Error: HTTP %d\n", resp.StatusCode)
		}
		os.Exit(exitError)
	}

	var response types.APIResponse[[]types.StoredProgramInfo]
	if err := json.Unmarshal(body, &response); err != nil {
		resp.Body.Close()
		fmt.Fprintf(os.Stderr, "Failed to parse response: %v\n", err)
		os.Exit(exitError)
	}

	resp.Body.Close()

	if len(response.Data) == 0 {
		fmt.Println("No programs found")
	} else {
		fmt.Println("Stored programs:")
		for _, programInfo := range response.Data {
			fmt.Printf("  %s (last modified: %s)\n", programInfo.Name, programInfo.LastModified)
		}
	}
}

func handleProgramGetCommand() {
	if len(os.Args) < 4 {
		fmt.Fprintf(os.Stderr, "Error: program name is required\n")
		fmt.Fprintf(os.Stderr, "Usage: %s programs get <program-name>\n", os.Args[0])
		os.Exit(exitError)
	}

	programName := os.Args[3]
	baseURL := getStorageAPIURL(globalConfig)
	url := baseURL + globalConfig.APIEndpoints.ControlUnit.Programs + "/" + programName

	if globalOpts.Verbose {
		fmt.Printf("Getting program '%s' from: %s\n", programName, url)
		fmt.Println()
	}

	resp, err := http.Get(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to storage service: %v\n", err)
		os.Exit(exitError)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		resp.Body.Close()
		fmt.Fprintf(os.Stderr, "Failed to read response: %v\n", err)
		os.Exit(exitError)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		var errorResp types.APIErrorResponse
		if err := json.Unmarshal(body, &errorResp); err == nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", errorResp.Err)
		} else {
			fmt.Fprintf(os.Stderr, "Error: HTTP %d\n", resp.StatusCode)
		}
		os.Exit(exitError)
	}

	var response types.APIResponse[types.Program]
	if err := json.Unmarshal(body, &response); err != nil {
		resp.Body.Close()
		fmt.Fprintf(os.Stderr, "Failed to parse response: %v\n", err)
		os.Exit(exitError)
	}

	// Pretty print the program as JSON
	prettyJSON, err := json.MarshalIndent(response.Data, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to format program: %v\n", err)
		os.Exit(exitError)
	}

	fmt.Println(string(prettyJSON))
}

func handleProgramCreateCommand() {
	if len(os.Args) < 4 {
		fmt.Fprintf(os.Stderr, "Error: program file path is required\n")
		fmt.Fprintf(os.Stderr, "Usage: %s programs create <program-file>\n", os.Args[0])
		os.Exit(exitError)
	}

	programPath := os.Args[3]

	// Read and parse the program file
	program, err := loadProgramFromFile(programPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load program file: %v\n", err)
		os.Exit(exitError)
	}

	baseURL := getStorageAPIURL(globalConfig)
	url := baseURL + globalConfig.APIEndpoints.ControlUnit.Programs

	if globalOpts.Verbose {
		fmt.Printf("Creating program '%s' from file: %s\n", program.ProgramName, programPath)
		fmt.Printf("Storage endpoint: %s\n", url)
		fmt.Println()
	}

	// Marshal program to JSON
	programJSON, err := json.Marshal(program)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to marshal program: %v\n", err)
		os.Exit(exitError)
	}

	// Make POST request
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(programJSON))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to storage service: %v\n", err)
		os.Exit(exitError)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		resp.Body.Close()
		fmt.Fprintf(os.Stderr, "Failed to read response: %v\n", err)
		os.Exit(exitError)
	}

	if resp.StatusCode != http.StatusCreated {
		resp.Body.Close()
		var errorResp types.APIErrorResponse
		if err := json.Unmarshal(body, &errorResp); err == nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", errorResp.Err)
		} else {
			fmt.Fprintf(os.Stderr, "Error: HTTP %d\n", resp.StatusCode)
		}
		os.Exit(exitError)
	}

	fmt.Printf("✓ Program '%s' created successfully!\n", program.ProgramName)
}

func handleProgramUpdateCommand() {
	if len(os.Args) < 5 {
		fmt.Fprintf(os.Stderr, "Error: program name and file path are required\n")
		fmt.Fprintf(os.Stderr, "Usage: %s programs update <program-name> <program-file>\n", os.Args[0])
		os.Exit(exitError)
	}

	programName := os.Args[3]
	programPath := os.Args[4]

	// Read and parse the program file
	program, err := loadProgramFromFile(programPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load program file: %v\n", err)
		os.Exit(exitError)
	}

	// Override the program name with the one specified in the command
	program.ProgramName = programName

	baseURL := getStorageAPIURL(globalConfig)
	url := baseURL + globalConfig.APIEndpoints.ControlUnit.Programs + "/" + programName

	if globalOpts.Verbose {
		fmt.Printf("Updating program '%s' from file: %s\n", programName, programPath)
		fmt.Printf("Storage endpoint: %s\n", url)
		fmt.Println()
	}

	// Marshal program to JSON
	programJSON, err := json.Marshal(program)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to marshal program: %v\n", err)
		os.Exit(exitError)
	}

	// Make POST request (the API uses POST for updates)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(programJSON))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to storage service: %v\n", err)
		os.Exit(exitError)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		resp.Body.Close()
		fmt.Fprintf(os.Stderr, "Failed to read response: %v\n", err)
		os.Exit(exitError)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		var errorResp types.APIErrorResponse
		if err := json.Unmarshal(body, &errorResp); err == nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", errorResp.Err)
		} else {
			fmt.Fprintf(os.Stderr, "Error: HTTP %d\n", resp.StatusCode)
		}
		os.Exit(exitError)
	}

	fmt.Printf("✓ Program '%s' updated successfully!\n", programName)
}

func handleProgramDeleteCommand() {
	if len(os.Args) < 4 {
		fmt.Fprintf(os.Stderr, "Error: program name is required\n")
		fmt.Fprintf(os.Stderr, "Usage: %s programs delete <program-name>\n", os.Args[0])
		os.Exit(exitError)
	}

	programName := os.Args[3]
	baseURL := getStorageAPIURL(globalConfig)
	url := baseURL + globalConfig.APIEndpoints.ControlUnit.Programs + "/" + programName

	if globalOpts.Verbose {
		fmt.Printf("Deleting program '%s' from: %s\n", programName, url)
		fmt.Println()
	}

	// Create DELETE request
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create request: %v\n", err)
		os.Exit(exitError)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to storage service: %v\n", err)
		os.Exit(exitError)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		resp.Body.Close()
		fmt.Fprintf(os.Stderr, "Failed to read response: %v\n", err)
		os.Exit(exitError)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		var errorResp types.APIErrorResponse
		if err := json.Unmarshal(body, &errorResp); err == nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", errorResp.Err)
		} else {
			fmt.Fprintf(os.Stderr, "Error: HTTP %d\n", resp.StatusCode)
		}
		os.Exit(exitError)
	}

	fmt.Printf("✓ Program '%s' deleted successfully!\n", programName)
}

func loadProgramFromFile(filename string) (*types.Program, error) {
	if !strings.HasSuffix(filename, ".json") && !isJSONFile(filename) {
		return nil, errors.New("file must be a JSON file")
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("unable to read file: %w", err)
	}

	var program types.Program
	if err := json.Unmarshal(data, &program); err != nil {
		return nil, fmt.Errorf("unable to parse JSON: %w", err)
	}

	if program.ProgramName == "" {
		// Try to use filename without extension as program name
		baseName := filepath.Base(filename)
		if ext := filepath.Ext(baseName); ext != "" {
			baseName = strings.TrimSuffix(baseName, ext)
		}
		program.ProgramName = baseName
	}

	return &program, nil
}

func isJSONFile(filename string) bool {
	file, err := os.Open(filename)
	if err != nil {
		return false
	}
	defer file.Close()

	// Read first few bytes to check if it looks like JSON
	buf := make([]byte, 64)
	n, err := file.Read(buf)
	if err != nil {
		return false
	}

	content := strings.TrimSpace(string(buf[:n]))
	return strings.HasPrefix(content, "{") || strings.HasPrefix(content, "[")
}

func showProgramsHelp() {
	fmt.Println("halkoctl programs - Manage Stored Programs")
	fmt.Println()
	fmt.Println("Manage programs stored in the Halko storage service.")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Printf("  %s programs <subcommand> [arguments]\n", os.Args[0])
	fmt.Println()
	fmt.Println("Available Subcommands:")
	fmt.Println("  list                     List all stored programs")
	fmt.Println("  get <program-name>       Get a specific program")
	fmt.Println("  create <program-file>    Create a new program from JSON file")
	fmt.Println("  update <name> <file>     Update existing program with new content")
	fmt.Println("  delete <program-name>    Delete a stored program")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Printf("  %s programs list\n", os.Args[0])
	fmt.Printf("  %s programs get my-program\n", os.Args[0])
	fmt.Printf("  %s programs create example/example-program-delta.json\n", os.Args[0])
	fmt.Printf("  %s programs update my-program updated-program.json\n", os.Args[0])
	fmt.Printf("  %s programs delete old-program\n", os.Args[0])
	fmt.Println()
	fmt.Println("Notes:")
	fmt.Println("  - Program files must be valid JSON")
	fmt.Println("  - Program names are derived from filenames if not specified in JSON")
	fmt.Println("  - Use --verbose for detailed operation information")
}

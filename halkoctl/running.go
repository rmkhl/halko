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

func handleRunningCommand() {
	if len(os.Args) > 2 {
		arg := os.Args[2]
		if arg == "-h" || arg == "--help" {
			showRunningHelp()
			os.Exit(exitSuccess)
		}
	}

	queryRunningProgram()
	os.Exit(exitSuccess)
}

func showRunningHelp() {
	fmt.Println("halkoctl running - Show currently running program")
	fmt.Println()
	fmt.Println("Displays information about the program currently executing in the controlunit.")
	fmt.Println("Shows program name, current phase, elapsed time, and temperature status.")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Printf("  %s [global-options] running\n", os.Args[0])
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
	fmt.Printf("  %s running                    # Show current program\n", os.Args[0])
	fmt.Printf("  %s --verbose running          # Show with verbose HTTP output\n", os.Args[0])
	fmt.Println()
}

func queryRunningProgram() {
	controlunitURL := globalConfig.APIEndpoints.ControlUnit.URL
	url := controlunitURL + "/engine/running"

	if globalOpts.Verbose {
		fmt.Printf("Querying running program at: %s\n", url)
		fmt.Println()
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	if globalOpts.Verbose {
		fmt.Printf("GET %s\n", url)
	}

	req, err := http.NewRequest("GET", url, nil)
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
		return
	}

	if globalOpts.Verbose {
		fmt.Printf("HTTP Status: %d %s\n", resp.StatusCode, resp.Status)
		if len(respBody) > 0 {
			fmt.Printf("Raw Response: %s\n", string(respBody))
		}
		fmt.Println()
	}

	// Handle 204 No Content as "no program running"
	if resp.StatusCode == http.StatusNoContent {
		fmt.Println("No program currently running")
		return
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "Error: HTTP %d - %s\n", resp.StatusCode, string(respBody))
		return
	}

	var result types.APIResponse[types.ExecutionStatus]

	if err := json.Unmarshal(respBody, &result); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing response: %v\n", err)
		return
	}

	// Calculate elapsed time
	elapsedTime := int(time.Now().Unix() - result.Data.StartedAt)

	// Find current step to get target temperature and calculate remaining time
	var targetTemp uint8
	var remainingTime int
	var hasRuntime bool
	for _, step := range result.Data.Program.ProgramSteps {
		if step.Name == result.Data.CurrentStep {
			targetTemp = step.TargetTemperature
			// For steps with runtime, calculate remaining time
			if step.Runtime != nil && result.Data.CurrentStepStartedAt > 0 {
				hasRuntime = true
				stepElapsed := int(time.Now().Unix() - result.Data.CurrentStepStartedAt)
				remainingTime = int(step.Runtime.Seconds()) - stepElapsed
				if remainingTime < 0 {
					remainingTime = 0
				}
			}
			break
		}
	}

	fmt.Println("Currently Running Program")
	fmt.Println("=========================")
	fmt.Printf("Program Name:       %s\n", result.Data.Program.ProgramName)
	fmt.Printf("Current Phase:      %s\n", result.Data.CurrentStep)
	fmt.Printf("Elapsed Time:       %s\n", formatDuration(elapsedTime))
	if hasRuntime {
		fmt.Printf("Remaining Time:     %s\n", formatDuration(remainingTime))
	}
	fmt.Printf("Current Temp:       %.1f°C\n", result.Data.Temperatures.Material)
	fmt.Printf("Target Temp:        %d°C\n", targetTemp)
}

func formatDuration(seconds int) string {
	duration := time.Duration(seconds) * time.Second
	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60
	secs := int(duration.Seconds()) % 60

	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, secs)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, secs)
	}
	return fmt.Sprintf("%ds", secs)
}

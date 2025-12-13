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

func handleHistoryCommand() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Error: history command requires a subcommand\n\n")
		showHistoryHelp()
		os.Exit(exitError)
	}

	subcommand := os.Args[2]

	// Check for help flag
	for _, arg := range os.Args[3:] {
		if arg == "-h" || arg == helpFlag {
			showHistoryHelp()
			os.Exit(exitSuccess)
		}
	}

	switch subcommand {
	case "list":
		handleHistoryListCommand()
	case "show":
		handleHistoryShowCommand()
	case "log":
		handleHistoryLogCommand()
	default:
		fmt.Fprintf(os.Stderr, "Error: Unknown history subcommand '%s'\n\n", subcommand)
		showHistoryHelp()
		os.Exit(exitError)
	}
}

func showHistoryHelp() {
	fmt.Println("halkoctl history - Manage program execution history")
	fmt.Println()
	fmt.Println("View the execution history of programs and detailed information about specific runs.")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Printf("  %s [global-options] history <subcommand> [arguments]\n", os.Args[0])
	fmt.Println()
	fmt.Println("Subcommands:")
	fmt.Println("  list                  List all executed programs")
	fmt.Println("  show <program-name>   Show detailed information about a specific program run")
	fmt.Println("  log <program-name> [-o output-file]")
	fmt.Println("                        Display the execution log for a specific program run")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -h, --help")
	fmt.Println("        Show this help message")
	fmt.Println("  -o, --output string   (log subcommand only)")
	fmt.Println("        Write log output to specified file instead of stdout")
	fmt.Println()
	fmt.Println("Global Options:")
	fmt.Println("  -c, --config string")
	fmt.Println("        Path to the halko.cfg configuration file")
	fmt.Println("  -v, --verbose")
	fmt.Println("        Enable verbose output for HTTP requests")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Printf("  %s history list\n", os.Args[0])
	fmt.Printf("  %s history show \"Four-Stage Kiln Drying Program@2025-12-12T05:30:54Z\"\n", os.Args[0])
	fmt.Printf("  %s history log \"Four-Stage Kiln Drying Program@2025-12-12T05:30:54Z\"\n", os.Args[0])
	fmt.Printf("  %s history log \"Program@2025-12-12T05:30:54Z\" -o program-log.csv\n", os.Args[0])
	fmt.Printf("  %s --verbose history list\n", os.Args[0])
	fmt.Println()
}

func handleHistoryListCommand() {
	queryExecutionHistory()
	os.Exit(exitSuccess)
}

func handleHistoryShowCommand() {
	if len(os.Args) < 4 {
		fmt.Fprintf(os.Stderr, "Error: program name is required\n")
		fmt.Fprintf(os.Stderr, "Usage: %s history show <program-name>\n", os.Args[0])
		os.Exit(exitError)
	}

	programName := os.Args[3]
	queryProgramDetails(programName)
	os.Exit(exitSuccess)
}

func handleHistoryLogCommand() {
	if len(os.Args) < 4 {
		fmt.Fprintf(os.Stderr, "Error: program name is required\n")
		fmt.Fprintf(os.Stderr, "Usage: %s history log <program-name> [-o output-file]\n", os.Args[0])
		os.Exit(exitError)
	}

	programName := os.Args[3]
	var outputFile string

	// Parse optional output flag
	for i := 4; i < len(os.Args); i++ {
		if (os.Args[i] == "-o" || os.Args[i] == "--output") && i+1 < len(os.Args) {
			outputFile = os.Args[i+1]
			i++ // Skip the next argument
		}
	}

	queryProgramLog(programName, outputFile)
	os.Exit(exitSuccess)
}

func queryExecutionHistory() {
	controlunitURL := globalConfig.APIEndpoints.ControlUnit.URL
	url := controlunitURL + "/engine/history"

	if globalOpts.Verbose {
		fmt.Printf("Querying execution history at: %s\n", url)
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

	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "Error: HTTP %d - %s\n", resp.StatusCode, string(respBody))
		return
	}

	var result types.APIResponse[[]types.RunHistory]

	if err := json.Unmarshal(respBody, &result); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing response: %v\n", err)
		return
	}

	if len(result.Data) == 0 {
		fmt.Println("No executed programs found")
		return
	}

	fmt.Println("Program Execution History")
	fmt.Println("=========================")
	fmt.Println()

	for _, run := range result.Data {
		fmt.Printf("Program: %s\n", run.Name)
		fmt.Printf("  State:      %s\n", run.State)
		if run.StartedAt > 0 {
			startTime := time.Unix(run.StartedAt, 0)
			fmt.Printf("  Started:    %s\n", startTime.Format("2006-01-02 15:04:05"))
		}
		if run.CompletedAt > 0 {
			endTime := time.Unix(run.CompletedAt, 0)
			fmt.Printf("  Completed:  %s\n", endTime.Format("2006-01-02 15:04:05"))

			// Calculate and display duration if both times available
			if run.StartedAt > 0 {
				duration := time.Duration(run.CompletedAt-run.StartedAt) * time.Second
				fmt.Printf("  Duration:   %s\n", formatDurationLong(duration))
			}
		}
		fmt.Println()
	}
}

func formatDurationLong(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}

func queryProgramDetails(programName string) {
	controlunitURL := globalConfig.APIEndpoints.ControlUnit.URL
	url := controlunitURL + "/engine/history/" + programName

	if globalOpts.Verbose {
		fmt.Printf("Querying program details at: %s\n", url)
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

	if resp.StatusCode == http.StatusNotFound {
		fmt.Fprintf(os.Stderr, "Program not found: %s\n", programName)
		return
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "Error: HTTP %d - %s\n", resp.StatusCode, string(respBody))
		return
	}

	var result types.APIResponse[types.ExecutedProgram]

	if err := json.Unmarshal(respBody, &result); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing response: %v\n", err)
		return
	}

	run := result.Data
	fmt.Println("Program Execution Details")
	fmt.Println("=========================")
	fmt.Println()
	fmt.Printf("Program Name: %s\n", run.Name)
	fmt.Printf("State:        %s\n", run.State)
	if run.StartedAt > 0 {
		startTime := time.Unix(run.StartedAt, 0)
		fmt.Printf("Started:      %s\n", startTime.Format("2006-01-02 15:04:05"))
	}
	if run.CompletedAt > 0 {
		endTime := time.Unix(run.CompletedAt, 0)
		fmt.Printf("Completed:    %s\n", endTime.Format("2006-01-02 15:04:05"))

		if run.StartedAt > 0 {
			duration := time.Duration(run.CompletedAt-run.StartedAt) * time.Second
			fmt.Printf("Duration:     %s\n", formatDurationLong(duration))
		}
	}
	fmt.Println()

	// Display program details
	fmt.Println("Program Configuration")
	fmt.Println("=====================")
	fmt.Println()
	fmt.Printf("Name:         %s\n", run.Program.ProgramName)
	fmt.Printf("Steps:        %d\n", len(run.Program.ProgramSteps))
	fmt.Println()

	fmt.Println("Program Steps:")
	for i, step := range run.Program.ProgramSteps {
		fmt.Printf("\n  Step %d: %s\n", i+1, step.Name)
		fmt.Printf("    Type:              %s\n", step.StepType)
		fmt.Printf("    Target Temp:       %d°C\n", step.TargetTemperature)
		if step.Runtime != nil {
			fmt.Printf("    Runtime:           %s\n", step.Runtime.String())
		}
		if step.Heater != nil {
			fmt.Printf("    Heater Control:    %s\n", formatPowerControl(step.Heater))
		}
		if step.Fan != nil {
			fmt.Printf("    Fan Control:       %s\n", formatPowerControl(step.Fan))
		}
		if step.Humidifier != nil {
			fmt.Printf("    Humidifier Control: %s\n", formatPowerControl(step.Humidifier))
		}
	}
	fmt.Println()
}

func queryProgramLog(programName string, outputFile string) {
	controlunitURL := globalConfig.APIEndpoints.ControlUnit.URL
	url := controlunitURL + "/engine/history/" + programName + "/log"

	if globalOpts.Verbose {
		fmt.Printf("Querying program log at: %s\n", url)
		if outputFile != "" {
			fmt.Printf("Output file: %s\n", outputFile)
		}
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

	req.Header.Set("Accept", "text/csv")

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
		fmt.Println()
	}

	if resp.StatusCode == http.StatusNotFound {
		fmt.Fprintf(os.Stderr, "Log file not found for program: %s\n", programName)
		return
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "Error: HTTP %d - %s\n", resp.StatusCode, string(respBody))
		return
	}

	// Write to file or display to stdout
	if outputFile != "" {
		err := os.WriteFile(outputFile, respBody, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing to file '%s': %v\n", outputFile, err)
			return
		}
		fmt.Printf("Log saved to: %s\n", outputFile)
	} else {
		// Display the CSV log content to stdout
		fmt.Print(string(respBody))
	}
}
func formatPowerControl(power *types.PowerPidSettings) string {
	if power.Power != nil {
		return fmt.Sprintf("Simple (%d%%)", *power.Power)
	}
	if power.MinDelta != nil && power.MaxDelta != nil {
		return fmt.Sprintf("Delta (min: %.1f°C, max: %.1f°C)", *power.MinDelta, *power.MaxDelta)
	}
	if power.Pid != nil {
		return fmt.Sprintf("PID (Kp: %.2f, Ki: %.2f, Kd: %.2f)", power.Pid.Kp, power.Pid.Ki, power.Pid.Kd)
	}
	return "Not specified"
}

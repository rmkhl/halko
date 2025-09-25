package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/rmkhl/halko/types"
)

func handleValidateCommand() {
	// Create a new FlagSet for the validate command
	validateFlags := flag.NewFlagSet("validate", flag.ExitOnError)

	var (
		programPath = validateFlags.String("program", "", "Path to the program.json file to validate (required)")
		verbose     = validateFlags.Bool("verbose", false, "Enable verbose output")
		help        = validateFlags.Bool("help", false, "Show help for validate command")
	)

	// Parse the arguments starting from os.Args[2] (after "validate")
	if err := validateFlags.Parse(os.Args[2:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		os.Exit(exitError)
	}

	if *help {
		showValidateHelp()
		os.Exit(exitSuccess)
	}

	if *programPath == "" {
		fmt.Fprintf(os.Stderr, "Error: -program flag is required\n\n")
		showValidateHelp()
		os.Exit(exitError)
	}

	if *verbose {
		fmt.Printf("Validating program: %s\n", *programPath)
		if globalConfig != nil {
			fmt.Printf("Using config from loaded configuration\n")
		}
		fmt.Println()
	}

	// Validate the program
	err := validateProgram(*programPath, *verbose)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Validation failed: %v\n", err)
		os.Exit(exitError)
	}

	fmt.Println("✓ Program validation successful!")
	os.Exit(exitSuccess)
}

func showValidateHelp() {
	fmt.Println("halkoctl validate - Validate a Halko program file")
	fmt.Println()
	fmt.Println("This command validates a program.json file against the Halko program schema and business rules.")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  halkoctl validate -program <path-to-program.json> [options]")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -program string")
	fmt.Println("        Path to the program.json file to validate (required)")
	fmt.Println("  -verbose")
	fmt.Println("        Enable verbose output")
	fmt.Println("  -help")
	fmt.Println("        Show this help message")
	fmt.Println()
	fmt.Println("Global Options:")
	fmt.Println("  -config string")
	fmt.Println("        Path to the halko.cfg file (applied before command)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  halkoctl validate -program example/example-program-delta.json")
	fmt.Println("  halkoctl -config my-halko.cfg validate -program my-program.json -verbose")
}

func validateProgram(programPath string, verbose bool) error {
	// Use the globally loaded config
	if globalConfig == nil {
		return errors.New("no configuration loaded - this should not happen")
	}

	config := globalConfig
	if config.ExecutorConfig == nil || config.ExecutorConfig.Defaults == nil {
		return errors.New("config file missing executor defaults")
	}

	if verbose {
		fmt.Println("✓ Configuration loaded successfully")
	}

	// Load the program file
	if verbose {
		fmt.Printf("Loading program from: %s\n", programPath)
	}

	// Check if file exists
	if _, err := os.Stat(programPath); os.IsNotExist(err) {
		return fmt.Errorf("program file does not exist: %s", programPath)
	}

	data, err := os.ReadFile(programPath)
	if err != nil {
		return fmt.Errorf("failed to read program file: %w", err)
	}

	if verbose {
		fmt.Println("✓ Program file loaded successfully")
	}

	// Parse the JSON
	if verbose {
		fmt.Println("Parsing JSON...")
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

	// Apply defaults to the program (no need to duplicate since we're not saving it)
	if verbose {
		fmt.Println("Applying defaults...")
	}

	program.ApplyDefaults(config.ExecutorConfig.Defaults)

	if verbose {
		fmt.Println("✓ Defaults applied successfully")
		fmt.Println("Running validation...")
	}

	// Validate the program
	err = program.Validate()
	if err != nil {
		return fmt.Errorf("program validation failed: %w", err)
	}

	if verbose {
		fmt.Println("✓ Program validation completed successfully")

		// Show program structure
		fmt.Println()
		fmt.Println("Program structure:")
		fmt.Printf("  Name: %s\n", program.ProgramName)
		fmt.Printf("  Steps: %d\n", len(program.ProgramSteps))
		for i, step := range program.ProgramSteps {
			fmt.Printf("    %d. %s (%s) - Target: %d°C\n",
				i+1, step.Name, step.StepType, step.TargetTemperature)
			if step.Runtime != nil {
				fmt.Printf("       Runtime: %s\n", step.Runtime.String())
			}
		}
	}

	return nil
}

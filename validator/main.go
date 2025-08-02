package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/rmkhl/halko/types"
)

const (
	exitSuccess = 0
	exitError   = 1
)

func main() {
	var (
		programPath = flag.String("program", "", "Path to the program.json file to validate (required)")
		configPath  = flag.String("config", "", "Path to the halko.cfg file (optional, defaults to templates/halko.cfg)")
		verbose     = flag.Bool("verbose", false, "Enable verbose output")
		help        = flag.Bool("help", false, "Show help message")
	)
	flag.Parse()

	if *help {
		showHelp()
		os.Exit(exitSuccess)
	}

	if *programPath == "" {
		fmt.Fprintf(os.Stderr, "Error: -program flag is required\n\n")
		showHelp()
		os.Exit(exitError)
	}

	// Set default config path if not provided
	if *configPath == "" {
		// Try to find the config in order of preference
		if _, err := os.Stat("/etc/opt/halko/halko.cfg"); err == nil {
			*configPath = "/etc/opt/halko/halko.cfg"
		} else if _, err := os.Stat("templates/halko.cfg"); err == nil {
			*configPath = "templates/halko.cfg"
		} else if _, err := os.Stat("../templates/halko.cfg"); err == nil {
			*configPath = "../templates/halko.cfg"
		} else {
			fmt.Fprintf(os.Stderr, "Error: Could not find default config file. Please specify -config flag\n")
			os.Exit(exitError)
		}
	}

	if *verbose {
		fmt.Printf("Validating program: %s\n", *programPath)
		fmt.Printf("Using config: %s\n", *configPath)
		fmt.Println()
	}

	// Validate the program
	err := validateProgram(*programPath, *configPath, *verbose)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Validation failed: %v\n", err)
		os.Exit(exitError)
	}

	fmt.Println("✓ Program validation successful!")
	os.Exit(exitSuccess)
}

func showHelp() {
	fmt.Println("Halko Program Validator")
	fmt.Println()
	fmt.Println("This tool validates a program.json file against the Halko program schema and business rules.")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Printf("  %s -program <path-to-program.json> [options]\n", os.Args[0])
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -program string")
	fmt.Println("        Path to the program.json file to validate (required)")
	fmt.Println("  -config string")
	fmt.Println("        Path to the halko.cfg file (optional, defaults to /etc/opt/halko/halko.cfg)")
	fmt.Println("  -verbose")
	fmt.Println("        Enable verbose output")
	fmt.Println("  -help")
	fmt.Println("        Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Printf("  %s -program example/example-program-delta.json\n", os.Args[0])
	fmt.Printf("  %s -program my-program.json -config my-halko.cfg -verbose\n", os.Args[0])
}

func validateProgram(programPath, configPath string, verbose bool) error {
	// Load the configuration file
	if verbose {
		fmt.Printf("Loading configuration from: %s\n", configPath)
	}

	config, err := types.ReadHalkoConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	if config.ExecutorConfig == nil || config.ExecutorConfig.Defaults == nil {
		return fmt.Errorf("config file missing executor defaults")
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

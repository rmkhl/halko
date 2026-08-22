package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/rmkhl/halko/types"
)

func handleValidateCommand() {
	opts, err := ParseValidateOptions()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		os.Exit(exitError)
	}

	if opts.Help {
		showValidateHelp()
		os.Exit(exitSuccess)
	}

	if opts.ProgramPath == "" {
		fmt.Fprintf(os.Stderr, "Error: program file path is required\n\n")
		showValidateHelp()
		os.Exit(exitError)
	}

	if globalOpts.Verbose {
		fmt.Printf("Validating program: %s\n", opts.ProgramPath)
		if globalConfig != nil {
			fmt.Printf("Using config from loaded configuration\n")
		}
		fmt.Println()
	}

	err = validateProgram(opts.ProgramPath, globalOpts.Verbose)
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
	fmt.Printf("  %s [global-options] validate <program-file> [options]\n", os.Args[0])
	fmt.Println()
	fmt.Println("Arguments:")
	fmt.Println("  program-file      Path to the program.json file to validate (required)")
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
	fmt.Printf("  %s validate example/example-program-delta.json\n", os.Args[0])
	fmt.Printf("  %s --config /path/to/halko.cfg validate my-program.json\n", os.Args[0])
	fmt.Printf("  %s --verbose validate my-program.json\n", os.Args[0])
}

func validateProgram(programPath string, verbose bool) error {
	if globalConfig == nil {
		return errors.New("no configuration loaded - this should not happen")
	}

	config := globalConfig
	if config.ControlUnitConfig == nil || config.ControlUnitConfig.Defaults == nil {
		return errors.New("config file missing controlunit defaults")
	}

	if verbose {
		fmt.Println("✓ Configuration loaded successfully")
	}

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

	if verbose {
		fmt.Println("Applying defaults...")
	}

	program.ApplyDefaults(config.ControlUnitConfig.Defaults)

	if verbose {
		fmt.Println("✓ Defaults applied successfully")
		fmt.Println("Running validation...")
	}

	err = program.Validate()
	if err != nil {
		return fmt.Errorf("program validation failed: %w", err)
	}

	if verbose {
		fmt.Println("✓ Program validation completed successfully")

		fmt.Println()
		fmt.Print(describeProgram(&program))
	}

	return nil
}

// describeProgram renders what a program actually contains, for `validate
// --verbose`. The description is operator-written and may run to several
// lines, so its continuation lines are indented under the first.
func describeProgram(program *types.Program) string {
	var out strings.Builder

	out.WriteString("Program structure:\n")
	fmt.Fprintf(&out, "  Name: %s\n", program.ProgramName)
	if program.Description != "" {
		for i, line := range strings.Split(program.Description, "\n") {
			if i == 0 {
				fmt.Fprintf(&out, "  Description: %s\n", line)
				continue
			}
			fmt.Fprintf(&out, "               %s\n", line)
		}
	}
	fmt.Fprintf(&out, "  Steps: %d\n", len(program.ProgramSteps))
	for i, step := range program.ProgramSteps {
		fmt.Fprintf(&out, "    %d. %s (%s) - Target: %d°C\n",
			i+1, step.Name, step.StepType, step.TargetTemperature)
		if step.Runtime != nil {
			fmt.Fprintf(&out, "       Runtime: %s\n", step.Runtime.String())
		}
	}

	return out.String()
}

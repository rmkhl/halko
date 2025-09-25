package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/rmkhl/halko/types"
)

const (
	exitSuccess = 0
	exitError   = 1
	helpFlag    = "--help"
)

func main() {
	// Parse global flags first
	var configPath string
	commandIndex := -1

	// Find command and extract config flag
	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		if arg == "-c" || arg == "--config" {
			if i+1 < len(os.Args) {
				configPath = os.Args[i+1]
				i++ // skip the config path value
			}
		} else if !strings.HasPrefix(arg, "-") {
			commandIndex = i
			break
		}
	}

	if commandIndex == -1 {
		showHelp()
		os.Exit(exitError)
	}

	// Check for help commands before loading config
	command := os.Args[commandIndex]
	if command == "help" || command == "-h" || command == helpFlag {
		showHelp()
		os.Exit(exitSuccess)
	}

	// Check for command-specific help
	if commandIndex+1 < len(os.Args) {
		nextArg := os.Args[commandIndex+1]
		if nextArg == "-h" || nextArg == helpFlag {
			switch command {
			case "send":
				showSendHelp()
				os.Exit(exitSuccess)
			case "status":
				showStatusHelp()
				os.Exit(exitSuccess)
			case "validate":
				showValidateHelp()
				os.Exit(exitSuccess)
			}
		}
	}

	// Load configuration
	config, err := types.LoadConfig(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(exitError)
	}
	globalConfig = config

	// Remove processed global flags from os.Args for command parsing
	newArgs := []string{os.Args[0]}
	newArgs = append(newArgs, os.Args[commandIndex:]...)
	os.Args = newArgs

	switch command {
	case "send":
		handleSendCommand()
	case "status":
		handleStatusCommand()
	case "validate":
		handleValidateCommand()
	case "help", "-help", helpFlag:
		showHelp()
		os.Exit(exitSuccess)
	default:
		fmt.Fprintf(os.Stderr, "Error: Unknown command '%s'\n\n", command)
		showHelp()
		os.Exit(exitError)
	}
}

package main

import (
	"flag"
	"os"
	"strings"

	"github.com/rmkhl/halko/types"
)

// CommonOptions represents command-line options common to multiple commands (help only now)
type CommonOptions struct {
	Help bool // Show help message
}

// SendOptions represents options specific to the send command
type SendOptions struct {
	CommonOptions
	ProgramPath string // Path to the program.json file (positional argument)
}

// StatusOptions represents options specific to the status command
type StatusOptions struct {
	CommonOptions
}

// ValidateOptions represents options specific to the validate command
type ValidateOptions struct {
	CommonOptions
	ProgramPath string // Path to the program.json file (positional argument)
}

// DisplayOptions represents options specific to the display command
type DisplayOptions struct {
	CommonOptions
	Message string // Text message to display (positional argument)
}

// TemperaturesOptions represents options specific to the temperatures command
type TemperaturesOptions struct {
	CommonOptions
}

// ParseGlobalOptions parses global options from command line arguments
// Returns the parsed options and the index where the command starts
func ParseGlobalOptions() (*types.GlobalOptions, int) {
	commandIndex := -1
	globalArgs := []string{os.Args[0]} // Start with program name

	// Separate global options from command and its arguments
	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		switch arg {
		case "-c", "--config":
			globalArgs = append(globalArgs, arg)
			if i+1 < len(os.Args) && !strings.HasPrefix(os.Args[i+1], "-") {
				i++
				globalArgs = append(globalArgs, os.Args[i]) // add the config value
			}
		case "-v", "--verbose":
			globalArgs = append(globalArgs, arg)
		default:
			if !strings.HasPrefix(arg, "-") {
				commandIndex = i
				goto done
			}
		}
	}
done:

	// Temporarily replace os.Args to parse only global options
	originalArgs := os.Args
	os.Args = globalArgs

	// Use unified global options parsing
	opts, err := types.ParseGlobalOptions()
	if err != nil {
		// Handle error gracefully, fallback to default
		opts = &types.GlobalOptions{
			ConfigPath: "/etc/opt/halko.cfg",
			Verbose:    false,
		}
	}

	// Restore original os.Args
	os.Args = originalArgs

	return opts, commandIndex
}

// SetupCommonFlags adds common flags (help only) to a FlagSet
func SetupCommonFlags(flagSet *flag.FlagSet, opts *CommonOptions) {
	flagSet.BoolVar(&opts.Help, "h", false, "Show help message")
	flagSet.BoolVar(&opts.Help, "help", false, "Show help message")
}

// ParseSendOptions parses command-line options for the send command
func ParseSendOptions() (*SendOptions, error) {
	opts := &SendOptions{}
	sendFlags := flag.NewFlagSet("send", flag.ExitOnError)

	SetupCommonFlags(sendFlags, &opts.CommonOptions)

	if err := sendFlags.Parse(os.Args[2:]); err != nil {
		return nil, err
	}

	// Get the program path from remaining arguments
	args := sendFlags.Args()
	if len(args) > 0 {
		opts.ProgramPath = args[0]
	}

	return opts, nil
}

// ParseStatusOptions parses command-line options for the status command
func ParseStatusOptions() (*StatusOptions, error) {
	opts := &StatusOptions{}
	statusFlags := flag.NewFlagSet("status", flag.ExitOnError)

	SetupCommonFlags(statusFlags, &opts.CommonOptions)

	if err := statusFlags.Parse(os.Args[2:]); err != nil {
		return nil, err
	}

	return opts, nil
}

// ParseValidateOptions parses command-line options for the validate command
func ParseValidateOptions() (*ValidateOptions, error) {
	opts := &ValidateOptions{}
	validateFlags := flag.NewFlagSet("validate", flag.ExitOnError)

	SetupCommonFlags(validateFlags, &opts.CommonOptions)

	if err := validateFlags.Parse(os.Args[2:]); err != nil {
		return nil, err
	}

	// Get the program path from remaining arguments
	args := validateFlags.Args()
	if len(args) > 0 {
		opts.ProgramPath = args[0]
	}

	return opts, nil
}

// ParseDisplayOptions parses command-line options for the display command
func ParseDisplayOptions() (*DisplayOptions, error) {
	opts := &DisplayOptions{}
	displayFlags := flag.NewFlagSet("display", flag.ExitOnError)

	SetupCommonFlags(displayFlags, &opts.CommonOptions)

	if err := displayFlags.Parse(os.Args[2:]); err != nil {
		return nil, err
	}

	// Get the message text from remaining arguments
	args := displayFlags.Args()
	if len(args) > 0 {
		opts.Message = args[0]
	}

	return opts, nil
}

// ParseTemperaturesOptions parses command-line options for the temperatures command
func ParseTemperaturesOptions() (*TemperaturesOptions, error) {
	opts := &TemperaturesOptions{}
	temperaturesFlags := flag.NewFlagSet("temperatures", flag.ExitOnError)

	SetupCommonFlags(temperaturesFlags, &opts.CommonOptions)

	if err := temperaturesFlags.Parse(os.Args[2:]); err != nil {
		return nil, err
	}

	return opts, nil
}

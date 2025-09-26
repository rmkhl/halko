package types

import (
	"flag"
	"fmt"

	"github.com/rmkhl/halko/types/log"
)

// GlobalOptions represents options common to all modules (executor, powerunit, sensorunit, halkoctl)
type GlobalOptions struct {
	ConfigPath string       // Path to halko.cfg configuration file
	Verbose    bool         // Enable verbose output
	LogLevel   log.LogLevel // Log level (0=ERROR, 1=WARN, 2=INFO, 3=DEBUG, 4=TRACE)
}

// ParseGlobalOptions parses command-line options for all modules (executor, powerunit, sensorunit, halkoctl)
func ParseGlobalOptions() (*GlobalOptions, error) {
	opts := &GlobalOptions{
		ConfigPath: "", // Empty path - LoadConfig will determine the default location
		Verbose:    false,
		LogLevel:   log.INFO, // Default to INFO level (2)
	}

	configPath := flag.String("config", "", "Path to configuration file (accepts --config)")
	flag.StringVar(&opts.ConfigPath, "c", "", "Path to configuration file (shorthand)")

	verbose := flag.Bool("verbose", false, "Enable verbose output (accepts --verbose)")
	flag.BoolVar(&opts.Verbose, "v", false, "Enable verbose output (shorthand)")

	logLevel := flag.Int("loglevel", int(log.INFO), "Log level (0=ERROR, 1=WARN, 2=INFO, 3=DEBUG, 4=TRACE)")

	flag.Parse()

	if *configPath != "" {
		opts.ConfigPath = *configPath
	}
	if *verbose {
		opts.Verbose = *verbose
	}

	// Validate and set log level
	if *logLevel < 0 || *logLevel > 4 {
		return nil, fmt.Errorf("invalid log level %d: must be between 0 (ERROR) and 4 (TRACE)", *logLevel)
	}
	opts.LogLevel = log.LogLevel(*logLevel)

	return opts, nil
}

// ApplyLogLevel sets the global log level based on the parsed options
func (opts *GlobalOptions) ApplyLogLevel() {
	log.SetLevel(opts.LogLevel)
}

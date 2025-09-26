package types

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/rmkhl/halko/types/log"
)

// GlobalOptions represents options common to all modules (executor, powerunit, sensorunit, halkoctl)
type GlobalOptions struct {
	ConfigPath string        // Path to halko.cfg configuration file
	Verbose    bool          // Enable verbose output
	LogLevel   log.LogLevel  // Log level (0=ERROR, 1=WARN, 2=INFO, 3=DEBUG, 4=TRACE)
}

// SimulatorOptions represents options specific to the simulator module
type SimulatorOptions struct {
	Port string // Port to listen on
}

// ConfigFilePath returns the path to the config file, checking environment variable first,
// then searching in multiple standard locations. Returns error if no config file is found.
func ConfigFilePath() (string, error) {
	// Check environment variable first
	if configPath := os.Getenv("HALKO_CONFIG"); configPath != "" {
		if _, err := os.Stat(configPath); err == nil {
			return configPath, nil
		}
	}

	// Define search paths in priority order
	searchPaths := []string{
		"halko.cfg", // Current directory
	}

	if homeDir, err := os.UserHomeDir(); err == nil {
		searchPaths = append(searchPaths,
			filepath.Join(homeDir, ".halko.cfg"),           // ~/.halko.cfg
			filepath.Join(homeDir, ".config", "halko.cfg"), // ~/.config/halko.cfg
		)
	}

	searchPaths = append(searchPaths,
		"/etc/halko/halko.cfg",     // System config directory
		"/etc/opt/halko/halko.cfg", // Optional system config directory
	)

	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		executablePath := filepath.Join(exeDir, "halko.cfg")
		searchPaths = append(searchPaths, executablePath)
	}

	for _, path := range searchPaths {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("no halko.cfg file found in any of the search paths: %v", searchPaths)
}

// ParseGlobalOptions parses command-line options for all modules (executor, powerunit, sensorunit, halkoctl)
func ParseGlobalOptions() (*GlobalOptions, error) {
	defaultConfigPath, err := ConfigFilePath()
	if err != nil {
		return nil, err
	}

	opts := &GlobalOptions{
		ConfigPath: defaultConfigPath,
		Verbose:    false,
		LogLevel:   log.INFO, // Default to INFO level (2)
	}

	configPath := flag.String("config", defaultConfigPath, "Path to configuration file (accepts --config)")
	flag.StringVar(&opts.ConfigPath, "c", defaultConfigPath, "Path to configuration file (shorthand)")

	verbose := flag.Bool("verbose", false, "Enable verbose output (accepts --verbose)")
	flag.BoolVar(&opts.Verbose, "v", false, "Enable verbose output (shorthand)")

	logLevel := flag.Int("loglevel", int(log.INFO), "Log level (0=ERROR, 1=WARN, 2=INFO, 3=DEBUG, 4=TRACE)")

	flag.Parse()

	if *configPath != defaultConfigPath {
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

// ParseSimulatorOptions parses command-line options for the simulator module
func ParseSimulatorOptions() *SimulatorOptions {
	opts := &SimulatorOptions{
		Port: "8088",
	}

	port := flag.String("l", "8088", "Port to listen on (Default: 8088)")
	flag.Parse()

	opts.Port = *port
	return opts
}

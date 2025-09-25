package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/rmkhl/halko/sensorunit/router"
	"github.com/rmkhl/halko/sensorunit/serial"
	"github.com/rmkhl/halko/types"
)

// ConfigFilePath returns the path to the config file
func ConfigFilePath() string {
	if configPath := os.Getenv("HALKO_CONFIG"); configPath != "" {
		return configPath
	}
	return "/etc/opt/halko/halko.cfg"
}

func main() {
	// Parse command line flags
	configPath := flag.String("config", ConfigFilePath(), "Path to configuration file")
	flag.String("c", ConfigFilePath(), "Path to configuration file (shorthand)")
	flag.Parse()

	// If -c was used instead of -config, use that value
	configPathValue := *configPath
	if flag.Lookup("c").Value.String() != ConfigFilePath() {
		configPathValue = flag.Lookup("c").Value.String()
	}

	// Load configuration
	halkoConfig, err := types.LoadConfig(configPathValue)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Default values for sensor unit config
	serialDevice := "/dev/ttyUSB0"
	baudRate := 9600
	port := 8089 // Default port if not specified in the URL

	// Extract sensorunit specific configuration if available
	if halkoConfig.SensorUnit != nil {
		if halkoConfig.SensorUnit.SerialDevice != "" {
			serialDevice = halkoConfig.SensorUnit.SerialDevice
		}
		if halkoConfig.SensorUnit.BaudRate != 0 {
			baudRate = halkoConfig.SensorUnit.BaudRate
		}
	}

	// Extract port from executor.sensor_unit_url
	if halkoConfig.ExecutorConfig != nil && halkoConfig.ExecutorConfig.SensorUnitURL != "" {
		// Extract port from URL
		parts := strings.Split(halkoConfig.ExecutorConfig.SensorUnitURL, ":")
		if len(parts) == 3 {
			var extractedPort int
			_, err := fmt.Sscanf(parts[2], "%d", &extractedPort)
			if err == nil {
				port = extractedPort
			}
		}
	}

	// Create sensor unit connection
	sensorUnit, err := serial.NewSensorUnit(serialDevice, baudRate)
	if err != nil {
		log.Fatalf("Failed to create sensor unit: %v", err)
	}

	// Try to connect
	if err := sensorUnit.Connect(); err != nil {
		log.Printf("Warning: Failed to connect to sensor unit: %v", err)
		log.Printf("Will retry connection when handling requests")
	} else {
		log.Printf("Connected to sensor unit on %s", serialDevice)
		defer sensorUnit.Close()
	}

	// Create API and router
	api := router.NewAPI(sensorUnit)
	r := router.SetupRouter(api)

	// Create http server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: r,
	}

	// Set up signal handling for graceful shutdown
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	// Start server in a goroutine
	go func() {
		log.Printf("Starting sensorunit service on port %d", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error starting server: %s", err)
		}
	}()

	// Block until signal received
	sig := <-sigs
	log.Printf("Shutdown signal received: %v", sig)

	// Create a context with timeout for server shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Shutdown the server gracefully
	log.Println("Shutting down HTTP server...")
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	// Close the sensor unit connection
	log.Println("Closing sensor unit connection...")
	if err := sensorUnit.Close(); err != nil {
		log.Printf("Error closing sensor unit connection: %v", err)
	}

	log.Println("Sensorunit service exited gracefully")
}

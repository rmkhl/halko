package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rmkhl/halko/powerunit/power"
	"github.com/rmkhl/halko/powerunit/router"
	"github.com/rmkhl/halko/powerunit/shelly"
	"github.com/rmkhl/halko/types"
)

func main() {
	var configFileName string

	flag.StringVar(&configFileName, "c", "/etc/opt/halko.cfg", "Specify config file. Default is /etc/opt/halko.cfg")
	flag.Parse()

	configuration, err := types.ReadHalkoConfig(configFileName)
	if err != nil {
		log.Fatal(err)
	} else if configuration.PowerUnit == nil {
		log.Fatal("power unit configuration missing")
	}

	// Create a new Shelly controller with the configured address
	shellyController := shelly.New(configuration.PowerUnit.ShellyAddress)

	// Get cycle length from config or use default
	cycleLength := 60000 // Default to 60 seconds (60000 ms)
	if configuration.PowerUnit.CycleLength > 0 {
		cycleLength = configuration.PowerUnit.CycleLength
	}

	// Get power mapping from config
	powerMapping := configuration.PowerUnit.PowerMapping
	if powerMapping == nil {
		// Provide a default mapping if not in config
		powerMapping = map[string]int{
			"fan":        0,
			"heater":     1,
			"humidifier": 2,
		}
		log.Println("Power mapping not found in config, using default.")
	}

	p := power.New(cycleLength, shellyController, powerMapping)
	// Inject the powerMapping into the router package
	router.SetPowerMapping(powerMapping)
	r := router.New(p)

	// Extract port from configured power_unit_url
	serverPort := "8090" // Default port
	if configuration.ExecutorConfig != nil && configuration.ExecutorConfig.PowerUnitURL != "" {
		if parsedURL, err := url.Parse(configuration.ExecutorConfig.PowerUnitURL); err == nil {
			if parsedURL.Port() != "" {
				serverPort = parsedURL.Port()
			}
		}
	}
	serverAddr := ":" + serverPort
	log.Printf("Starting power unit server on %s", serverAddr)

	// Start the power controller in a goroutine
	go func() {
		err := p.Start()
		if err != nil {
			log.Printf("POWERUNIT START ERROR --- %s", err)
		}
	}()

	// Create a server
	srv := &http.Server{
		Addr:    serverAddr,
		Handler: r,
	}

	// Start the server in a goroutine
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Channel to listen for interrupt signals
	quit := make(chan os.Signal, 1)
	// Listen for SIGINT (Ctrl+C) and SIGTERM (systemctl stop)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Block until we receive a signal
	sig := <-quit
	log.Printf("Received signal %s, shutting down gracefully...", sig)

	log.Printf("Stopping power controller...")
	p.Stop()

	// Create a deadline for the shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Attempt graceful shutdown of the HTTP server
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server shutdown complete")
}

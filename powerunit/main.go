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

	configuration, err := types.LoadConfig(configFileName)
	if err != nil {
		log.Fatal(err)
	} else if configuration.PowerUnit == nil {
		log.Fatal("power unit configuration missing")
	}

	shellyController := shelly.New(configuration.PowerUnit.ShellyAddress)

	if configuration.PowerUnit.CycleLength <= 0 {
		log.Fatal("Invalid or missing cycle length")
	}
	cycleLength := configuration.PowerUnit.CycleLength

	if configuration.PowerUnit.MaxIdleTime <= 0 {
		log.Fatal("Invalid or missing max idle time")
	}
	maxIdleTime := configuration.PowerUnit.MaxIdleTime

	powerMapping := configuration.PowerUnit.PowerMapping
	if powerMapping == nil {
		log.Fatal("Power mapping not found in config, using default.")
	}

	idMapping := [shelly.NumberOfDevices]string{}
	for name, id := range powerMapping {
		idMapping[id] = name
	}
	// Extract port from configured power_unit_url
	if configuration.ExecutorConfig == nil || configuration.ExecutorConfig.PowerUnitURL == "" {
		log.Fatal("ExecutorConfig or PowerUnitURL not found in config")
	}

	var serverPort string
	if parsedURL, err := url.Parse(configuration.ExecutorConfig.PowerUnitURL); err == nil {
		if parsedURL.Port() == "" {
			log.Fatal("PowerUnitURL must include a port")
		}
		serverPort = parsedURL.Port()
	}
	serverAddr := ":" + serverPort

	p := power.New(maxIdleTime, cycleLength, shellyController)
	r := router.New(p, powerMapping, idMapping)

	log.Printf("Starting power unit server on %s", serverAddr)

	// Start the power controller in a goroutine
	go func() {
		err := p.Start()
		if err != nil {
			log.Printf("POWERUNIT START ERROR --- %s", err)
		}
	}()

	// Start the server in a goroutine
	srv := &http.Server{
		Addr:    serverAddr,
		Handler: r,
	}

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

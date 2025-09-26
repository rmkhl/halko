package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rmkhl/halko/sensorunit/router"
	"github.com/rmkhl/halko/sensorunit/serial"
	"github.com/rmkhl/halko/types"
	"github.com/rmkhl/halko/types/log"
)

func main() {
	log.Trace("Starting sensorunit main function")

	log.Trace("Parsing global options")
	opts, err := types.ParseGlobalOptions()
	if err != nil {
		log.Fatal("Failed to parse options: %v", err)
	}
	log.Trace("Global options parsed successfully")

	// Apply log level from command line
	opts.ApplyLogLevel()

	log.Trace("Loading configuration from %s", opts.ConfigPath)
	halkoConfig, err := types.LoadConfig(opts.ConfigPath)
	if err != nil {
		log.Fatal("Failed to load configuration: %v", err)
	}
	log.Trace("Configuration loaded successfully")

	serialDevice := halkoConfig.SensorUnit.SerialDevice
	baudRate := halkoConfig.SensorUnit.BaudRate
	log.Trace("Serial device configuration: device=%s, baudRate=%d", serialDevice, baudRate)

	port := halkoConfig.SensorUnit.Port
	if port == 0 {
		port = 8093 // Default port
		log.Trace("Using default port %d", port)
	} else {
		log.Trace("Using configured port %d", port)
	}

	log.Trace("Creating new sensor unit instance")
	sensorUnit, err := serial.NewSensorUnit(serialDevice, baudRate)
	if err != nil {
		log.Fatal("Failed to create sensor unit: %v", err)
	}
	log.Trace("Sensor unit instance created successfully")

	log.Trace("Attempting initial connection to sensor unit")
	if err := sensorUnit.Connect(); err != nil {
		log.Warning("Failed to connect to sensor unit: %v", err)
		log.Info("Will retry connection when handling requests")
		log.Trace("Initial connection failed, continuing with startup")
	} else {
		log.Info("Connected to sensor unit on %s", serialDevice)
		log.Trace("Initial connection successful, deferring close")
		defer sensorUnit.Close()
	}

	log.Trace("Creating API instance")
	api := router.NewAPI(sensorUnit)
	log.Trace("Setting up HTTP router")
	r := router.SetupRouter(api, halkoConfig.APIEndpoints)

	log.Trace("Creating HTTP server on port %d", port)
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: r,
	}

	log.Trace("Setting up signal handling for graceful shutdown")
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	go func() {
		log.Info("Starting sensorunit service on port %d", port)
		log.Trace("HTTP server listening and serving requests")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Error starting server: %s", err)
		}
		log.Trace("HTTP server stopped")
	}()

	log.Trace("Waiting for shutdown signal")
	sig := <-sigs
	log.Info("Shutdown signal received: %v", sig)

	log.Trace("Creating shutdown context with 5 second timeout")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Info("Shutting down HTTP server...")
	log.Trace("Initiating graceful HTTP server shutdown")
	if err := srv.Shutdown(ctx); err != nil {
		log.Error("Server forced to shutdown: %v", err)
		log.Trace("HTTP server shutdown completed with error")
	} else {
		log.Trace("HTTP server shutdown completed successfully")
	}

	log.Info("Closing sensor unit connection...")
	log.Trace("Closing sensor unit connection")
	if err := sensorUnit.Close(); err != nil {
		log.Error("Error closing sensor unit connection: %v", err)
		log.Trace("Sensor unit connection closed with error")
	} else {
		log.Trace("Sensor unit connection closed successfully")
	}

	log.Info("Sensorunit service exited gracefully")
	log.Trace("Main function completed")
}

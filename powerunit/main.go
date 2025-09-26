package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rmkhl/halko/powerunit/power"
	"github.com/rmkhl/halko/powerunit/router"
	"github.com/rmkhl/halko/powerunit/shelly"
	"github.com/rmkhl/halko/types"
	"github.com/rmkhl/halko/types/log"
)

func main() {
	opts, err := types.ParseGlobalOptions()
	if err != nil {
		log.Fatal("Failed to parse global options: %v", err)
	}

	// Apply the log level from the parsed options
	opts.ApplyLogLevel()
	log.Trace("Starting powerunit application")
	log.Debug("Parsed global options: config=%s, loglevel=%d, verbose=%t", opts.ConfigPath, opts.LogLevel, opts.Verbose)

	configuration, err := types.LoadConfig(opts.ConfigPath)
	if err != nil {
		log.Fatal("Failed to load configuration: %v", err)
	}
	log.Debug("Loaded configuration from %s", opts.ConfigPath)

	shellyController := shelly.New(configuration.PowerUnit.ShellyAddress)
	log.Debug("Created Shelly controller for address: %s", configuration.PowerUnit.ShellyAddress)

	cycleLength := configuration.PowerUnit.CycleLength
	maxIdleTime := configuration.PowerUnit.MaxIdleTime
	powerMapping := configuration.PowerUnit.PowerMapping
	log.Debug("Power unit configuration: cycleLength=%ds, maxIdleTime=%ds, powerMapping=%v",
		cycleLength, maxIdleTime, powerMapping)

	idMapping := [shelly.NumberOfDevices]string{}
	for name, id := range powerMapping {
		idMapping[id] = name
	}
	log.Trace("Created ID mapping: %v", idMapping)

	port, err := configuration.APIEndpoints.PowerUnit.GetPort()
	if err != nil {
		log.Fatal("Failed to get powerunit port: %v", err)
	}
	serverAddr := ":" + port
	log.Debug("Server will listen on %s", serverAddr)

	p := power.New(maxIdleTime, cycleLength, shellyController)
	log.Trace("Created power controller")

	r := router.New(p, powerMapping, idMapping, configuration.APIEndpoints)
	log.Trace("Created HTTP router")

	log.Info("Starting power unit server on %s", serverAddr)

	go func() {
		log.Trace("Starting power controller goroutine")
		err := p.Start()
		if err != nil {
			log.Error("POWERUNIT START ERROR --- %s", err)
		}
	}()

	srv := &http.Server{
		Addr:    serverAddr,
		Handler: r,
	}
	log.Trace("Created HTTP server")

	go func() {
		log.Trace("Starting HTTP server goroutine")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server failed to start: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	log.Debug("Signal handling setup complete, waiting for shutdown signal")

	sig := <-quit
	log.Info("Received signal %s, shutting down gracefully...", sig)

	log.Info("Stopping power controller...")
	p.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	log.Debug("Attempting graceful server shutdown with 5s timeout")
	if err := srv.Shutdown(ctx); err != nil {
		log.Error("Server forced to shutdown: %v", err)
	} else {
		log.Debug("Server shutdown completed successfully")
	}

	log.Info("Server shutdown complete")
}

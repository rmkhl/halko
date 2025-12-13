package main

import (
	"context"
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

	opts.ApplyLogLevel()
	log.Info("Sensorunit service starting")

	halkoConfig, err := types.LoadConfig(opts.ConfigPath)
	if err != nil {
		log.Fatal("Failed to load configuration: %v", err)
	}

	serialDevice := halkoConfig.SensorUnit.SerialDevice
	baudRate := halkoConfig.SensorUnit.BaudRate
	log.Trace("Serial device configuration: device=%s, baudRate=%d", serialDevice, baudRate)

	log.Trace("Creating new sensor unit instance")
	sensorUnit, err := serial.NewSensorUnit(serialDevice, baudRate)
	if err != nil {
		log.Fatal("Failed to create sensor unit: %v", err)
	}
	log.Trace("Sensor unit instance created successfully")

	log.Info("Initializing sensor unit connection on device: %s (baud: %d)", serialDevice, baudRate)
	if err := sensorUnit.Connect(); err != nil {
		log.Warning("Failed initial connection to sensor unit: %v", err)
		log.Info("Sensor unit connection will be established on demand")
	} else {
		log.Info("Sensor unit connected successfully on %s", serialDevice)
		defer sensorUnit.Close()
	}

	port, err := halkoConfig.APIEndpoints.SensorUnit.GetPort()
	if err != nil {
		log.Error("Failed to get sensorunit port: %v", err)
		return
	}

	api := router.NewAPI(sensorUnit)
	r := router.SetupRouter(api, halkoConfig.APIEndpoints)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	go func() {
		log.Info("Sensorunit HTTP server starting on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Error starting server: %s", err)
		}
		log.Info("HTTP server stopped")
	}()

	log.Info("Sensorunit service ready - waiting for requests")
	sig := <-sigs
	log.Info("Shutdown signal received: %v", sig)

	log.Trace("Creating shutdown context with 5 second timeout")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Info("Initiating graceful shutdown...")
	if err := srv.Shutdown(ctx); err != nil {
		log.Error("HTTP server forced shutdown: %v", err)
	} else {
		log.Info("HTTP server shutdown completed")
	}

	log.Info("Closing sensor unit connection...")
	if err := sensorUnit.Shutdown(); err != nil {
		log.Error("Error closing sensor unit connection: %v", err)
	} else {
		log.Info("Sensor unit connection closed")
	}

	log.Info("Sensorunit service shutdown complete")
	log.Trace("Main function completed")
}

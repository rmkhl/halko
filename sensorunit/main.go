package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rmkhl/halko/sensorunit/router"
	"github.com/rmkhl/halko/sensorunit/serial"
	"github.com/rmkhl/halko/types"
)

func main() {
	opts, err := types.ParseGlobalOptions()
	if err != nil {
		log.Fatalf("Failed to parse options: %v", err)
	}

	halkoConfig, err := types.LoadConfig(opts.ConfigPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	serialDevice := halkoConfig.SensorUnit.SerialDevice
	baudRate := halkoConfig.SensorUnit.BaudRate

	port := halkoConfig.SensorUnit.Port
	if port == 0 {
		port = 8093 // Default port
	}

	sensorUnit, err := serial.NewSensorUnit(serialDevice, baudRate)
	if err != nil {
		log.Fatalf("Failed to create sensor unit: %v", err)
	}

	if err := sensorUnit.Connect(); err != nil {
		log.Printf("Warning: Failed to connect to sensor unit: %v", err)
		log.Printf("Will retry connection when handling requests")
	} else {
		log.Printf("Connected to sensor unit on %s", serialDevice)
		defer sensorUnit.Close()
	}

	api := router.NewAPI(sensorUnit)
	r := router.SetupRouter(api, halkoConfig.APIEndpoints)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: r,
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	go func() {
		log.Printf("Starting sensorunit service on port %d", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error starting server: %s", err)
		}
	}()

	sig := <-sigs
	log.Printf("Shutdown signal received: %v", sig)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Println("Shutting down HTTP server...")
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Closing sensor unit connection...")
	if err := sensorUnit.Close(); err != nil {
		log.Printf("Error closing sensor unit connection: %v", err)
	}

	log.Println("Sensorunit service exited gracefully")
}

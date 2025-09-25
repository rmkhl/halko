package main

import (
	"context"
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
	opts, err := types.ParseGlobalOptions()
	if err != nil {
		log.Fatal(err)
	}

	configuration, err := types.LoadConfig(opts.ConfigPath)
	if err != nil {
		log.Fatal(err)
	}

	shellyController := shelly.New(configuration.PowerUnit.ShellyAddress)
	cycleLength := configuration.PowerUnit.CycleLength
	maxIdleTime := configuration.PowerUnit.MaxIdleTime
	powerMapping := configuration.PowerUnit.PowerMapping

	idMapping := [shelly.NumberOfDevices]string{}
	for name, id := range powerMapping {
		idMapping[id] = name
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

	go func() {
		err := p.Start()
		if err != nil {
			log.Printf("POWERUNIT START ERROR --- %s", err)
		}
	}()

	srv := &http.Server{
		Addr:    serverAddr,
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	sig := <-quit
	log.Printf("Received signal %s, shutting down gracefully...", sig)

	log.Printf("Stopping power controller...")
	p.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server shutdown complete")
}

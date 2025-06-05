package main

import (
	"context"
	"flag"
	"log"
	"net/http"
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

	flag.StringVar(&configFileName, "c", "/etc/halko.cfg", "Specify config file. Default is /etc/halko.cfg")
	flag.Parse()

	configuration, err := types.ReadHalkoConfig(configFileName)
	if err != nil {
		log.Fatal(err)
	} else if configuration.PowerUnit == nil {
		log.Fatal("power unit configuration missing")
	}

	s := shelly.New(configuration.PowerUnit.ShellyAddress)
	p := power.New(s)
	r := router.New(p)

	// Start the power controller in a goroutine
	go func() {
		err := p.Start()
		if err != nil {
			log.Printf("POWERUNIT START ERROR --- %s", err)
		}
	}()

	// Create a server
	srv := &http.Server{
		Addr:    ":8090",
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

	// Create a deadline for the shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Attempt graceful shutdown of the HTTP server
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	// Shutdown the Shelly client
	if err := s.Shutdown(); err != nil {
		log.Printf("SHELLY SHUTDOWN ERROR --- %s", err)
	}

	log.Println("Server shutdown complete")
}

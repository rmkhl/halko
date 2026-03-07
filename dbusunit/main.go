package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rmkhl/halko/dbusunit/dbus"
	"github.com/rmkhl/halko/dbusunit/router"
	"github.com/rmkhl/halko/types"
	"github.com/rmkhl/halko/types/log"
)

func main() {
	opts, err := types.ParseGlobalOptions()
	if err != nil {
		log.Fatal("Failed to parse global options: %v", err)
	}

	opts.ApplyLogLevel()
	log.Trace("Starting dbusunit application")
	log.Debug("Parsed global options: config=%s, loglevel=%d, verbose=%t", opts.ConfigPath, opts.LogLevel, opts.Verbose)

	configuration, err := types.LoadConfig(opts.ConfigPath)
	if err != nil {
		log.Fatal("Failed to load configuration: %v", err)
	}
	log.Debug("Loaded configuration from %s", opts.ConfigPath)

	dbusManager, err := dbus.NewManager()
	if err != nil {
		log.Fatal("Failed to create D-Bus manager: %v", err)
	}
	defer dbusManager.Close()
	log.Trace("D-Bus manager created successfully")

	mux := http.NewServeMux()
	router.SetupRoutes(mux, dbusManager, configuration.APIEndpoints)
	log.Trace("HTTP routes configured")

	port, err := configuration.APIEndpoints.DBusUnit.GetPort()
	if err != nil {
		log.Fatal("Failed to get dbusunit port: %v", err)
	}
	serverAddr := ":" + port
	log.Debug("Server will listen on %s", serverAddr)

	srv := &http.Server{
		Addr:    serverAddr,
		Handler: mux,
	}

	go func() {
		log.Info("Starting dbusunit server on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server failed to start: %v", err)
		}
		log.Info("HTTP server stopped")
	}()

	log.Info("DBusunit service ready - waiting for requests")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit

	log.Info("Received signal %v, shutting down dbusunit server...", sig)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("Server forced to shutdown: %v", err)
	}

	log.Info("DBusunit service exited cleanly")
}

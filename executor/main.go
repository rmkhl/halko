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

	"github.com/rmkhl/halko/executor/engine"
	"github.com/rmkhl/halko/executor/heartbeat"
	"github.com/rmkhl/halko/executor/router"
	"github.com/rmkhl/halko/executor/storage"
	"github.com/rmkhl/halko/types"
)

func addCORSHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:1234")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Length")
		w.Header().Set("Access-Control-Max-Age", "43200") // 12 hours

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	opts, err := types.ParseGlobalOptions()
	if err != nil {
		log.Fatal(err)
	}

	configuration, err := types.LoadConfig(opts.ConfigPath)
	if err != nil {
		log.Fatal(err)
	}

	storage, err := storage.NewFileStorage(configuration.ExecutorConfig.BasePath)
	if err != nil {
		log.Fatal(err)
	}

	engine := engine.NewEngine(configuration, storage, configuration.APIEndpoints)

	heartbeatManager, err := heartbeat.NewManager(configuration.ExecutorConfig)
	if err != nil {
		log.Fatalf("Failed to create heartbeat manager: %v", err)
	}
	if err := heartbeatManager.Start(); err != nil {
		log.Fatalf("Failed to start heartbeat manager: %v", err)
	}

	mux := http.NewServeMux()
	router.SetupRoutes(mux, storage, engine, configuration.APIEndpoints)

	corsHandler := addCORSHeaders(mux)

	port, err := configuration.APIEndpoints.Executor.GetPort()
	if err != nil {
		log.Fatalf("Failed to get executor port: %v", err)
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: corsHandler,
	}

	go func() {
		log.Printf("Starting executor server on port %d", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	sig := <-quit
	log.Printf("Received signal %s, shutting down gracefully...", sig)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	if err := engine.StopEngine(); err != nil {
		log.Printf("Error stopping engine: %s", err.Error())
	}
	if err := heartbeatManager.Stop(); err != nil {
		log.Printf("Error stopping heartbeat manager: %v", err)
	}

	log.Println("Server shutdown complete")
}

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

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/rmkhl/halko/executor/engine"
	"github.com/rmkhl/halko/executor/heartbeat"
	"github.com/rmkhl/halko/executor/router"
	"github.com/rmkhl/halko/executor/storage"
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

	storage, err := storage.NewFileStorage(configuration.ExecutorConfig.BasePath)
	if err != nil {
		log.Fatal(err)
	}

	engine := engine.NewEngine(configuration.ExecutorConfig, storage)

	heartbeatManager, err := heartbeat.NewManager(configuration.ExecutorConfig)
	if err != nil {
		log.Fatalf("Failed to create heartbeat manager: %v", err)
	}
	if err := heartbeatManager.Start(); err != nil {
		log.Fatalf("Failed to start heartbeat manager: %v", err)
	}

	server := gin.Default()
	server.Use(cors.New(cors.Config{
		AllowOrigins:  []string{"http://localhost:1234"},
		AllowMethods:  []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:  []string{"Origin", "Content-Type"},
		ExposeHeaders: []string{"Content-Length"},
		MaxAge:        12 * time.Hour,
	}))
	router.SetupRoutes(server, storage, engine)

	port := 8089
	if configuration.ExecutorConfig.Port > 0 {
		port = configuration.ExecutorConfig.Port
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: server,
	}

	// Start the server in a goroutine
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

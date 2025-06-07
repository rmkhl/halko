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

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/rmkhl/halko/executor/engine"
	"github.com/rmkhl/halko/executor/router"
	"github.com/rmkhl/halko/executor/storage"
	"github.com/rmkhl/halko/types"
)

func main() {
	var configFileName string

	flag.StringVar(&configFileName, "c", "/etc/opt/halko.cfg", "Specify config file. Default is /etc/opt/halko.cfg")
	flag.Parse()

	configuration, err := types.ReadHalkoConfig(configFileName)
	if err != nil {
		log.Fatal(err)
	}

	storage, err := storage.NewProgramStorage(configuration.ExecutorConfig.BasePath)
	if err != nil {
		log.Fatal(err)
	}

	engine := engine.NewEngine(configuration.ExecutorConfig, storage)
	server := gin.Default()
	server.Use(cors.New(cors.Config{
		AllowOrigins:  []string{"http://localhost:1234"},
		AllowMethods:  []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:  []string{"Origin", "Content-Type"},
		ExposeHeaders: []string{"Content-Length"},
		MaxAge:        12 * time.Hour,
	}))
	router.SetupRoutes(server, storage, engine)

	// Create a server
	srv := &http.Server{
		Addr:    ":8089",
		Handler: server,
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

	// Stop the engine
	if err := engine.StopEngine(); err != nil {
		log.Printf("Error stopping engine: %s", err.Error())
	}

	log.Println("Server shutdown complete")
}

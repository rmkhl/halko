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

	"github.com/rmkhl/halko/storage/filestorage"
	"github.com/rmkhl/halko/storage/router"
	"github.com/rmkhl/halko/types"
)

// addCORSHeaders adds CORS headers to responses
func addCORSHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:1234")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Length")
		w.Header().Set("Access-Control-Max-Age", "43200") // 12 hours

		// Handle preflight OPTIONS requests
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

	// Use storage config from configuration
	storageBasePath := "/tmp/halko" // Default fallback
	port := 8091                   // Default port

	if configuration.StorageConfig != nil {
		if configuration.StorageConfig.BasePath != "" {
			storageBasePath = configuration.StorageConfig.BasePath
		}
		if configuration.StorageConfig.Port != 0 {
			port = configuration.StorageConfig.Port
		}
	} else if configuration.ExecutorConfig != nil && configuration.ExecutorConfig.BasePath != "" {
		// Fallback to executor config if storage config is not available
		storageBasePath = configuration.ExecutorConfig.BasePath
	}

	storage, err := filestorage.NewFileStorage(storageBasePath)
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	router.SetupRoutes(mux, storage)

	// Add CORS middleware
	corsHandler := addCORSHeaders(mux)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: corsHandler,
	}

	// Start the server in a goroutine
	go func() {
		log.Printf("Starting storage server on port %d", port)
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

	log.Println("Storage server shutdown complete")
}

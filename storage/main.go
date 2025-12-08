package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rmkhl/halko/storage/router"
	"github.com/rmkhl/halko/storage/storagefs"
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

	var storageBasePath string
	switch {
	case configuration.StorageConfig != nil && configuration.StorageConfig.BasePath != "":
		storageBasePath = configuration.StorageConfig.BasePath
	case configuration.ExecutorConfig != nil && configuration.ExecutorConfig.BasePath != "":
		storageBasePath = configuration.ExecutorConfig.BasePath
	default:
		log.Fatal("Storage base path is not configured. Set storage.base_path or executor.base_path in configuration")
	}

	// Get port from storage endpoint
	if configuration.APIEndpoints == nil {
		log.Fatal("API endpoints configuration is required")
	}
	port, err := configuration.APIEndpoints.Storage.GetPort()
	if err != nil {
		log.Fatalf("Failed to get storage port from configuration: %v", err)
	}

	storage, err := storagefs.NewProgramStorage(storageBasePath)
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	router.SetupRoutes(mux, storage, configuration.APIEndpoints)

	corsHandler := addCORSHeaders(mux)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: corsHandler,
	}

	go func() {
		log.Printf("Starting storage server on port %s", port)
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

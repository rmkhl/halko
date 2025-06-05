package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/rmkhl/halko/simulator/elements"
	"github.com/rmkhl/halko/simulator/engine"
	"github.com/rmkhl/halko/simulator/router"
)

func main() {
	var wg sync.WaitGroup

	port := flag.String("l", "8088", "Port to listen on (Default: 8088)")
	flag.Parse()

	fan := elements.NewPower("Fan")
	humidifier := elements.NewPower("Humidifier")
	wood := elements.NewWood(20)
	heater := elements.NewHeater("oven", 20, wood)
	temperatureSensors := map[string]engine.TemperatureSensor{"oven": heater, "material": wood}
	shellyControls := map[int8]interface{}{0: heater, 1: fan, 2: heater}

	ticker := time.NewTicker(6000 * time.Millisecond)

	// Channel for shutdown signals
	stop := make(chan struct{})

	// Setup signal catching
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	server := gin.Default()
	server.Use(cors.New(cors.Config{
		AllowOrigins:  []string{"http://localhost:1234"},
		AllowMethods:  []string{"GET"},
		AllowHeaders:  []string{"Origin", "Content-Type"},
		ExposeHeaders: []string{"Content-Length"},
		MaxAge:        12 * time.Hour,
	}))
	router.SetupRoutes(server, temperatureSensors, shellyControls)

	// Create http server
	srv := &http.Server{
		Addr:    ":" + *port,
		Handler: server,
	}

	// Start simulation goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Println("Starting simulation loop")

		// run "simulated" environment
		for {
			select {
			case <-ticker.C:
				fan.Tick()
				humidifier.Tick()
				heater.Tick()
			case <-stop:
				log.Println("Stopping simulation loop")
				ticker.Stop()
				return
			}
		}
	}()

	// Start server in a goroutine
	go func() {
		log.Printf("Server running on port %s", *port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error starting server: %s", err)
		}
	}()

	// Block until signal received
	<-sigs
	log.Println("Shutdown signal received")

	// Close the stop channel to terminate the simulation loop
	close(stop)

	// Create a context with timeout for server shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Shutdown the server gracefully
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Waiting for simulation to complete...")
	wg.Wait()
	log.Println("Server exited gracefully")
}

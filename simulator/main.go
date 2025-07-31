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

	stop := make(chan struct{})

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

	srv := &http.Server{
		Addr:    ":" + *port,
		Handler: server,
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Println("Starting simulation loop")

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

	go func() {
		log.Printf("Server running on port %s", *port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error starting server: %s", err)
		}
	}()

	<-sigs
	log.Println("Shutdown signal received")

	close(stop)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Waiting for simulation to complete...")
	wg.Wait()
	log.Println("Server exited gracefully")
}

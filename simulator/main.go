package main

import (
	"context"
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

	// Sensor unit emulation on port 8088
	sensorServer := gin.Default()
	sensorServer.Use(cors.New(cors.Config{
		AllowOrigins:  []string{"http://localhost:1234"},
		AllowMethods:  []string{"GET", "POST"},
		AllowHeaders:  []string{"Origin", "Content-Type"},
		ExposeHeaders: []string{"Content-Length"},
		MaxAge:        12 * time.Hour,
	}))
	router.SetupSensorRoutes(sensorServer, temperatureSensors)

	sensorSrv := &http.Server{
		Addr:    ":8088",
		Handler: sensorServer,
	}

	// Shelly device emulation on port 8087
	shellyServer := gin.Default()
	shellyServer.Use(cors.New(cors.Config{
		AllowOrigins:  []string{"http://localhost:1234"},
		AllowMethods:  []string{"GET"},
		AllowHeaders:  []string{"Origin", "Content-Type"},
		ExposeHeaders: []string{"Content-Length"},
		MaxAge:        12 * time.Hour,
	}))
	router.SetupShellyRoutes(shellyServer, shellyControls)

	shellySrv := &http.Server{
		Addr:    ":8087",
		Handler: shellyServer,
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
		log.Println("Starting sensor unit server on port 8088")
		if err := sensorSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error starting sensor server: %s", err)
		}
	}()

	go func() {
		log.Println("Starting Shelly device server on port 8087")
		if err := shellySrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error starting Shelly server: %s", err)
		}
	}()

	<-sigs
	log.Println("Shutdown signal received")

	close(stop)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sensorSrv.Shutdown(ctx); err != nil {
		log.Printf("Sensor server forced to shutdown: %v", err)
	}
	if err := shellySrv.Shutdown(ctx); err != nil {
		log.Printf("Shelly server forced to shutdown: %v", err)
	}

	log.Println("Waiting for simulation to complete...")
	wg.Wait()
	log.Println("Server exited gracefully")
}

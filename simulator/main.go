package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/rmkhl/halko/simulator/elements"
	"github.com/rmkhl/halko/simulator/engine"
	"github.com/rmkhl/halko/simulator/router"
	"github.com/rmkhl/halko/types"
	"github.com/rmkhl/halko/types/log"
)

func main() {
	var wg sync.WaitGroup

	// Define simulator-specific flags
	tickDurationStr := flag.String("tick", "6s", "Simulation tick duration (e.g., 1s, 500ms, 100ms)")
	statusInterval := flag.Int("status-interval", 10, "Log simulation status every N ticks (0 to disable)")

	// Parse global options and load configuration like other services
	opts, err := types.ParseGlobalOptions()
	if err != nil {
		log.Fatal("Failed to parse global options: %v", err)
	}
	opts.ApplyLogLevel()

	// Parse tick duration
	tickDuration, err := time.ParseDuration(*tickDurationStr)
	if err != nil {
		log.Fatal("Invalid tick duration '%s': %v", *tickDurationStr, err)
	}
	log.Info("Simulation tick duration: %v", tickDuration)
	if *statusInterval > 0 {
		log.Info("Status logging every %d ticks", *statusInterval)
	} else {
		log.Info("Status logging disabled")
	}

	config, err := types.LoadConfig(opts.ConfigPath)
	if err != nil {
		log.Fatal("Failed to load configuration: %v", err)
	}

	// Extract Shelly port from powerunit.shelly_address
	shellyPort, err := config.APIEndpoints.SensorUnit.GetPort(config.PowerUnit.ShellyAddress)
	if err != nil {
		log.Fatal("Failed to extract shelly port from configuration: %v", err)
	}

	// Extract sensor port from api_endpoints.sensorunit.url
	sensorPort, err := config.APIEndpoints.SensorUnit.GetPort()
	if err != nil {
		log.Fatal("Failed to extract sensor port from configuration: %v", err)
	}

	log.Info("Starting Halko Simulator")
	log.Debug("Using configuration - ShellyPort=%s (from powerunit), SensorPort=%s (from api_endpoints)",
		shellyPort, sensorPort)

	fan := elements.NewPower("Fan")
	fan.TurnOn(false) // Start the power controller in off state
	humidifier := elements.NewPower("Humidifier")
	humidifier.TurnOn(false) // Start the power controller in off state
	wood := elements.NewWood(20)
	heater := elements.NewHeater("oven", 20, wood)
	heater.TurnOn(false) // Start the heater power controller in off state
	log.Info("Initialized simulation elements: Fan, Humidifier, Heater (oven), Wood (material)")

	// Build element lookup map
	elementsByName := map[string]interface{}{
		"heater":     heater,
		"fan":        fan,
		"humidifier": humidifier,
	}

	// Map power controls using configuration
	shellyControls := make(map[int8]interface{})
	for name, id := range config.PowerUnit.PowerMapping {
		if element, exists := elementsByName[name]; exists {
			shellyControls[int8(id)] = element
			log.Trace("Mapped Shelly switch %d to %s", id, name)
		} else {
			log.Warning("Power mapping references unknown element: %s", name)
		}
	}
	log.Info("Configured %d Shelly switch mappings from power_unit.power_mapping", len(shellyControls))

	temperatureSensors := map[string]engine.TemperatureSensor{"oven": heater, "material": wood}

	ticker := time.NewTicker(tickDuration)
	stop := make(chan struct{})
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// Create Shelly emulation server
	shellyMux := http.NewServeMux()
	router.SetupShellyRoutes(shellyMux, shellyControls)
	shellyHandler := router.CORSMiddleware(shellyMux)

	// Create SensorUnit emulation server
	sensorMux := http.NewServeMux()
	router.SetupSensorUnitRoutes(sensorMux, temperatureSensors, config.APIEndpoints.SensorUnit)
	sensorHandler := router.CORSMiddleware(sensorMux)

	shellySrv := &http.Server{
		Addr:    ":" + shellyPort,
		Handler: shellyHandler,
	}

	sensorSrv := &http.Server{
		Addr:    ":" + sensorPort,
		Handler: sensorHandler,
	}

	// Start simulation loop
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Info("Starting simulation loop")
		tickCount := 0

		for {
			select {
			case <-ticker.C:
				tickCount++
				log.Trace("Simulation tick #%d: updating elements", tickCount)
				fan.Tick()
				humidifier.Tick()
				heater.Tick()

				// Log status summary at configured interval
				if *statusInterval > 0 && tickCount%*statusInterval == 0 {
					log.Info("Simulation status - Tick #%d: Oven=%.1f°C, Material=%.1f°C, Heater=%v, Fan=%v, Humidifier=%v",
						tickCount, heater.Temperature(), wood.Temperature(), heater.IsOn(), fan.IsOn(), humidifier.IsOn())
				}
			case <-stop:
				log.Info("Stopping simulation loop at tick #%d", tickCount)
				ticker.Stop()
				return
			}
		}
	}()

	// Start Shelly emulation server
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Info("Shelly emulation server running on port %s", shellyPort)
		if err := shellySrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("Shelly server error: %s", err)
		}
	}()

	// Start SensorUnit emulation server
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Info("SensorUnit emulation server running on port %s", sensorPort)
		if err := sensorSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("SensorUnit server error: %s", err)
		}
	}()

	<-sigs
	log.Info("Shutdown signal received")

	close(stop)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Shutdown both servers
	log.Debug("Shutting down HTTP servers...")
	if err := shellySrv.Shutdown(ctx); err != nil {
		log.Warning("Shelly server forced to shutdown: %v", err)
	}

	if err := sensorSrv.Shutdown(ctx); err != nil {
		log.Warning("SensorUnit server forced to shutdown: %v", err)
	}

	log.Info("Waiting for all servers and simulation to complete...")
	wg.Wait()
	log.Info("All servers exited gracefully")
}

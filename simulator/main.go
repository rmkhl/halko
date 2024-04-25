package main

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rmkhl/halko/simulator/elements"
	"github.com/rmkhl/halko/simulator/router"
	"github.com/rmkhl/halko/simulator/types"
)

func main() {
	var wg sync.WaitGroup

	fan := elements.NewPower("Fan")
	humidifier := elements.NewPower("Humidifier")
	wood := elements.NewWood(20, 200)
	heater := elements.NewHeater("oven", 20, 200, wood)
	temperature_sensors := map[string]types.TemperatureSensor{"oven": heater, "material": wood}
	power_sensors := map[string]types.PowerSensor{"heater": heater, "fan": fan, "humidifier": humidifier}
	power_controls := map[string]types.PowerManager{"heater": heater, "fan": fan, "humidifier": humidifier}

	ticker := time.NewTicker(6000 * time.Millisecond)

	server := gin.Default()
	router.SetupRoutes(server, temperature_sensors, power_sensors, power_controls)

	wg.Add(1)
	go func() {
		defer wg.Done()
		// run "simulated" environment
		for {
			<-ticker.C
			fan.Tick()
			humidifier.Tick()
			heater.Tick()
		}
	}()

	server.Run(":8088")

	wg.Wait()
}

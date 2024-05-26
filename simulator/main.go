package main

import (
	"log"
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
	wood := elements.NewWood(20)
	heater := elements.NewHeater("oven", 20, wood)
	temperatureSensors := map[string]types.TemperatureSensor{"oven": heater, "material": wood}
	powerSensors := map[string]types.PowerSensor{"heater": heater, "fan": fan, "humidifier": humidifier}
	powerControls := map[string]types.PowerManager{"heater": heater, "fan": fan, "humidifier": humidifier}

	ticker := time.NewTicker(6000 * time.Millisecond)

	server := gin.Default()
	router.SetupRoutes(server, temperatureSensors, powerSensors, powerControls)

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

	err := server.Run(":8088")
	if err != nil {
		log.Println(err.Error())
	}

	wg.Wait()
}

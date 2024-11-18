package main

import (
	"log"
	"sync"
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
	powerSensors := map[string]engine.PowerSensor{"heater": heater, "fan": fan, "humidifier": humidifier}
	powerControls := map[string]engine.PowerManager{"heater": heater, "fan": fan, "humidifier": humidifier}

	ticker := time.NewTicker(6000 * time.Millisecond)

	server := gin.Default()
	server.Use(cors.New(cors.Config{
		AllowOrigins:  []string{"http://localhost:1234"},
		AllowMethods:  []string{"GET"},
		AllowHeaders:  []string{"Origin", "Content-Type"},
		ExposeHeaders: []string{"Content-Length"},
		MaxAge:        12 * time.Hour,
	}))
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

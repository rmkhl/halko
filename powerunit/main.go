package main

import (
	"flag"
	"log"

	"github.com/rmkhl/halko/powerunit/power"
	"github.com/rmkhl/halko/powerunit/router"
	"github.com/rmkhl/halko/powerunit/shelly"
	"github.com/rmkhl/halko/types"
)

func main() {
	var configFileName string

	flag.StringVar(&configFileName, "c", "/etc/halko.cfg", "Specify config file. Default is /etc/halko.cfg")
	flag.Parse()

	configuration, err := types.ReadHalkoConfig(configFileName)
	if err != nil {
		log.Fatal(err)
	} else if configuration.PowerUnit == nil {
		log.Fatal("power unit configuration missing")
	}

	s := shelly.New(configuration.PowerUnit.ShellyAddress)
	p := power.New(s)
	r := router.New(p)

	defer func() {
		if err := s.Shutdown(); err != nil {
			log.Printf("SHELLY SHUTDOWN ERROR --- %s", err)
		}
	}()

	go func() {
		err := p.Start()
		if err != nil {
			log.Printf("POWERUNIT START ERROR --- %s", err)
		}
	}()

	if err := r.Run(":8090"); err != nil {
		log.Printf("ROUTER RUN ERROR --- %s", err)
	}
}

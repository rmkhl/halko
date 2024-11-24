package main

import (
	"flag"
	"log"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/rmkhl/halko/executor/engine"
	"github.com/rmkhl/halko/executor/router"
	"github.com/rmkhl/halko/executor/storage"
	"github.com/rmkhl/halko/types"
)

func main() {
	var configFileName string

	flag.StringVar(&configFileName, "c", "/etc/halko.cfg", "Specify config file. Default is /etc/halko.cfg")
	flag.Parse()

	configuration, err := types.ReadHalkoConfig(configFileName)
	if err != nil {
		log.Fatal(err)
	}

	storage, err := storage.NewProgramStorage(configuration.ExecutorConfig.BasePath)
	if err != nil {
		log.Fatal(err)
	}

	engine := engine.NewEngine(configuration.ExecutorConfig, storage)
	server := gin.Default()
	server.Use(cors.New(cors.Config{
		AllowOrigins:  []string{"http://localhost:1234"},
		AllowMethods:  []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:  []string{"Origin", "Content-Type"},
		ExposeHeaders: []string{"Content-Length"},
		MaxAge:        12 * time.Hour,
	}))
	router.SetupRoutes(server, storage, engine)

	err = server.Run(":8089")
	if err != nil {
		log.Println(err.Error())
	}
	err = engine.StopEngine()
	if err != nil {
		log.Println(err.Error())
	}
}

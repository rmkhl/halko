package main

import (
	"encoding/json"
	"flag"
	"io"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/rmkhl/halko/executor/engine"
	"github.com/rmkhl/halko/executor/router"
	"github.com/rmkhl/halko/executor/storage"
	"github.com/rmkhl/halko/executor/types"
)

func readConfiguration(fileName string) (*types.ExecutorConfig, error) {
	jsonFile, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}

	defer jsonFile.Close()

	content, err := io.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}

	var config types.ExecutorConfig
	err = json.Unmarshal(content, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func main() {
	var configFileName string

	flag.StringVar(&configFileName, "c", ".halko.cfg", "Specify config file. Default is .halko.cfg")
	flag.Parse()

	configuration, err := readConfiguration(configFileName)
	if err != nil {
		log.Fatal(err)
	}

	storage, err := storage.NewProgramStorage(configuration.BasePath)
	if err != nil {
		log.Fatal(err)
	}

	engine := engine.NewEngine(configuration, storage)
	server := gin.Default()
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

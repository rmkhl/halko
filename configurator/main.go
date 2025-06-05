package main

import (
	"log"

	"github.com/rmkhl/halko/configurator/filesystem"
	"github.com/rmkhl/halko/configurator/router"
)

func main() {
	db := filesystem.New()
	r := router.New(db)

	if err := r.Run(); err != nil {
		log.Printf("Error starting server: %v", err)
	}
}

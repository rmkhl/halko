package main

import (
	"github.com/rmkhl/halko/configurator/filesystem"
	"github.com/rmkhl/halko/configurator/router"
)

func main() {
	db := filesystem.New()
	r := router.New(db)

	r.Run()
}

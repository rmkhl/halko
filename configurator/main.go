package main

import (
	"github.com/rmkhl/halko/configurator/database"
	"github.com/rmkhl/halko/configurator/filesystem"
	"github.com/rmkhl/halko/configurator/router"
)

func main() {
	programs := filesystem.Programs{}
	r := router.New(database.Interface{Programs: programs})

	r.Run()
}

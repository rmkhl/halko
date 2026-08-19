package main

import (
	"fmt"

	"github.com/rmkhl/halko/types"
)

// handleVersionCommand reports the version this binary was built from. Like
// the config command it is dispatched before the global config load: the
// version of a deployment is exactly what you want to know when its
// configuration is broken, so it must not depend on one.
func handleVersionCommand() {
	fmt.Println(versionLine())
}

// versionLine is the single rendering of the version, shared with the test so
// the two cannot drift.
func versionLine() string {
	return "halkoctl " + types.Version
}

func showVersionHelp() {
	fmt.Println("Usage: halkoctl version")
	fmt.Println()
	fmt.Println("Print the Halko version this binary was built from.")
	fmt.Println("Does not read the configuration or contact any service.")
}

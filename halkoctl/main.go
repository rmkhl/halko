package main

import (
	"fmt"
	"os"
)

const (
	exitSuccess = 0
	exitError   = 1
)

func main() {
	if len(os.Args) < 2 {
		showHelp()
		os.Exit(exitError)
	}

	command := os.Args[1]
	switch command {
	case "send":
		handleSendCommand()
	case "status":
		handleStatusCommand()
	case "help", "-help", "--help":
		showHelp()
		os.Exit(exitSuccess)
	default:
		fmt.Fprintf(os.Stderr, "Error: Unknown command '%s'\n\n", command)
		showHelp()
		os.Exit(exitError)
	}
}

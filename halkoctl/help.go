package main

import (
	"fmt"
	"os"
)

func showHelp() {
	fmt.Println("halkoctl - Halko Control Tool")
	fmt.Println()
	fmt.Println("A command-line tool for interacting with the Halko wood drying kiln control system.")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Printf("  %s <command> [options]\n", os.Args[0])
	fmt.Println()
	fmt.Println("Available Commands:")
	fmt.Println("  send        Send a program.json file to the executor to start execution")
	fmt.Println("  status      Get the status of the currently running program")
	fmt.Println("  validate    Validate a program.json file against schema and business rules")
	fmt.Println("  help        Show this help message")
	fmt.Println()
	fmt.Println("Use \"halkoctl <command> -h\" for more information about a command.")
	fmt.Println()
	fmt.Println("Global Options:")
	fmt.Println("  -c, --config     Path to halko.cfg configuration file (optional)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Printf("  %s send example/example-program-delta.json\n", os.Args[0])
	fmt.Printf("  %s --config /path/to/halko.cfg status -v\n", os.Args[0])
	fmt.Printf("  %s validate my-program.json --verbose\n", os.Args[0])
}

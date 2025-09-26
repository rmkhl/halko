package main

import (
	"fmt"
	"os"
)

func showHelp() {
	fmt.Println("halkoctl - Halko Control Tool")
	fmt.Println()
	fmt.Println("A command-line tool for interacting with the Halko system.")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Printf("  %s [global-options] <command> [command-options] [arguments]\n", os.Args[0])
	fmt.Println()
	fmt.Println("Global Options:")
	fmt.Println("  -c, --config string")
	fmt.Println("        Path to the halko.cfg configuration file")
	fmt.Println("  -v, --verbose")
	fmt.Println("        Enable verbose output")
	fmt.Println("  -h, --help")
	fmt.Println("        Show this help message")
	fmt.Println()
	fmt.Println("Available Commands:")
	fmt.Println("  send                  Send program to executor")
	fmt.Println("  status                Get program status")
	fmt.Println("  validate              Validate a program file")
	fmt.Println("  display               Send text to sensor unit display")
	fmt.Println("  temperatures          Get current temperatures from sensor unit")
	fmt.Println()
	fmt.Println("Command Help:")
	fmt.Printf("  %s <command> --help   Show help for a specific command\n", os.Args[0])
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Printf("  %s send example/example-program-delta.json\n", os.Args[0])
	fmt.Printf("  %s --config /path/to/halko.cfg status --verbose\n", os.Args[0])
	fmt.Printf("  %s --verbose validate my-program.json\n", os.Args[0])
	fmt.Printf("  %s display \"Hello World\"\n", os.Args[0])
	fmt.Printf("  %s temperatures\n", os.Args[0])
}

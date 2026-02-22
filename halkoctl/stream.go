package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/websocket"
)

func showStreamHelp() {
	fmt.Println("Usage: halkoctl stream")
	fmt.Println()
	fmt.Println("Connect to the live execution log WebSocket and display messages.")
	fmt.Println("This command is useful for debugging to see exactly what data the WebSocket sends.")
	fmt.Println()
	fmt.Println("Press Ctrl+C to stop.")
}

func handleStreamCommand() {
	streamFlags := flag.NewFlagSet("stream", flag.ExitOnError)
	streamFlags.Usage = showStreamHelp
	if err := streamFlags.Parse(os.Args[2:]); err != nil {
		os.Exit(exitError)
	}

	baseURL := globalConfig.APIEndpoints.ControlUnit.GetURL()
	// Convert http://host:port to ws://host:port
	wsURL := baseURL
	if len(wsURL) >= 7 && wsURL[:7] == "http://" {
		wsURL = "ws://" + wsURL[7:]
	} else if len(wsURL) >= 8 && wsURL[:8] == "https://" {
		wsURL = "wss://" + wsURL[8:]
	}
	wsURL += "/engine/running/logws"

	fmt.Printf("Connecting to WebSocket at %s...\n", wsURL)

	conn, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error connecting to WebSocket: %v\n", err)
		os.Exit(exitError)
	}
	defer conn.Close()
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}

	fmt.Println("Connected! Receiving messages (press Ctrl+C to stop):")
	fmt.Println()

	// Handle Ctrl+C gracefully
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	done := make(chan struct{})

	go func() {
		defer close(done)
		messageCount := 0
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				fmt.Printf("\nWebSocket closed: %v\n", err)
				return
			}
			messageCount++
			fmt.Printf("[Message %d] %s\n", messageCount, string(message))
		}
	}()

	<-interrupt
	fmt.Println("\nInterrupted, closing connection...")
	conn.Close()
	<-done
}

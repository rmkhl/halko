package router

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rmkhl/halko/controlunit/engine"
	"github.com/rmkhl/halko/types/log"
)

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(_ *http.Request) bool { return true },
}

// Serve live CSV log for running program via websocket
func StreamLiveRunLog(engine *engine.ControlEngine) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := wsUpgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Warning("WebSocket upgrade failed: %v", err)
			return
		}
		log.Debug("WebSocket client connected for live run log")
		defer conn.Close()

		status := engine.CurrentStatus()
		if status == nil {
			if err := conn.WriteMessage(websocket.TextMessage, []byte("No program running")); err != nil {
				log.Warning("WebSocket write error: %v", err)
			}
			return
		}

		// Write CSV header (same as ExecutionLogWriter)
		var buf bytes.Buffer
		csvWriter := csv.NewWriter(&buf)
		if err := csvWriter.Write([]string{"time", "step", "steptime", "material", "oven", "heater", "fan", "humidifier"}); err != nil {
			log.Warning("CSV header write error: %v", err)
		}
		csvWriter.Flush()
		if err := conn.WriteMessage(websocket.TextMessage, bytes.TrimSpace(buf.Bytes())); err != nil {
			log.Warning("WebSocket write error: %v", err)
			return
		}

		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		// Set up ping/pong to detect dead connections
		if err := conn.SetReadDeadline(time.Now().Add(60 * time.Second)); err != nil {
			log.Warning("Failed to set read deadline: %v", err)
		}
		conn.SetPongHandler(func(string) error {
			return conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		})

		// Start a goroutine to read (and ignore) messages, but detect closure
		done := make(chan struct{})
		go func() {
			defer close(done)
			for {
				if _, _, err := conn.ReadMessage(); err != nil {
					return
				}
			}
		}()

		for {
			select {
			case <-done:
				// Connection closed
				log.Debug("WebSocket client disconnected from live run log")
				return
			case <-ticker.C:
				// Check if program is still running
				status := engine.CurrentStatus()
				if status == nil {
					return
				}
				// Only skip true initialization (Waiting), stream Pre-Heat, Initializing, and all program steps
				if status.CurrentStep == "" || status.CurrentStep == "Waiting" || status.CurrentStep == "Completed" {
					continue
				}
				// Compose CSV line
				var line bytes.Buffer
				csvw := csv.NewWriter(&line)
				elapsed := int(time.Now().Unix() - status.StartedAt)
				steptime := int(time.Now().Unix() - status.CurrentStepStartedAt)
				if err := csvw.Write([]string{
					strconv.Itoa(elapsed),
					status.CurrentStep,
					strconv.Itoa(steptime),
					fmt.Sprintf("%.1f", status.Temperatures.Material),
					fmt.Sprintf("%.1f", status.Temperatures.Oven),
					strconv.Itoa(int(status.PowerStatus.Heater)),
					strconv.Itoa(int(status.PowerStatus.Fan)),
					strconv.Itoa(int(status.PowerStatus.Humidifier)),
				}); err != nil {
					log.Warning("CSV line write error: %v", err)
					continue
				}
				csvw.Flush()
				if err := conn.WriteMessage(websocket.TextMessage, bytes.TrimSpace(line.Bytes())); err != nil {
					log.Debug("WebSocket client disconnected from live run log: %v", err)
					return
				}
			}
		}
	}
}

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
		log.Info("WebSocket client connected for live run log")
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
		}

		start := status.StartedAt
		for {
			// Check if program is still running
			status := engine.CurrentStatus()
			if status == nil {
				break
			}
			// Compose CSV line
			var line bytes.Buffer
			csvw := csv.NewWriter(&line)
			elapsed := int(time.Now().Unix() - start)
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
			}
			csvw.Flush()
			if err := conn.WriteMessage(websocket.TextMessage, bytes.TrimSpace(line.Bytes())); err != nil {
				log.Info("WebSocket client disconnected from live run log: %v", err)
				break // Stop on write failure (client disconnected)
			}
			time.Sleep(1 * time.Second)
		}
	}
}

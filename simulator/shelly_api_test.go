package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

// An emulated switch only applies a queued state change after ten simulation
// ticks, modelling the relay's switching cycle. The loop ticks at
// controlunit.tick_length, so a short tick keeps that cycle observable in a
// test instead of the 60s it takes in a real run.
const testTickLength = "20ms"

// The simulator needs a halko configuration to start. Only the Shelly address,
// the tick length and the emulated serial device matter here; the rest is the
// shape LoadConfig insists on.
const shellyTestConfig = `{
  "controlunit": {
    "base_path": "%s",
    "tick_length": "%s",
    "network_interface": "enp4s0",
    "defaults": {
      "pid_settings": {"acclimate": {"kp": 2.0, "ki": 1.0, "kd": 0.5}},
      "max_delta_heating": 10.0,
      "min_delta_heating": 5.0
    }
  },
  "power_unit": {
    "shelly_address": "http://localhost:8088",
    "cycle_length": "60s",
    "max_idle_time": "70s",
    "power_mapping": {"heater": 0, "humidifier": 1, "fan": 2}
  },
  "sensorunit": {"serial_device": "%s", "baud_rate": 9600},
  "api_endpoints": {
    "controlunit": {"url": "http://localhost:8090", "status": "/status", "programs": "/programs", "engine": "/engine"},
    "sensorunit": {"url": "http://localhost:8093", "status": "/status", "temperatures": "/temperatures", "display": "/display"},
    "powerunit": {"url": "http://localhost:8092", "status": "/status", "power": "/power"}
  }
}`

// writeTestConfig writes a halko config into a temporary directory. The
// simulator creates the serial device it emulates, so the config must name a
// path it may create rather than real hardware.
func writeTestConfig(t *testing.T) string {
	t.Helper()

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test_halko.cfg")
	config := fmt.Sprintf(shellyTestConfig, tempDir, testTickLength, filepath.Join(tempDir, "esp32"))

	if err := os.WriteFile(configPath, []byte(config), 0o644); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}
	return configPath
}

func TestShellyAPI(t *testing.T) {
	// The simulator reads the Shelly port from power_unit.shelly_address
	// in the halko configuration (8088 in shellyTestConfig).
	port := "8088"
	baseURL := "http://localhost:" + port

	configPath := writeTestConfig(t)

	// Build the simulator and run the binary directly: signals must reach the
	// simulator process itself so Wait() only returns once it has fully shut
	// down and released its ports (go run exits before its child does).
	simBinary := filepath.Join(t.TempDir(), "simulator")
	if out, err := exec.Command("go", "build", "-o", simBinary, ".").CombinedOutput(); err != nil {
		t.Fatalf("Error building simulator: %v\n%s", err, out)
	}

	t.Log("Starting simulator...")
	cmd := exec.Command(simBinary, "-c", configPath, "-s", "simulator.conf")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("Error creating stdout pipe: %v", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		t.Fatalf("Error creating stderr pipe: %v", err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatalf("Error starting simulator: %v", err)
	}

	// Ensure simulator is terminated when test ends
	defer func() {
		t.Log("Terminating simulator...")
		if err := cmd.Process.Signal(syscall.SIGTERM); err != nil {
			t.Logf("Error sending SIGTERM to simulator: %v", err)
			// Force kill if SIGTERM fails
			if err := cmd.Process.Kill(); err != nil {
				t.Logf("Error killing simulator: %v", err)
			}
		}

		// Wait for the simulator to exit
		if err := cmd.Wait(); err != nil {
			// Exit status 143 is normal for SIGTERM, so don't treat it as an error
			if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() != 143 {
				t.Logf("Simulator exited with error: %v", err)
			}
		}
		t.Log("Simulator terminated")
	}()

	// Stream simulator output to the test log
	go func() {
		combined := io.MultiReader(stdout, stderr)
		buf := make([]byte, 1024)
		for {
			n, err := combined.Read(buf)
			if n > 0 {
				t.Log(strings.TrimRight(string(buf[:n]), "\n"))
			}
			if err != nil {
				if err != io.EOF {
					t.Logf("Error reading simulator output: %v", err)
				}
				return
			}
		}
	}()

	// Wait for the Shelly server to accept requests
	statusURL := baseURL + "/rpc/Switch.GetStatus?id=0"
	deadline := time.Now().Add(30 * time.Second)
	for {
		resp, err := http.Get(statusURL)
		if err == nil {
			resp.Body.Close()
			t.Log("Simulator is ready")
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("Simulator did not become ready in time: %v", err)
		}
		time.Sleep(200 * time.Millisecond)
	}

	// Switch.Set takes "on=true|false", the same parameter powerunit's Shelly
	// client sends and the one a real Gen2 Shelly expects.
	t.Run("SwitchSetTogglesTheOutput", func(t *testing.T) {
		const switchID = 0
		getURL := fmt.Sprintf("%s/rpc/Switch.GetStatus?id=%d", baseURL, switchID)

		if got := switchOutput(t, getURL); got {
			t.Fatalf("switch %d output = true before any Set, want false", switchID)
		}

		setOn := fmt.Sprintf("%s/rpc/Switch.Set?id=%d&on=true", baseURL, switchID)
		if wasOn := setSwitch(t, setOn); wasOn {
			t.Errorf("Switch.Set reported was_on = true, want false")
		}
		waitForOutput(t, getURL, true)

		setOff := fmt.Sprintf("%s/rpc/Switch.Set?id=%d&on=false", baseURL, switchID)
		if wasOn := setSwitch(t, setOff); !wasOn {
			t.Errorf("Switch.Set reported was_on = false, want true")
		}
		waitForOutput(t, getURL, false)
	})

	t.Run("SwitchSetRejectsABadRequest", func(t *testing.T) {
		for _, tt := range []struct {
			name string
			url  string
		}{
			{"missing id", baseURL + "/rpc/Switch.Set?on=true"},
			{"missing on", baseURL + "/rpc/Switch.Set?id=0"},
			{"unknown switch", baseURL + "/rpc/Switch.Set?id=9&on=true"},
		} {
			t.Run(tt.name, func(t *testing.T) {
				status, body := request(t, tt.url)
				if status == http.StatusOK {
					t.Errorf("status = 200 for %s, want a client error (body %q)", tt.name, body)
				}
			})
		}
	})
}

// request performs a GET and returns the status code and body.
func request(t *testing.T, url string) (int, string) {
	t.Helper()

	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("GET %s: %v", url, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("reading response from %s: %v", url, err)
	}
	return resp.StatusCode, strings.TrimSpace(string(body))
}

// switchOutput reads a switch's current output state via Switch.GetStatus.
func switchOutput(t *testing.T, url string) bool {
	t.Helper()

	status, body := request(t, url)
	if status != http.StatusOK {
		t.Fatalf("Switch.GetStatus returned %d: %s", status, body)
	}

	var decoded struct {
		Output bool `json:"output"`
	}
	if err := json.Unmarshal([]byte(body), &decoded); err != nil {
		t.Fatalf("decoding Switch.GetStatus response %q: %v", body, err)
	}
	return decoded.Output
}

// waitForOutput polls Switch.GetStatus until the output reaches want, giving
// the emulated relay time to complete its switching cycle.
func waitForOutput(t *testing.T, url string, want bool) {
	t.Helper()

	deadline := time.Now().Add(10 * time.Second)
	for {
		if switchOutput(t, url) == want {
			return
		}
		if time.Now().After(deadline) {
			t.Errorf("switch output never became %v", want)
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// setSwitch drives Switch.Set and returns the state the switch was in before.
func setSwitch(t *testing.T, url string) bool {
	t.Helper()

	status, body := request(t, url)
	if status != http.StatusOK {
		t.Fatalf("Switch.Set returned %d: %s", status, body)
	}

	var decoded struct {
		WasOn bool `json:"was_on"`
	}
	if err := json.Unmarshal([]byte(body), &decoded); err != nil {
		t.Fatalf("decoding Switch.Set response %q: %v", body, err)
	}
	return decoded.WasOn
}

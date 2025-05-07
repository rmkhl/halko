package simulator_test

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"testing"
	"time"
)

// TestShellyAPI tests the Shelly API endpoints of the simulator
func TestShellyAPI(t *testing.T) {
	// Get port from environment variable or use default
	port := "8088"
	if envPort := os.Getenv("TEST_PORT"); envPort != "" {
		port = envPort
	}

	baseURL := fmt.Sprintf("http://localhost:%s", port)

	// Relative path to the simulator main.go from the test directory
	simulatorPath := "../../simulator/main.go"

	// Start the simulator as a subprocess
	t.Log("Starting simulator...")
	cmd := exec.Command("go", "run", simulatorPath, "-l", port)

	// Set up pipes for stdout and stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("Error creating stdout pipe: %v", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		t.Fatalf("Error creating stderr pipe: %v", err)
	}

	// Start the command
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

	// Create a channel to indicate when the simulator is ready
	ready := make(chan struct{})

	// Monitor simulator output in a goroutine
	go func() {
		defer close(ready)

		// Combine stdout and stderr into a single reader
		combined := io.MultiReader(stdout, stderr)
		buf := make([]byte, 1024)

		for {
			n, err := combined.Read(buf)
			if err != nil {
				if err != io.EOF {
					t.Logf("Error reading simulator output: %v", err)
				}
				return
			}

			// Print simulator output
			output := string(buf[:n])
			t.Log(output)

			// If we see the "Server running" message, signal that the simulator is ready
			if strings.Contains(output, "Server running") {
				ready <- struct{}{}
			}
		}
	}()

	// Wait for the simulator to be ready or timeout
	select {
	case <-ready:
		t.Log("Simulator is ready")
	case <-time.After(5 * time.Second):
		t.Log("Proceeding with tests (timeout waiting for ready message)")
	}

	// Wait a bit to ensure the server is actually ready
	time.Sleep(1 * time.Second)

	// Test cases
	t.Run("ShellyAPIOperations", func(t *testing.T) {
		switchID := 0
		turnOn := true

		// Get initial switch status
		getURL := fmt.Sprintf("%s/rpc/Switch.GetStatus?id=%d", baseURL, switchID)
		t.Logf("Getting status for switch %d", switchID)
		getResp, err := makeRequest(getURL)
		if err != nil {
			t.Fatalf("Error getting switch status: %v", err)
		}
		t.Logf("Initial status response: %s", getResp)

		// Set switch state
		setURL := fmt.Sprintf("%s/rpc/Switch.Set?id=%d&turn=%s", baseURL, switchID, boolToOnOff(turnOn))
		t.Logf("Setting switch %d to %s", switchID, boolToOnOff(turnOn))
		setResp, err := makeRequest(setURL)
		if err != nil {
			t.Fatalf("Error setting switch state: %v", err)
		}
		t.Logf("Set response: %s", setResp)

		// Get switch status again to verify change
		t.Logf("Getting status for switch %d after change", switchID)
		getResp2, err := makeRequest(getURL)
		if err != nil {
			t.Fatalf("Error getting switch status: %v", err)
		}
		t.Logf("Final status response: %s", getResp2)

		// Test that the simulator logged the query parameters
		// This is a simple test to verify that the simulator is working as expected
		// The actual validation would depend on what is returned by the simulator
		if getResp == "" {
			t.Error("Expected non-empty response from Switch.GetStatus")
		}
		if setResp == "" {
			t.Error("Expected non-empty response from Switch.Set")
		}
	})
}

// makeRequest performs an HTTP GET request and returns the response body as a string
func makeRequest(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// boolToOnOff converts a boolean to "on" or "off"
func boolToOnOff(b bool) string {
	if b {
		return "on"
	}
	return "off"
}

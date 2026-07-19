package tests

import (
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestShellyAPI(t *testing.T) {
	// The simulator reads the Shelly port from power_unit.shelly_address
	// in the halko configuration (8088 in testConfigData).
	port := "8088"
	baseURL := "http://localhost:" + port

	configPath := createTestConfigFile(t)

	// Build the simulator and run the binary directly: signals must reach the
	// simulator process itself so Wait() only returns once it has fully shut
	// down and released its ports (go run exits before its child does).
	simBinary := filepath.Join(t.TempDir(), "simulator")
	if out, err := exec.Command("go", "build", "-o", simBinary, "../simulator").CombinedOutput(); err != nil {
		t.Fatalf("Error building simulator: %v\n%s", err, out)
	}

	t.Log("Starting simulator...")
	cmd := exec.Command(simBinary, "-c", configPath, "-s", "../simulator/simulator.conf")

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

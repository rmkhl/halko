package esp32

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/rmkhl/halko/simulator/elements"
	"github.com/rmkhl/halko/simulator/faults"
)

// openSlave opens the emulated device the way sensorunit does — read/write,
// no controlling terminal, non-blocking so deadlines work.
func openSlave(t *testing.T, path string) *os.File {
	t.Helper()

	f, err := os.OpenFile(path, os.O_RDWR|syscall.O_NOCTTY|syscall.O_NONBLOCK, 0)
	if err != nil {
		t.Fatalf("failed to open the emulated device at %s: %v", path, err)
	}
	t.Cleanup(func() { f.Close() })
	return f
}

// exchange writes a command and returns the reply, failing rather than
// hanging if none arrives.
func exchange(t *testing.T, slave *os.File, command string) string {
	t.Helper()

	if _, err := slave.Write([]byte(command)); err != nil {
		t.Fatalf("failed to write %q: %v", command, err)
	}

	type result struct {
		text string
		err  error
	}
	done := make(chan result, 1)
	go func() {
		buf := make([]byte, 256)
		n, err := slave.Read(buf)
		done <- result{text: string(buf[:n]), err: err}
	}()

	select {
	case got := <-done:
		if got.err != nil {
			t.Fatalf("failed to read the reply to %q: %v", command, got.err)
		}
		return got.text
	case <-time.After(2 * time.Second):
		t.Fatalf("timed out waiting for a reply to %q", command)
		return ""
	}
}

func newTestDevice(t *testing.T) (*Device, string) {
	t.Helper()

	path := filepath.Join(t.TempDir(), "esp32")
	device, err := Open(path)
	if err != nil {
		t.Fatalf("failed to open the device: %v", err)
	}
	t.Cleanup(func() { device.Close() })

	wood := elements.NewWood(20.0, 20.0)
	heater := elements.NewHeater("kiln", 20.0, 20.0, wood)
	responder := NewResponder([]Probe{
		{Name: "KilnPrimary", Sensor: heater},
		{Name: "KilnSecondary", Sensor: heater},
		{Name: "Wood", Sensor: wood},
	}, faults.New(false), &recordingDisplay{})

	go device.Serve(responder)
	return device, path
}

func TestDeviceAnswersHeloOverThePTY(t *testing.T) {
	_, path := newTestDevice(t)
	slave := openSlave(t, path)

	if got := exchange(t, slave, "helo;"); got != "helo\r\n" {
		t.Fatalf("expected %q, got %q", "helo\r\n", got)
	}
}

func TestDeviceAnswersReadOverThePTY(t *testing.T) {
	_, path := newTestDevice(t)
	slave := openSlave(t, path)

	want := "KilnPrimary=20.00C,KilnSecondary=20.00C,Wood=20.00C\r\n"
	if got := exchange(t, slave, "read;"); got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestDeviceDoesNotEchoTheCommandBack(t *testing.T) {
	_, path := newTestDevice(t)
	slave := openSlave(t, path)

	// A tty with ECHO left on would return the command itself before the
	// reply, which would desynchronise the sensor unit's scanner.
	if got := exchange(t, slave, "helo;"); strings.Contains(got, "helo;") {
		t.Fatalf("expected no echo of the command, got %q", got)
	}
}

func TestOpenReplacesItsOwnStaleSymlink(t *testing.T) {
	path := filepath.Join(t.TempDir(), "esp32")

	first, err := Open(path)
	if err != nil {
		t.Fatalf("failed to open the device: %v", err)
	}
	first.Close()

	second, err := Open(path)
	if err != nil {
		t.Fatalf("expected a stale symlink to be replaced, got %v", err)
	}
	second.Close()
}

func TestOpenRefusesToReplaceARealFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ttyUSB1")
	if err := os.WriteFile(path, []byte("not a tty"), 0o600); err != nil {
		t.Fatalf("failed to create the blocking file: %v", err)
	}

	device, err := Open(path)
	if err == nil {
		device.Close()
		t.Fatalf("expected Open to refuse a path that is not a symlink")
	}
	if !strings.Contains(err.Error(), path) {
		t.Fatalf("expected the error to name %s, got %v", path, err)
	}
}

func TestCloseRemovesTheSymlink(t *testing.T) {
	path := filepath.Join(t.TempDir(), "esp32")

	device, err := Open(path)
	if err != nil {
		t.Fatalf("failed to open the device: %v", err)
	}
	if err := device.Close(); err != nil {
		t.Fatalf("failed to close the device: %v", err)
	}

	if _, err := os.Lstat(path); !os.IsNotExist(err) {
		t.Fatalf("expected the symlink removed, got %v", err)
	}
}

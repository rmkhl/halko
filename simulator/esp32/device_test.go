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

func TestDeviceMasterDoesNotReceiveItsOwnReply(t *testing.T) {
	// The sensor unit connects on the slave side, so it would never see a
	// stray echo of the command it sent (writes to the slave are not fed
	// back through the line discipline that ECHO governs). The real hazard
	// of leaving ECHO on is on the master side: Serve writes its reply with
	// d.master.Write, and with ECHO on, the tty would loop that write
	// straight back into d.master's own read side, feeding the emulator its
	// own reply as if it were a new incoming command. Exercise that path
	// directly, bypassing Serve, so this test fails if setRaw ever stops
	// disabling ECHO.
	path := filepath.Join(t.TempDir(), "esp32")
	device, err := Open(path)
	if err != nil {
		t.Fatalf("failed to open the device: %v", err)
	}
	t.Cleanup(func() { device.Close() })

	if _, err := device.master.Write([]byte("helo\r\n")); err != nil {
		t.Fatalf("failed to write the reply: %v", err)
	}

	if err := device.master.SetReadDeadline(time.Now().Add(100 * time.Millisecond)); err != nil {
		t.Fatalf("failed to set a read deadline: %v", err)
	}
	buf := make([]byte, 256)
	n, err := device.master.Read(buf)
	if err == nil {
		t.Fatalf("expected no data looped back to the master, got %q", buf[:n])
	}
	if !os.IsTimeout(err) {
		t.Fatalf("expected a read timeout (no loopback), got %v", err)
	}
}

func TestOpenReplacesItsOwnStaleSymlink(t *testing.T) {
	path := filepath.Join(t.TempDir(), "esp32")

	// A stale symlink is what a SIGKILLed simulator leaves behind: the
	// target pty is long gone, but the link itself remains. Create that
	// directly, rather than via a device we then close, so Open genuinely
	// exercises the "replace an existing symlink" branch instead of the
	// "path does not exist" one.
	if err := os.Symlink("/dev/pts/nonexistent", path); err != nil {
		t.Fatalf("failed to create the stale symlink: %v", err)
	}

	device, err := Open(path)
	if err != nil {
		t.Fatalf("expected a stale symlink to be replaced, got %v", err)
	}
	t.Cleanup(func() { device.Close() })

	target, err := os.Readlink(path)
	if err != nil {
		t.Fatalf("failed to read the replaced symlink: %v", err)
	}
	if !strings.HasPrefix(target, "/dev/pts/") {
		t.Fatalf("expected the stale symlink replaced with a pty under /dev/pts/, got %q", target)
	}
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

func TestOpenRefusesARealDevicePath(t *testing.T) {
	// This is the common development case: the ESP32 is not plugged in, so
	// the configured path does not exist, and without this guard Open would
	// fall through to os.Symlink and fail with a confusing EACCES instead of
	// naming the actual problem. This must not depend on whether
	// /dev/ttyUSB1 exists on the machine running the test.
	path := "/dev/ttyUSB1"

	device, err := Open(path)
	if err == nil {
		device.Close()
		t.Fatalf("expected Open to refuse a path under /dev/ that is not under /dev/pts/")
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

func TestServeReturnsAfterClose(t *testing.T) {
	// main.go runs Serve in a goroutine and calls Close before wg.Wait(). If
	// Serve ever stopped returning once closed, the simulator would hang
	// forever on shutdown instead of failing a test.
	path := filepath.Join(t.TempDir(), "esp32")
	device, err := Open(path)
	if err != nil {
		t.Fatalf("failed to open the device: %v", err)
	}

	responder, _ := newTestResponder(faults.New(false))

	done := make(chan struct{})
	go func() {
		device.Serve(responder)
		close(done)
	}()

	if err := device.Close(); err != nil {
		t.Fatalf("failed to close the device: %v", err)
	}

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatalf("Serve did not return within 2s of Close")
	}
}

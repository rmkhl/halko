package esp32

import (
	"errors"
	"fmt"
	"os"
	"syscall"
	"time"

	"github.com/rmkhl/halko/types/log"
	"golang.org/x/sys/unix"
)

// readRetryDelay keeps a persistently failing read from spinning.
const readRetryDelay = 50 * time.Millisecond

// Device is the pseudo-terminal the sensor unit opens in place of the USB
// serial device.
type Device struct {
	path   string
	master *os.File
	// slave stays open for the lifetime of the device. Without a slave
	// handle of our own, every read on the master returns EIO the moment
	// the sensor unit closes its end — for instance while it is being
	// restarted.
	slave *os.File
}

// Open allocates a pseudo-terminal and links it at path. It refuses to
// replace anything there that is not a symlink, so a configuration still
// naming real hardware fails loudly instead of producing a confusing
// permission error.
func Open(path string) (*Device, error) {
	// O_NONBLOCK matters beyond the flag itself: it makes the runtime treat
	// the master as pollable, so Close unblocks a Read that Serve is parked
	// in. Without it, shutdown would block forever on that goroutine.
	master, err := os.OpenFile("/dev/ptmx", os.O_RDWR|syscall.O_NOCTTY|syscall.O_NONBLOCK, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to open /dev/ptmx: %w", err)
	}

	if err := unix.IoctlSetPointerInt(int(master.Fd()), unix.TIOCSPTLCK, 0); err != nil {
		master.Close()
		return nil, fmt.Errorf("failed to unlock the pseudo-terminal: %w", err)
	}

	number, err := unix.IoctlGetInt(int(master.Fd()), unix.TIOCGPTN)
	if err != nil {
		master.Close()
		return nil, fmt.Errorf("failed to read the pseudo-terminal number: %w", err)
	}
	slavePath := fmt.Sprintf("/dev/pts/%d", number)

	slave, err := os.OpenFile(slavePath, os.O_RDWR|syscall.O_NOCTTY, 0)
	if err != nil {
		master.Close()
		return nil, fmt.Errorf("failed to open %s: %w", slavePath, err)
	}

	if err := setRaw(int(slave.Fd())); err != nil {
		slave.Close()
		master.Close()
		return nil, fmt.Errorf("failed to put %s in raw mode: %w", slavePath, err)
	}

	if err := link(path, slavePath); err != nil {
		slave.Close()
		master.Close()
		return nil, err
	}

	log.Info("ESP32 emulation listening on %s (%s)", path, slavePath)
	return &Device{path: path, master: master, slave: slave}, nil
}

// Serve reads commands from the device and writes the responder's replies,
// returning once the device is closed.
func (d *Device) Serve(responder *Responder) {
	var buffer commandBuffer
	chunk := make([]byte, 256)

	for {
		n, err := d.master.Read(chunk)
		if err != nil {
			if errors.Is(err, os.ErrClosed) {
				log.Debug("ESP32 emulation stopped")
				return
			}
			// EIO here means every reader has gone away. Our own slave
			// handle should prevent that, but a transient error must not
			// kill the emulation for the rest of the run.
			log.Debug("ESP32 emulation read failed, retrying: %v", err)
			time.Sleep(readRetryDelay)
			continue
		}

		for _, command := range buffer.Feed(chunk[:n]) {
			log.Trace("ESP32 emulation received command %q", command)
			reply := responder.Respond(command, time.Now())
			if reply == nil {
				continue
			}
			if _, err := d.master.Write(reply); err != nil {
				log.Error("ESP32 emulation failed to reply to %q: %v", command, err)
			}
		}
	}
}

// Close releases the pseudo-terminal and removes the symlink.
func (d *Device) Close() error {
	err := d.master.Close()
	if slaveErr := d.slave.Close(); err == nil {
		err = slaveErr
	}
	if linkErr := os.Remove(d.path); err == nil && !os.IsNotExist(linkErr) {
		err = linkErr
	}
	return err
}

// link points path at target, replacing only a symlink.
func link(path, target string) error {
	info, err := os.Lstat(path)
	switch {
	case err == nil && info.Mode()&os.ModeSymlink == 0:
		return fmt.Errorf("refusing to create the emulated device at %s: it already exists and is not a symlink (sensorunit.serial_device must name a path the simulator may create, not real hardware)", path)
	case err == nil:
		if err := os.Remove(path); err != nil {
			return fmt.Errorf("failed to remove the stale device link %s: %w", path, err)
		}
	case !os.IsNotExist(err):
		return fmt.Errorf("failed to inspect %s: %w", path, err)
	}

	if err := os.Symlink(target, path); err != nil {
		return fmt.Errorf("failed to link %s to %s: %w", path, target, err)
	}
	return nil
}

// setRaw disables line discipline processing on the slave, so bytes arrive
// as written and the emulation does not read back its own replies.
func setRaw(fd int) error {
	termios, err := unix.IoctlGetTermios(fd, unix.TCGETS)
	if err != nil {
		return err
	}

	termios.Iflag &^= unix.IGNBRK | unix.BRKINT | unix.PARMRK | unix.ISTRIP |
		unix.INLCR | unix.IGNCR | unix.ICRNL | unix.IXON
	termios.Oflag &^= unix.OPOST
	termios.Lflag &^= unix.ECHO | unix.ECHONL | unix.ICANON | unix.ISIG | unix.IEXTEN
	termios.Cflag &^= unix.CSIZE | unix.PARENB
	termios.Cflag |= unix.CS8

	return unix.IoctlSetTermios(fd, unix.TCSETS, termios)
}

package dbus

import (
	"fmt"

	"github.com/godbus/dbus/v5"
	"github.com/rmkhl/halko/types/log"
)

const (
	login1Destination = "org.freedesktop.login1"
	login1Path        = "/org/freedesktop/login1"
	login1Interface   = "org.freedesktop.login1.Manager"
)

// Shutdown powers off the system
func (m *Manager) Shutdown(delay int) error {
	log.Warning("System shutdown requested")

	// Get separate D-Bus connection for logind
	conn, err := dbus.ConnectSystemBus()
	if err != nil {
		return fmt.Errorf("failed to connect to system bus: %w", err)
	}
	defer conn.Close()

	// Get logind object
	obj := conn.Object(login1Destination, dbus.ObjectPath(login1Path))

	// Call PowerOff method
	// First parameter (true) means "interactive" - allows non-root users with polkit
	call := obj.Call(login1Interface+".PowerOff", 0, true)
	if call.Err != nil {
		return fmt.Errorf("failed to shutdown system: %w", call.Err)
	}

	log.Info("System shutdown initiated")
	return nil
}

// Reboot restarts the system
func (m *Manager) Reboot(delay int) error {
	log.Warning("System reboot requested")

	// Get separate D-Bus connection for logind
	conn, err := dbus.ConnectSystemBus()
	if err != nil {
		return fmt.Errorf("failed to connect to system bus: %w", err)
	}
	defer conn.Close()

	// Get logind object
	obj := conn.Object(login1Destination, dbus.ObjectPath(login1Path))

	// Call Reboot method
	// First parameter (true) means "interactive" - allows non-root users with polkit
	call := obj.Call(login1Interface+".Reboot", 0, true)
	if call.Err != nil {
		return fmt.Errorf("failed to reboot system: %w", call.Err)
	}

	log.Info("System reboot initiated")
	return nil
}

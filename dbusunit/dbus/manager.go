package dbus

import (
	"context"
	"fmt"
	"os"

	"github.com/coreos/go-systemd/v22/dbus"
	"github.com/rmkhl/halko/types/log"
)

// Manager handles D-Bus connections and operations
type Manager struct {
	conn *dbus.Conn
}

// NewManager creates a new D-Bus manager with system bus connection.
// A non-empty socketPath overrides the default system bus socket location.
func NewManager(socketPath string) (*Manager, error) {
	if socketPath != "" {
		// godbus reads DBUS_SYSTEM_BUS_ADDRESS when dialing the system bus
		if err := os.Setenv("DBUS_SYSTEM_BUS_ADDRESS", "unix:path="+socketPath); err != nil {
			return nil, fmt.Errorf("failed to set D-Bus address: %w", err)
		}
		log.Info("Using D-Bus system bus socket %s", socketPath)
	}
	conn, err := dbus.NewSystemConnectionContext(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to D-Bus: %w", err)
	}
	log.Info("Connected to systemd via D-Bus")
	return &Manager{conn: conn}, nil
}

// Close closes the D-Bus connection
func (m *Manager) Close() {
	if m.conn != nil {
		m.conn.Close()
		log.Debug("D-Bus connection closed")
	}
}

// IsConnected checks if the D-Bus connection is active
func (m *Manager) IsConnected() bool {
	return m.conn != nil
}

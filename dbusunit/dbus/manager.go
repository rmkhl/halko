package dbus

import (
	"fmt"

	"github.com/coreos/go-systemd/v22/dbus"
	"github.com/rmkhl/halko/types/log"
)

// Manager handles D-Bus connections and operations
type Manager struct {
	conn *dbus.Conn
}

// NewManager creates a new D-Bus manager with system bus connection
func NewManager() (*Manager, error) {
	conn, err := dbus.NewSystemConnection()
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

package dbus

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/rmkhl/halko/types/log"
)

// VPNStatus represents the status of an OpenVPN client
type VPNStatus struct {
	Name     string `json:"name"`
	Status   string `json:"status"` // active, inactive, failed
	Enabled  bool   `json:"enabled"`
	TunnelIP string `json:"tunnel_ip,omitempty"`
}

// openvpnUnitPrefixes lists the systemd unit naming conventions used for
// OpenVPN client connections. Debian's openvpn package historically runs
// each /etc/openvpn/<name>.conf as "openvpn@<name>.service", while newer
// setups (see RASPBERRY_PI.md) use "openvpn-client@<name>.service" with
// configs under /etc/openvpn/client/.
var openvpnUnitPrefixes = []string{"openvpn-client@", "openvpn@"}

// vpnNameFromUnit extracts the connection name from a systemd unit name if
// it matches one of the known OpenVPN naming conventions.
func vpnNameFromUnit(unitName string) (string, bool) {
	for _, prefix := range openvpnUnitPrefixes {
		if strings.HasPrefix(unitName, prefix) && strings.HasSuffix(unitName, ".service") {
			return strings.TrimSuffix(strings.TrimPrefix(unitName, prefix), ".service"), true
		}
	}
	return "", false
}

// resolveVPNUnit finds the systemd unit name for a VPN connection, trying
// each known OpenVPN naming convention in turn.
func (m *Manager) resolveVPNUnit(ctx context.Context, name string) (string, error) {
	for _, prefix := range openvpnUnitPrefixes {
		unitName := prefix + name + ".service"
		properties, err := m.conn.GetUnitPropertiesContext(ctx, unitName)
		if err != nil {
			continue
		}
		if loadState, ok := properties["LoadState"].(string); ok && loadState == "loaded" {
			return unitName, nil
		}
	}
	return "", fmt.Errorf("vpn %q not found", name)
}

// ListVPNs returns all OpenVPN client services
func (m *Manager) ListVPNs() ([]VPNStatus, error) {
	ctx := context.Background()
	units, err := m.conn.ListUnitsContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list units: %w", err)
	}

	vpns := make([]VPNStatus, 0, len(units))
	for _, unit := range units {
		name, ok := vpnNameFromUnit(unit.Name)
		if !ok {
			continue
		}

		// Get enabled status
		enabled := false
		unitFiles, err := m.conn.ListUnitFilesByPatternsContext(ctx, nil, []string{unit.Name})
		if err == nil && len(unitFiles) > 0 {
			enabled = unitFiles[0].Type == "enabled"
		}

		// Try to get tunnel IP for active VPNs
		tunnelIP := ""
		if unit.ActiveState == "active" {
			// Try common tunnel interface names (tun0, tun1, etc.)
			for i := 0; i < 10; i++ {
				ifname := fmt.Sprintf("tun%d", i)
				if ip, err := getTunnelIP(ifname); err == nil && ip != "" {
					tunnelIP = ip
					break
				}
			}
		}

		vpns = append(vpns, VPNStatus{
			Name:     name,
			Status:   unit.ActiveState,
			Enabled:  enabled,
			TunnelIP: tunnelIP,
		})
	}

	log.Debug("Found %d VPN(s)", len(vpns))
	return vpns, nil
}

// GetVPNStatus returns status for a specific VPN
func (m *Manager) GetVPNStatus(name string) (*VPNStatus, error) {
	ctx := context.Background()
	unitName, err := m.resolveVPNUnit(ctx, name)
	if err != nil {
		return nil, err
	}

	// Get unit properties
	properties, err := m.conn.GetUnitPropertiesContext(ctx, unitName)
	if err != nil {
		return nil, fmt.Errorf("failed to get unit properties: %w", err)
	}

	activeState, ok := properties["ActiveState"].(string)
	if !ok {
		return nil, errors.New("failed to get ActiveState")
	}

	// Get enabled status
	enabled := false
	unitFiles, err := m.conn.ListUnitFilesByPatternsContext(ctx, nil, []string{unitName})
	if err == nil && len(unitFiles) > 0 {
		enabled = unitFiles[0].Type == "enabled"
	}

	// Try to get tunnel IP for active VPNs
	tunnelIP := ""
	if activeState == "active" {
		for i := 0; i < 10; i++ {
			ifname := fmt.Sprintf("tun%d", i)
			if ip, err := getTunnelIP(ifname); err == nil && ip != "" {
				tunnelIP = ip
				break
			}
		}
	}

	status := &VPNStatus{
		Name:     name,
		Status:   activeState,
		Enabled:  enabled,
		TunnelIP: tunnelIP,
	}

	log.Debug("VPN status for %s: %s (enabled: %v)", name, activeState, enabled)
	return status, nil
}

// StartVPN starts the specified VPN connection
func (m *Manager) StartVPN(name string) error {
	ctx := context.Background()
	unitName, err := m.resolveVPNUnit(ctx, name)
	if err != nil {
		return err
	}

	log.Info("Starting VPN: %s", name)
	responseChan := make(chan string)
	_, err = m.conn.StartUnitContext(ctx, unitName, "replace", responseChan)
	if err != nil {
		return fmt.Errorf("failed to start VPN %s: %w", name, err)
	}

	// Wait for response
	status := <-responseChan
	log.Debug("VPN start response: %s", status)

	if status != "done" {
		return fmt.Errorf("failed to start VPN %s: status=%s", name, status)
	}

	log.Info("VPN %s started successfully", name)
	return nil
}

// StopVPN stops the specified VPN connection
func (m *Manager) StopVPN(name string) error {
	ctx := context.Background()
	unitName, err := m.resolveVPNUnit(ctx, name)
	if err != nil {
		return err
	}

	log.Info("Stopping VPN: %s", name)
	responseChan := make(chan string)
	_, err = m.conn.StopUnitContext(ctx, unitName, "replace", responseChan)
	if err != nil {
		return fmt.Errorf("failed to stop VPN %s: %w", name, err)
	}

	// Wait for response
	status := <-responseChan
	log.Debug("VPN stop response: %s", status)

	if status != "done" {
		return fmt.Errorf("failed to stop VPN %s: status=%s", name, status)
	}

	log.Info("VPN %s stopped successfully", name)
	return nil
}

// getTunnelIP extracts the IP address from a tunnel interface
func getTunnelIP(interfaceName string) (string, error) {
	iface, err := net.InterfaceByName(interfaceName)
	if err != nil {
		return "", err
	}

	addrs, err := iface.Addrs()
	if err != nil {
		return "", err
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}

	return "", fmt.Errorf("no IPv4 address found on %s", interfaceName)
}

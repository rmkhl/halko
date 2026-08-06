package dbus

import "testing"

func TestVPNNameFromUnit(t *testing.T) {
	tests := []struct {
		name     string
		unit     string
		wantName string
		wantOK   bool
	}{
		{"client convention", "openvpn-client@kiln.service", "kiln", true},
		{"legacy convention", "openvpn@kiln.service", "kiln", true},
		{"name with a dash", "openvpn-client@home-vpn.service", "home-vpn", true},
		{"name with a dot", "openvpn-client@vpn.example.service", "vpn.example", true},
		{"not openvpn", "sshd.service", "", false},
		{"right prefix, no .service", "openvpn-client@kiln", "", false},
		{"prefix only", "openvpn-client@.service", "", true},
		{"empty", "", "", false},
		{"similar but different unit", "openvpn-server@kiln.service", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotName, gotOK := vpnNameFromUnit(tt.unit)
			if gotOK != tt.wantOK {
				t.Fatalf("expected ok=%v, got ok=%v", tt.wantOK, gotOK)
			}
			if gotName != tt.wantName {
				t.Fatalf("expected name %q, got %q", tt.wantName, gotName)
			}
		})
	}
}

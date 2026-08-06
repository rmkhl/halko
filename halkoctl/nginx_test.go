package main

import (
	"strings"
	"testing"
)

const localhost = "localhost"

func TestExtractHostPort(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		wantHost string
		wantPort string
	}{
		{"host and port", "http://localhost:8088", localhost, "8088"},
		{"no port defaults to 80", "http://kiln.local", "kiln.local", "80"},
		{"ipv4 literal", "http://192.168.1.10:8092", "192.168.1.10", "8092"},
		{"ipv6 literal keeps its brackets", "http://[::1]:8092", "[::1]", "8092"},
		{"https defaults to 443", "https://kiln.local", "kiln.local", "443"},
		{"trailing slash", "http://localhost:8088/", localhost, "8088"},
		// A path must not leak into the port, or the generated upstream is
		// unusable.
		{"path is ignored", "http://localhost:8088/api/v1", localhost, "8088"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			host, port, err := extractHostPort(tt.url)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if host != tt.wantHost {
				t.Fatalf("expected host %q, got %q", tt.wantHost, host)
			}
			if port != tt.wantPort {
				t.Fatalf("expected port %q, got %q", tt.wantPort, port)
			}
		})
	}
}

func TestExtractHostPortRejectsBadInput(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{"empty", ""},
		{"no scheme", "localhost:8088"},
		{"unsupported scheme", "ftp://localhost:8088"},
		{"no host", "http://"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, _, err := extractHostPort(tt.url); err == nil {
				t.Fatalf("expected an error for %q", tt.url)
			}
		})
	}
}

func TestGenerateNginxConfigNamesEveryUpstream(t *testing.T) {
	config := generateNginxConfig(80,
		"cu-host", "8088",
		"sensor-host", "8090",
		"power-host", "8092",
		"dbus-host", "8094")

	for _, want := range []string{
		"cu-host:8088",
		"sensor-host:8090",
		"power-host:8092",
		"dbus-host:8094",
		"listen 80",
	} {
		if !strings.Contains(config, want) {
			t.Fatalf("expected the config to mention %q, got:\n%s", want, config)
		}
	}
}

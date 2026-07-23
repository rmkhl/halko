package heartbeat

import "testing"

func TestBuildDisplayRequestSendsMessageAndAddress(t *testing.T) {
	req := buildDisplayRequest("Drying 2:15", "192.168.1.42")
	if req.Message != "Drying 2:15" {
		t.Fatalf("expected message %q, got %q", "Drying 2:15", req.Message)
	}
	if req.Address != "192.168.1.42" {
		t.Fatalf("expected address %q, got %q", "192.168.1.42", req.Address)
	}
}

func TestBuildDisplayRequestDefaultsEmptyMessageToIdle(t *testing.T) {
	req := buildDisplayRequest("", "192.168.1.42")
	if req.Message != "idle" {
		t.Fatalf("expected message %q, got %q", "idle", req.Message)
	}
}

func TestBuildDisplayRequestOmitsUnknownAddress(t *testing.T) {
	req := buildDisplayRequest("idle", "")
	if req.Address != "" {
		t.Fatalf("expected empty address, got %q", req.Address)
	}
}

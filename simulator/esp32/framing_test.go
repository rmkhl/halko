package esp32

import (
	"strings"
	"testing"
)

// heloReply is the firmware's answer to `helo;`, repeated across these tests.
const heloCommand = "helo"

func TestFeedReturnsCompleteCommands(t *testing.T) {
	var b commandBuffer

	got := b.Feed([]byte("helo;"))
	if len(got) != 1 || got[0] != heloCommand {
		t.Fatalf("expected [helo], got %v", got)
	}
}

func TestFeedReturnsEveryCommandInOneChunk(t *testing.T) {
	var b commandBuffer

	got := b.Feed([]byte("helo;read;"))
	if len(got) != 2 || got[0] != heloCommand || got[1] != "read" {
		t.Fatalf("expected [helo read], got %v", got)
	}
}

func TestFeedReassemblesCommandSplitAcrossChunks(t *testing.T) {
	var b commandBuffer

	if got := b.Feed([]byte("re")); len(got) != 0 {
		t.Fatalf("expected no complete command yet, got %v", got)
	}
	if got := b.Feed([]byte("ad")); len(got) != 0 {
		t.Fatalf("expected no complete command yet, got %v", got)
	}
	got := b.Feed([]byte(";"))
	if len(got) != 1 || got[0] != "read" {
		t.Fatalf("expected [read], got %v", got)
	}
}

func TestFeedIgnoresLineEndings(t *testing.T) {
	var b commandBuffer

	got := b.Feed([]byte("\r\nhelo;\r\n"))
	if len(got) != 1 || got[0] != heloCommand {
		t.Fatalf("expected [helo], got %v", got)
	}
}

func TestFeedKeepsCommandArguments(t *testing.T) {
	var b commandBuffer

	got := b.Feed([]byte("show Pre-Heat;"))
	if len(got) != 1 || got[0] != "show Pre-Heat" {
		t.Fatalf("expected [show Pre-Heat], got %v", got)
	}
}

func TestFeedTruncatesOverlongCommands(t *testing.T) {
	var b commandBuffer

	long := strings.Repeat("x", maxCommandLength+8)
	got := b.Feed([]byte(long + ";"))
	if len(got) != 1 {
		t.Fatalf("expected one command, got %v", got)
	}
	if len(got[0]) != maxCommandLength {
		t.Fatalf("expected the command truncated to %d characters, got %d", maxCommandLength, len(got[0]))
	}
}

func TestFeedRecoversAfterOverflow(t *testing.T) {
	var b commandBuffer

	b.Feed([]byte(strings.Repeat("x", maxCommandLength+8) + ";"))

	// The buffer must be empty again, so the next command is unaffected.
	got := b.Feed([]byte("helo;"))
	if len(got) != 1 || got[0] != heloCommand {
		t.Fatalf("expected [helo] after an overflowing command, got %v", got)
	}
}

func TestFeedYieldsEmptyCommandForBareSemicolon(t *testing.T) {
	var b commandBuffer

	got := b.Feed([]byte(";"))
	if len(got) != 1 || got[0] != "" {
		t.Fatalf("expected one empty command, got %v", got)
	}
}

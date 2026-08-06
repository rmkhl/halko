package esp32

import (
	"testing"
	"time"

	"github.com/rmkhl/halko/simulator/elements"
	"github.com/rmkhl/halko/simulator/faults"
)

// recordingDisplay captures the display texts the responder forwards. It is
// a real observer, not a mock: the responder's only contract with the
// display side is that it forwards the text.
type recordingDisplay struct {
	messages []string
}

func (d *recordingDisplay) OnDisplayMessage(text string) {
	d.messages = append(d.messages, text)
}

// newTestResponder builds a responder over real elements sitting at 20°C.
func newTestResponder(injector *faults.Injector) (*Responder, *recordingDisplay) {
	wood := elements.NewWood(20.0, 20.0)
	heater := elements.NewHeater("kiln", 20.0, 20.0, wood)
	display := &recordingDisplay{}

	probes := []Probe{
		{Name: "KilnPrimary", Sensor: heater},
		{Name: "KilnSecondary", Sensor: heater},
		{Name: "Wood", Sensor: wood},
	}
	return NewResponder(probes, injector, display), display
}

func TestRespondHelo(t *testing.T) {
	r, _ := newTestResponder(faults.New(false))

	if got := string(r.Respond("helo", time.Now())); got != "helo\r\n" {
		t.Fatalf("expected %q, got %q", "helo\r\n", got)
	}
}

func TestRespondReadMatchesTheFirmwareFormat(t *testing.T) {
	r, _ := newTestResponder(faults.New(false))

	want := "KilnPrimary=20.00C,KilnSecondary=20.00C,Wood=20.00C\r\n"
	if got := string(r.Respond(readCommand, time.Now())); got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestRespondReadReportsFailedProbesAsNaN(t *testing.T) {
	injector := faults.New(true)
	r, _ := newTestResponder(injector)

	// Arm the schedule, then read past the final stage, where every probe
	// has failed.
	start := time.Now()
	injector.Observe(90.0, 20.0, start)

	want := "KilnPrimary=NaN,KilnSecondary=NaN,Wood=NaN\r\n"
	if got := string(r.Respond(readCommand, start.Add(5*time.Minute))); got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestRespondShowForwardsTheDisplayTextAndStaysSilent(t *testing.T) {
	r, display := newTestResponder(faults.New(false))

	if got := r.Respond("show Pre-Heat", time.Now()); got != nil {
		t.Fatalf("expected no response to show, got %q", got)
	}
	if len(display.messages) != 1 || display.messages[0] != "Pre-Heat" {
		t.Fatalf("expected the display text forwarded, got %v", display.messages)
	}
}

func TestRespondShowWithoutTextForwardsAnEmptyString(t *testing.T) {
	r, display := newTestResponder(faults.New(false))

	r.Respond("show", time.Now())
	if len(display.messages) != 1 || display.messages[0] != "" {
		t.Fatalf("expected one empty display text, got %v", display.messages)
	}
}

func TestRespondAddrStaysSilentAndDoesNotTouchTheDisplay(t *testing.T) {
	r, display := newTestResponder(faults.New(false))

	if got := r.Respond("addr 192.168.1.10", time.Now()); got != nil {
		t.Fatalf("expected no response to addr, got %q", got)
	}
	if len(display.messages) != 0 {
		t.Fatalf("expected addr not to reach the display observer, got %v", display.messages)
	}
}

func TestRespondIgnoresUnknownCommands(t *testing.T) {
	r, _ := newTestResponder(faults.New(false))

	if got := r.Respond("wibble", time.Now()); got != nil {
		t.Fatalf("expected no response to an unknown command, got %q", got)
	}
}

package esp32

import (
	"fmt"
	"strings"
	"time"

	"github.com/rmkhl/halko/simulator/engine"
	"github.com/rmkhl/halko/simulator/faults"
	"github.com/rmkhl/halko/types"
	"github.com/rmkhl/halko/types/log"
)

// readCommand is the firmware's temperature-report request. It is a named
// constant, rather than a literal repeated across this package's tests,
// because golangci-lint's goconst check flags three-plus repeats.
const readCommand = "read"

// Probe pairs a firmware sensor name with the simulated element it reads.
// Both kiln probes read the heater: the simulation has one kiln temperature,
// and inventing a difference between the two would be simulating a problem we
// do not have.
type Probe struct {
	Name   string
	Sensor engine.TemperatureSensor
}

// DisplayObserver receives the text of each `show` command, which is how the
// simulator learns that a program has started or stopped.
type DisplayObserver interface {
	OnDisplayMessage(text string)
}

// Responder answers the commands the sensor unit sends, as
// sensorunit.ino:276-330 does.
type Responder struct {
	probes  []Probe
	faults  *faults.Injector
	display DisplayObserver
}

// NewResponder returns a responder over the given probes, which are reported
// in the order supplied.
func NewResponder(probes []Probe, injector *faults.Injector, display DisplayObserver) *Responder {
	return &Responder{probes: probes, faults: injector, display: display}
}

// Respond returns the bytes to write back for one command, or nil where the
// firmware stays silent.
func (r *Responder) Respond(command string, now time.Time) []byte {
	verb, argument, _ := strings.Cut(command, " ")

	switch verb {
	case "helo":
		return []byte("helo\r\n")
	case readCommand:
		return r.readLine(now)
	case "show":
		r.display.OnDisplayMessage(argument)
		return nil
	case "addr":
		log.Trace("Display address line set to %q", argument)
		return nil
	default:
		log.Warning("Ignoring unknown command from sensor unit: %q", command)
		return nil
	}
}

// readLine formats one temperature report, applying any injected failures
// first so a failed probe reports NaN exactly as the firmware does.
func (r *Responder) readLine(now time.Time) []byte {
	values := make(types.TemperatureResponse, len(r.probes))
	for _, probe := range r.probes {
		values[probe.Name] = probe.Sensor.Temperature()
	}
	r.faults.Apply(values, now)

	var line strings.Builder
	for n, probe := range r.probes {
		if n > 0 {
			line.WriteByte(',')
		}
		line.WriteString(probe.Name)
		line.WriteByte('=')
		if values[probe.Name] == types.InvalidTemperatureReading {
			line.WriteString("NaN")
		} else {
			fmt.Fprintf(&line, "%.2fC", values[probe.Name])
		}
	}
	line.WriteString("\r\n")

	return []byte(line.String())
}

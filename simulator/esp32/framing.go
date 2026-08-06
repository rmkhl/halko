// Package esp32 emulates the Halko sensor unit's ESP32 firmware over a
// pseudo-terminal, so the real sensorunit service runs against the simulator
// exactly as it runs against the device.
package esp32

// maxCommandLength matches MAX_COMMAND_LENGTH in
// sensorunit/esp32/sensorunit/sensorunit.ino:130.
const maxCommandLength = 32

// commandBuffer reassembles `;`-terminated commands from a byte stream, as
// processSerial does in sensorunit.ino:248-274. The zero value is ready to
// use.
type commandBuffer struct {
	pending []byte
}

// Feed appends a chunk of input and returns every command it completed, in
// order and without the terminating semicolon. Carriage returns and newlines
// are ignored, and a command longer than maxCommandLength is truncated to
// that many characters — the firmware stops advancing its write index and
// then overwrites the final slot with the terminating NUL.
func (b *commandBuffer) Feed(chunk []byte) []string {
	var commands []string

	for _, c := range chunk {
		switch c {
		case '\r', '\n':
			continue
		case ';':
			commands = append(commands, string(b.pending))
			b.pending = b.pending[:0]
		default:
			if len(b.pending) < maxCommandLength {
				b.pending = append(b.pending, c)
			}
		}
	}

	return commands
}

package engine

import (
	"testing"
	"time"
)

// The run loop asks for a sample every tick, so the request must never block
// when the reader is still busy with the previous one.
func TestRequestSensorReadDoesNotBlockWhenReaderIsBusy(t *testing.T) {
	commands := make(chan string) // unbuffered, nobody receiving
	done := make(chan bool, 1)

	go func() { done <- requestSensorRead(commands) }()

	select {
	case sent := <-done:
		if sent {
			t.Fatal("requestSensorRead() = true with no reader listening, want false")
		}
	case <-time.After(time.Second):
		t.Fatal("requestSensorRead() blocked while no reader was listening")
	}
}

func TestRequestSensorReadDeliversWhenReaderIsListening(t *testing.T) {
	commands := make(chan string)
	got := make(chan string, 1)

	go func() { got <- <-commands }()

	// A non-blocking send only succeeds once the reader is parked on the
	// channel, so retry until the goroutine gets there.
	deadline := time.Now().Add(2 * time.Second)
	for !requestSensorRead(commands) {
		if time.Now().After(deadline) {
			t.Fatal("requestSensorRead() never delivered to a listening reader")
		}
		time.Sleep(time.Millisecond)
	}

	select {
	case msg := <-got:
		if msg != sensorRead {
			t.Errorf("reader received %q, want %q", msg, sensorRead)
		}
	case <-time.After(time.Second):
		t.Fatal("reader did not receive the command")
	}
}

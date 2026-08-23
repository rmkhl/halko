package engine

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/rmkhl/halko/types"
)

// waitForReader fails the test unless the reader goroutine has returned. A
// leaked reader keeps the runner's WaitGroup up forever, which is what pins the
// engine to a program that already ended.
func waitForReader(t *testing.T, wg *sync.WaitGroup, what string) {
	t.Helper()

	stopped := make(chan struct{})
	go func() {
		wg.Wait()
		close(stopped)
	}()

	select {
	case <-stopped:
	case <-time.After(5 * time.Second):
		t.Fatalf("%s: reader still running after shutdown was signalled", what)
	}
}

func temperatureServer(t *testing.T, gate <-chan struct{}) *httptest.Server {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if gate != nil {
			<-gate
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"kiln_primary":28.5,"kiln_secondary":29.0,"material":22.25}}`))
	}))
	t.Cleanup(server.Close)
	return server
}

// Each power device must be read from its own key in the power unit response.
// Distinct percentages catch a reading that is sourced from the wrong device.
func TestReadSensorsMapsEachDeviceToItsOwnReading(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"fan":{"percent":100},"heater":{"percent":0},"steam":{"percent":50}}}`))
	}))
	defer server.Close()

	reader := psuSensorReader{
		sensorReader: sensorReader{client: server.Client(), sensorURL: server.URL},
	}

	readings, err := reader.readSensors()
	if err != nil {
		t.Fatalf("readSensors() returned error: %v", err)
	}

	if readings.Fan.Percent != 100 {
		t.Errorf("Fan.Percent = %d, want 100", readings.Fan.Percent)
	}
	if readings.Heater.Percent != 0 {
		t.Errorf("Heater.Percent = %d, want 0", readings.Heater.Percent)
	}
	if readings.Steam.Percent != 50 {
		t.Errorf("Steam.Percent = %d, want 50", readings.Steam.Percent)
	}
}

func TestReadTemperaturesMapsBothKilnSensorsAndMaterial(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"kiln_primary":28.5,"kiln_secondary":29.0,"material":22.25}}`))
	}))
	defer server.Close()

	reader := temperatureSensorReader{
		sensorReader: sensorReader{client: server.Client(), sensorURL: server.URL},
	}

	readings, err := reader.readTemperatures()
	if err != nil {
		t.Fatalf("readTemperatures() returned error: %v", err)
	}

	// The reader reports what the sensors said and resolves nothing; Kiln is
	// filled in later, by observe.
	if readings.KilnPrimary != 28.5 {
		t.Errorf("KilnPrimary = %v, want 28.5", readings.KilnPrimary)
	}
	if readings.KilnSecondary != 29.0 {
		t.Errorf("KilnSecondary = %v, want 29.0", readings.KilnSecondary)
	}
	if readings.Material != 22.25 {
		t.Errorf("Material = %v, want 22.25", readings.Material)
	}
}

// A reader parked on its command channel must notice the shutdown signal.
func TestTemperatureReaderStopsWhileWaitingForACommand(t *testing.T) {
	server := temperatureServer(t, nil)

	shutdown := make(chan struct{})
	reader, err := newTemperatureSensorReader(server.URL, temperatureServer(t, nil).URL, make(chan string), make(chan temperatureReadings), shutdown)
	if err != nil {
		t.Fatalf("newTemperatureSensorReader() returned error: %v", err)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go reader.Run(&wg)

	close(shutdown)
	waitForReader(t, &wg, "parked on commands")
}

// The run loop stops receiving responses once it has left its loop, so a reader
// holding a fresh reading has nobody to hand it to. It must abandon the send
// instead of blocking on it forever.
func TestTemperatureReaderStopsWhileBlockedPublishingAReading(t *testing.T) {
	server := temperatureServer(t, nil)

	commands := make(chan string)
	shutdown := make(chan struct{})
	// Unbuffered and never received from: the runner has already gone away.
	reader, err := newTemperatureSensorReader(server.URL, temperatureServer(t, nil).URL, commands, make(chan temperatureReadings), shutdown)
	if err != nil {
		t.Fatalf("newTemperatureSensorReader() returned error: %v", err)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go reader.Run(&wg)

	commands <- sensorRead
	close(shutdown)
	waitForReader(t, &wg, "blocked publishing a reading")
}

// This is the shape of the real failure: the failsafe trips while the reader is
// still inside a slow read of a sensor unit that has stopped answering.
func TestTemperatureReaderStopsWhenShutdownArrivesDuringARead(t *testing.T) {
	// One token lets the constructor's verification read through; the read the
	// test triggers then blocks until the gate is closed.
	gate := make(chan struct{}, 1)
	gate <- struct{}{}
	server := temperatureServer(t, gate)

	commands := make(chan string)
	shutdown := make(chan struct{})
	reader, err := newTemperatureSensorReader(server.URL, temperatureServer(t, nil).URL, commands, make(chan temperatureReadings), shutdown)
	if err != nil {
		t.Fatalf("newTemperatureSensorReader() returned error: %v", err)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go reader.Run(&wg)

	commands <- sensorRead
	close(shutdown)
	close(gate) // let the in-flight read complete after shutdown
	waitForReader(t, &wg, "shut down during a read")
}

func TestPSUReaderStopsWhileBlockedPublishingAReading(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"fan":{"percent":100},"heater":{"percent":0},"steam":{"percent":50}}}`))
	}))
	defer server.Close()

	commands := make(chan string)
	shutdown := make(chan struct{})
	reader, err := newPSUSensorReader(server.URL, commands, make(chan psuReadings), shutdown)
	if err != nil {
		t.Fatalf("newPSUSensorReader() returned error: %v", err)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go reader.Run(&wg)

	commands <- sensorRead
	close(shutdown)
	waitForReader(t, &wg, "blocked publishing a reading")
}

// The cold junctions come from their own endpoint, so the reader has to go
// and get them rather than find them in the temperature response.
func TestReadTemperaturesFetchesDieReadingsFromTheirOwnEndpoint(t *testing.T) {
	dieServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"material_die":41.875,"kiln_primary_die":27.5,"kiln_secondary_die":28.125}}`))
	}))
	defer dieServer.Close()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"kiln_primary":28.5,"kiln_secondary":29.0,"material":22.25}}`))
	}))
	defer server.Close()

	reader := temperatureSensorReader{
		sensorReader: sensorReader{client: server.Client(), sensorURL: server.URL},
		dieURL:       dieServer.URL,
	}

	readings, err := reader.readTemperatures()
	if err != nil {
		t.Fatalf("readTemperatures() returned error: %v", err)
	}

	if readings.MaterialDie != 41.875 {
		t.Errorf("MaterialDie = %v, want 41.875", readings.MaterialDie)
	}
	if readings.KilnPrimaryDie != 27.5 {
		t.Errorf("KilnPrimaryDie = %v, want 27.5", readings.KilnPrimaryDie)
	}
	if readings.KilnSecondaryDie != 28.125 {
		t.Errorf("KilnSecondaryDie = %v, want 28.125", readings.KilnSecondaryDie)
	}
}

// Losing the diagnostic endpoint must cost the diagnostic columns and nothing
// else: the run depends on the kiln and material readings in the same call.
func TestReadTemperaturesSurvivesAFailingDieEndpoint(t *testing.T) {
	dieServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer dieServer.Close()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"kiln_primary":28.5,"kiln_secondary":29.0,"material":22.25}}`))
	}))
	defer server.Close()

	reader := temperatureSensorReader{
		sensorReader: sensorReader{client: server.Client(), sensorURL: server.URL},
		dieURL:       dieServer.URL,
	}

	readings, err := reader.readTemperatures()
	if err != nil {
		t.Fatalf("readTemperatures() returned error: %v", err)
	}

	if readings.KilnPrimary != 28.5 || readings.KilnSecondary != 29.0 || readings.Material != 22.25 {
		t.Errorf("kiln primary/secondary/material = %v/%v/%v, want 28.5/29.0/22.25",
			readings.KilnPrimary, readings.KilnSecondary, readings.Material)
	}
	if readings.MaterialDie != types.InvalidTemperatureReading {
		t.Errorf("MaterialDie = %v, want the invalid sentinel", readings.MaterialDie)
	}
}

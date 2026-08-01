package engine

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// Each power device must be read from its own key in the power unit response.
// Distinct percentages catch a reading that is sourced from the wrong device.
func TestReadSensorsMapsEachDeviceToItsOwnReading(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"fan":{"percent":100},"heater":{"percent":0},"humidifier":{"percent":50}}}`))
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
	if readings.Humidifier.Percent != 50 {
		t.Errorf("Humidifier.Percent = %d, want 50", readings.Humidifier.Percent)
	}
}

func TestReadTemperaturesMapsKilnAndMaterial(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"kiln":28.5,"material":22.25}}`))
	}))
	defer server.Close()

	reader := temperatureSensorReader{
		sensorReader: sensorReader{client: server.Client(), sensorURL: server.URL},
	}

	readings, err := reader.readTemperatures()
	if err != nil {
		t.Fatalf("readTemperatures() returned error: %v", err)
	}

	if readings.Kiln != 28.5 {
		t.Errorf("Kiln = %v, want 28.5", readings.Kiln)
	}
	if readings.Material != 22.25 {
		t.Errorf("Material = %v, want 22.25", readings.Material)
	}
}

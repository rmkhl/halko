# SensorUnit

The SensorUnit is a critical component of the Halko wood drying kiln control
system that provides temperature monitoring capabilities. It consists of
ESP32-based hardware for physical sensor readings and a Go service that
exposes the data via a REST API.

## Overview

The system includes an ESP32-based sensor unit for temperature monitoring in
the kiln and a service that provides a REST API for integration with the
controlunit component.

## Hardware Components

The sensor unit is based on an ESP32 (NodeMCU-32S) and uses MAX31855
thermocouple amplifiers to measure temperatures from three probes (two kiln
sensors and one wood sensor). It displays readings and status messages on an
I2C OLED display (SSD1306 or SH1106).

See [esp32/README.md](esp32/README.md) for board details and
[esp32/WIRING-ESP32.md](esp32/WIRING-ESP32.md) for the wiring guide.

## ESP32 Firmware

The firmware for the sensor unit is located at
`esp32/sensorunit/sensorunit.ino`. This firmware handles:

- Reading from the MAX31855 thermocouple sensors (with moving-average
  smoothing and validity tracking)
- Displaying temperature readings and status on the OLED display
- Responding to serial commands from the Go service
- Managing connection status with visual indicators

### Building and Uploading Firmware

The firmware is compiled and uploaded with Arduino CLI via Make targets from
the project root:

```bash
## Setup (one-time)
make prepare-esp32        # Installs Arduino CLI, ESP32 core, and libraries

## Build firmware
make build-esp32          # Compiles firmware to firmware-esp32/

## Upload to the ESP32
make upload-esp32                      # Uploads to /dev/ttyUSB0 (default)
make upload-esp32 PORT=/dev/ttyUSB1    # Custom port

## Serial monitor
make monitor-esp32        # Connect to the ESP32 serial port

## Cleanup
make clean-esp32          # Remove firmware build artifacts
```

Run `make esp32-help` for full usage information.

**Technical details**:

- Board: NodeMCU-32S (ESP32-WROOM-32), board identifier `esp32:esp32:nodemcu-32s`
- USB-to-serial: CP2102
- Libraries: Adafruit MAX31855, Adafruit GFX, Adafruit SSD1306
- WiFi and Bluetooth are disabled by the firmware for simplicity and power
  savings

## Serial Commands

The Go service communicates with the ESP32 at 9600 baud. The unit accepts
the following commands over the serial interface:

- `helo;` - Initial handshake, responds with "helo"
- `read;` - Request temperature readings, returns values in format:
  `KilnPrimary=XX.XC,KilnSecondary=XX.XC,Wood=XX.XC`
  (a sensor with no valid reading reports `NaN` instead of a value)
- `show TEXT;` - Updates the status text on the OLED display

## Connection Status

The sensor unit's service provides an endpoint to check the connection status
and another to update the status message shown on the OLED display. The
service handles sensor unplug/replug by reconnecting to the serial device
automatically.

## Integration

The ControlUnit component communicates with the SensorUnit's REST API to
retrieve temperature data during drying programs. The SensorUnit continues to
display temperatures locally even when disconnected from the main system. The
service part of the SensorUnit handles the serial communication with the
ESP32 and exposes the data via HTTP.

## Configuration

The SensorUnit reads its settings from the shared Halko configuration file
(`/etc/opt/halko.cfg` in production). The relevant sections are:

```json
{
  "sensorunit": {
    "serial_device": "/dev/ttyUSB0",
    "baud_rate": 9600
  },
  "api_endpoints": {
    "sensorunit": {
      "url": "http://localhost:8093",
      "temperatures": "/temperatures",
      "display": "/display",
      "status": "/status"
    }
  }
}
```

## Systemd Service

The SensorUnit runs under the templated Halko service unit installed by
`make systemd-units`. It can be controlled with:

```bash
sudo systemctl enable --now halko@sensorunit
```

## API Endpoints

For detailed API documentation, see the main [API.md](../API.md) file. The
SensorUnit provides endpoints for:

- Temperature readings from all sensors
- Connection status checking
- OLED status message updates

## Building

The SensorUnit Go service is built as part of the main project build process:

```bash
# From the project root
make all
```

This will create the `sensorunit` binary in the `bin/` directory.

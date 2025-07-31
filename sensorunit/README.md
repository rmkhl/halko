# SensorUnit

The SensorUnit is a critical component of the Halko wood drying kiln control system that provides temperature monitoring capabilities. It consists of Arduino-based hardware for physical sensor readings and a Go service that exposes the data via a REST API.

## Overview

The system includes an Arduino-based sensor unit for temperature monitoring in the kiln and a service that provides a REST API for integration with the executor component.

## Hardware Components

The sensor unit is based on an Arduino and uses MAX6675 thermocouple sensors to measure temperatures. It can display status messages on an LCD.

## Arduino Firmware

The Arduino firmware for the sensor unit is located at `arduino/sensorunit/sensorunit.ino`. This firmware handles:

- Reading from the MAX6675 thermocouple sensors
- Displaying temperature readings and status on the LCD
- Responding to serial commands from the Go service
- Managing connection status with visual indicators

## Serial Commands

The unit accepts the following commands over the serial interface:

- `helo;` - Initial handshake, responds with "helo"
- `read;` - Request temperature readings, returns values in format: `OvenPrimary=XX.XC,OvenSecondary=XX.XC,Wood=XX.XC`
- `show TEXT;` - Updates the status text on the LCD display

## Connection Status

The sensor unit's service provides an endpoint to check the connection status and another to update the status message displayed on the LCD.

## Integration

The Executor component communicates with the SensorUnit's REST API to retrieve temperature data during drying programs. The SensorUnit continues to display temperatures locally even when disconnected from the main system. The service part of the SensorUnit handles the serial communication with the Arduino and exposes the data via HTTP.

## Configuration

The SensorUnit configuration is stored in a JSON file (`/etc/opt/halko/sensorunit.json`) and includes parameters like `SerialPort` and `BaudRate`.

Example configuration:

```json
{
  "sensorunit": {
    "serial_device": "/dev/ttyUSB0",
    "baud_rate": 9600
  }
}
```

## Systemd Service

The SensorUnit service file (`sensorunit.service`) is located in the `../templates` directory and installed to `/etc/systemd/system/`. It can be enabled and started with:

```bash
sudo systemctl enable --now sensorunit
```

## API Endpoints

For detailed API documentation, see the main [API.md](../API.md) file. The SensorUnit provides endpoints for:

- Temperature readings from all sensors
- Connection status checking
- LCD status message updates

## Building

The SensorUnit Go service is built as part of the main project build process:

```bash
# From the project root
make all
```

This will create the `sensorunit` binary in the `bin/` directory.

# SensorUnit

The SensorUnit is a critical component of the Halko wood drying kiln control
system that provides temperature monitoring capabilities. It consists of
Arduino-based hardware for physical sensor readings and a Go service that
exposes the data via a REST API.

## Overview

The system includes an Arduino-based sensor unit for temperature monitoring in
the kiln and a service that provides a REST API for integration with the
executor component.

## Hardware Components

The sensor unit is based on an Arduino and uses MAX6675 thermocouple sensors
to measure temperatures. It can display status messages on an LCD.

## Arduino Firmware

The Arduino firmware for the sensor unit is located at
`arduino/sensorunit/sensorunit.ino`. This firmware handles:

- Reading from the MAX6675 thermocouple sensors
- Displaying temperature readings and status on the LCD
- Responding to serial commands from the Go service
- Managing connection status with visual indicators

### Building and Uploading Firmware

The Arduino firmware can be compiled and uploaded using Make targets from the
project root:

```bash
## Setup (one-time)
make install-arduino-cli  # Installs Arduino CLI, AVR core, and required libraries

## Build firmware
make build-arduino        # Compiles firmware to firmware/sensorunit.ino.hex

## Upload to Arduino
make upload-arduino       # Uploads to /dev/ttyUSB0 (default)
make upload-arduino PORT=/dev/ttyUSB1  # Custom port

## Backup existing firmware
make backup-arduino       # Backs up current firmware from Arduino
make backup-arduino PORT=/dev/ttyUSB1  # Custom port

## Restore backed-up firmware
make restore-arduino BACKUP=firmware/backup/arduino_backup_20260308_143022.hex
make restore-arduino BACKUP=firmware/backup/arduino_backup_20260308_143022.hex PORT=/dev/ttyUSB1
```

**Backup functionality**: The `backup-arduino` target reads both flash memory
and EEPROM from the connected Arduino and saves timestamped files to
`firmware/backup/`:
- `arduino_backup_YYYYMMDD_HHMMSS.hex` - Flash memory (program code)
- `arduino_backup_YYYYMMDD_HHMMSS.eep` - EEPROM data

This is useful before uploading new firmware or for disaster recovery purposes.

**Restore functionality**: The `restore-arduino` target uploads a previously
backed-up .hex file to the Arduino. When called without parameters, it lists
available backups. The backup .eep (EEPROM) file is not automatically restored
and must be uploaded separately if needed using avrdude directly.

**Technical details**:
- Board: Arduino Nano
- Processor: ATmega328P
- Libraries: MAX6675 library 1.1.2 (Adafruit), LiquidCrystal 1.0.7
- Bootloader: Arduino/STK500v1 at 57600 baud

## Serial Commands

The unit accepts the following commands over the serial interface:

- `helo;` - Initial handshake, responds with "helo"
- `read;` - Request temperature readings, returns values in format:
  `OvenPrimary=XX.XC,OvenSecondary=XX.XC,Wood=XX.XC`
- `show TEXT;` - Updates the status text on the LCD display

## Connection Status

The sensor unit's service provides an endpoint to check the connection status
and another to update the status message displayed on the LCD.

## Integration

The Executor component communicates with the SensorUnit's REST API to retrieve
temperature data during drying programs. The SensorUnit continues to display
temperatures locally even when disconnected from the main system. The service
part of the SensorUnit handles the serial communication with the Arduino and
exposes the data via HTTP.

## Configuration

The SensorUnit configuration is stored in a JSON file
(`/etc/opt/halko/sensorunit.json`) and includes parameters like `SerialPort`
and `BaudRate`.

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

The SensorUnit service file (`sensorunit.service`) is located in the
`../templates` directory and installed to `/etc/systemd/system/`. It can be
enabled and started with:

```bash
sudo systemctl enable --now sensorunit
```

## API Endpoints

For detailed API documentation, see the main [API.md](../API.md) file. The
SensorUnit provides endpoints for:

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

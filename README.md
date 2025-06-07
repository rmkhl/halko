# Halko

Halko is a distributed system for controlling and monitoring wood drying kilns. It consists of multiple components that work together to provide temperature control, power management, and program execution capabilities.

## Overview

The system is built with a microservices architecture with these main components:

- **Configurator**: Manages program and phase configurations
- **Executor**: Runs drying programs and controls the kiln
- **PowerUnit**: Interfaces with Shelly devices to control power
- **Simulator**: Provides a simulated environment for testing
- **WebApp**: User interface for controlling and monitoring the system

## Building the Project

### Prerequisites

- Go 1.23 or higher
- Node.js and npm (for the webapp)
- golangci-lint (for linting)

### Build Commands

The project uses a Makefile to simplify building and deployment:

```bash
# Build all components
make all

# Clean and rebuild all components from scratch
make rebuild

# Remove all built binaries
make clean

# Run linter on all modules
make lint

# Update Go module dependencies
make update-modules

# Install binaries to /opt/halko and config to /etc/opt
make install

# Install and enable systemd service units
make systemd-units

# Reformat changed Go files
make fmt-changed
```

## Project Structure

### Components

#### `/configurator`

Storage service for program and phase configurations. Provides REST API for CRUD operations on configuration data, which is stored in the filesystem.

#### `/executor`

Core service that executes drying programs. It manages the state machine for program execution, controls power units, and monitors sensor data.

#### `/powerunit`

Interfaces with Shelly smart switches to control power to heaters, fans, and humidifiers. Provides a REST API for power control operations.

#### `/simulator`

Simulates the physical components of the kiln for development and testing purposes. Includes emulated temperature sensors and power controls.

#### `/webapp`

React-based frontend for the system. Allows users to create and modify programs, monitor active drying sessions, and control the system.

### Supporting Directories

#### `/bin`

Contains built executables for all components.

#### `/schemas`

JSON schemas for validation of program and phase data.

#### `/templates`

Contains systemd service templates and configuration samples.

#### `/tests`

Integration tests for the system components.

#### `/types`

Shared Go type definitions used across multiple components.

#### `/sensorunit`

Contains both Arduino code for the physical temperature sensor unit and a REST API webservice that interfaces with the Arduino over USB serial connection.

## Sensor Unit

The system includes an Arduino-based sensor unit for temperature monitoring in the kiln and a Go webservice that provides a REST API for integration with the executor component.

### Hardware Components

- **Controller**: Arduino Nano board
- **Temperature Sensors**: 3× MAX6675 thermocouples for measuring:
  - Primary oven temperature
  - Secondary oven temperature
  - Wood temperature
- **Display**: 16×2 LCD (LCM 1602C) for local temperature readings

### Functionality

The sensor unit performs the following functions:

- Reads temperatures from all three thermocouples every second
- Displays current temperatures on the LCD screen
- Provides temperature data over serial connection when requested
- Shows connection status on the LCD
- Accepts commands via serial port (9600 baud)

### Serial Commands

The unit accepts the following commands over the serial interface:

- `helo;` - Initial handshake, responds with "helo"
- `read;` - Request temperature readings, returns values in format: `OvenPrimary=XX.XC,OvenSecondary=XX.XC,Wood=XX.XC`
- `show TEXT;` - Updates the status text on the LCD display

### Connection Status

The LCD displays the current connection status:

- Shows custom status text when connected
- Automatically displays "Disconnected" after 30 seconds without commands
- Displays a visual indicator (alternating *,+) to show the unit is operational

### Integration

The Executor component communicates with the sensor unit to retrieve temperature data during drying programs using the REST API.
The sensor unit continues to display temperatures locally even when disconnected from the main system.

### REST API Web Service

The sensorunit directory includes a Go webservice that provides a REST API for interacting with the Arduino-based sensor unit:

#### API Endpoints

- `GET /api/temperature` - Fetch current temperature readings from all three sensors
- `GET /api/status` - Check connection status of the sensor unit
- `POST /api/status` - Update the status text displayed on the LCD screen

#### Configuration

The sensorunit webservice is configured through the main Halko configuration file (`halko.cfg`) with the following settings:

```json
"sensorunit": {
  "serial_device": "/dev/ttyUSB0",  // Path to the USB serial device
  "baud_rate": 9600                 // Baud rate for serial communication
}
```

The service automatically extracts the port to use from the `executor.sensor_unit_url` setting.

The executor component communicates with the sensorunit through the `sensor_unit_url` setting:

```json
"executor": {
  "sensor_unit_url": "http://localhost:8089"
}
```

#### Systemd Service

The sensor unit service runs as a systemd service like other Halko components. It includes support for graceful termination, properly handling the following signals:

- `SIGTERM` - For graceful shutdown (e.g., from systemctl stop)
- `SIGINT` - For manual interruption (e.g., Ctrl+C)
- `SIGHUP` - For service reload

The service implements proper connection cleanup, ensuring the serial port is properly closed before termination. This prevents issues with reconnecting to the Arduino after a restart.

## Deployment

The system components are designed to run as systemd services. After building, use `make install` to install the binaries and `make systemd-units` to set up the systemd services.

Each component can be controlled independently:

```bash
# Start a specific component
sudo systemctl start halko@configurator

# Stop a component
sudo systemctl stop halko@executor

# Check status
sudo systemctl status halko@powerunit
```

## Development

For development, you can run the simulator instead of connecting to real hardware:

```bash
./bin/simulator -l 8088
```

The webapp can be run in development mode from the `/webapp` directory:

```bash
cd webapp
npm install
npm start
```

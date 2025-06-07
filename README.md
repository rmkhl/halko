# Halko

Halko is a distributed system for controlling and monitoring wood drying kilns. It consists of multiple components that work together to provide temperature control, power management, and program execution capabilities.

## Overview

The system is built with a microservices architecture with these main components:

- **Configurator**: Manages program and phase configurations.
- **Executor**: Runs drying programs and controls the kiln.
- **PowerUnit**: Interfaces with Shelly devices to control power.
- **SensorUnit**: Reads temperature data from physical sensors and provides an API.
- **Simulator**: Provides a simulated environment for testing.
- **WebApp**: User interface for controlling and monitoring the system.

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

The Configurator is a storage service for program and phase configurations. It allows for CRUD operations on this data, which is stored in the filesystem. It exposes a REST API for these operations.

#### `/executor`

The Executor is the core service that executes drying programs. It manages the state machine for program execution, interacts with the PowerUnit to control heating elements, and with the SensorUnit (or Simulator) to monitor temperatures. It also provides a REST API to manage and monitor program execution.

#### `/powerunit`

The PowerUnit interfaces with Shelly smart switches to control power to heaters, fans, and humidifiers. It provides a REST API for direct power control operations.

#### `/sensorunit`

The SensorUnit component includes:

1. Arduino firmware (`sensorunit/arduino/sensorunit/sensorunit.ino`) for a physical unit that reads from MAX6675 thermocouples and can display status on an LCD.
2. A Go service (`sensorunit/main.go`) that communicates with the Arduino via USB serial and exposes a REST API for temperature and status.

#### `/simulator`

The Simulator emulates the physical components of the kiln, such as temperature sensors and Shelly power controls. This is useful for development and testing without requiring actual hardware. It mimics the REST APIs of the SensorUnit and parts of the PowerUnit (Shelly devices).

#### `/webapp`

A React-based frontend application that serves as the user interface for the Halko system. It allows users to create and modify drying programs, monitor active drying sessions, and control the overall system.

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

## Sensor Unit

The system includes an Arduino-based sensor unit for temperature monitoring in the kiln and a service that provides a REST API for integration with the executor component.

### Hardware Components

The sensor unit is based on an Arduino and uses MAX6675 thermocouple sensors to measure temperatures. It can display status messages on an LCD.

### Arduino Firmware

The Arduino firmware for the sensor unit is located at `/sensorunit/arduino/sensorunit/sensorunit.ino`. This firmware handles:

- Reading from the MAX6675 thermocouple sensors
- Displaying temperature readings and status on the LCD
- Responding to serial commands from the Go service
- Managing connection status with visual indicators

### Serial Commands

The unit accepts the following commands over the serial interface:

- `helo;` - Initial handshake, responds with "helo"
- `read;` - Request temperature readings, returns values in format: `OvenPrimary=XX.XC,OvenSecondary=XX.XC,Wood=XX.XC`
- `show TEXT;` - Updates the status text on the LCD display

### Connection Status

The sensor unit's service provides an endpoint to check the connection status and another to update the status message displayed on the LCD.

### Integration

The Executor component communicates with the SensorUnit's REST API to retrieve temperature data during drying programs. The SensorUnit continues to display temperatures locally even when disconnected from the main system. The service part of the SensorUnit handles the serial communication with the Arduino and exposes the data via HTTP.

#### Configuration

The SensorUnit configuration is stored in a JSON file (`/etc/opt/halko/sensorunit.json`) and includes parameters like `SerialPort` and `BaudRate`.

#### Systemd Service

The SensorUnit service file (`sensorunit.service`) is located in the `templates` directory and installed to `/etc/systemd/system/`. It can be enabled and started with:

```bash
sudo systemctl enable --now sensorunit
```

## API Endpoints

This section outlines the REST API endpoints provided by each module.

### Configurator (`/configurator`)

Base Path: `/api/v1`

- **Programs:**
  - `GET /programs`: List all programs.
  - `POST /programs`: Create a new program.
  - `GET /programs/:name`: Get a specific program by name.
  - `PUT /programs/:name`: Update a specific program by name.
- **Phases:**
  - `GET /phases`: List all phases.
  - `POST /phases`: Create a new phase.
  - `GET /phases/:name`: Get a specific phase by name.
  - `PUT /phases/:name`: Update a specific phase by name.

### Executor (`/executor`)

Base Path: `/engine/api/v1`

- **Program Storage:**
  - `GET /programs`: List all available programs (definitions loaded by executor).
  - `GET /programs/:name`: Get a specific program definition by name.
  - `DELETE /programs/:name`: Delete/unload a specific program definition.
- **Engine Control:**
  - `GET /running`: Get the status of the currently running program.
  - `POST /running`: Start a new program (by providing its definition or name).
  - `DELETE /running`: Cancel the currently running program.

### PowerUnit (`/powerunit`)

Base Path: `/powers/api/v1`

- **Powers:**
  - `GET /`: Get the status of all power channels.
  - `GET /:power`: Get the status of a specific power channel (e.g., `heater`, `fan`, `humidifier`).
  - `POST /:power`: Operate a specific power channel (turn on/off, set percentage). Also supports `PUT` and `PATCH` methods for the same operation.

### SensorUnit (`/sensorunit`)

Base Path: `/sensors/api/v1`

- **Temperature:**
  - `GET /temperature`: Fetch current temperature readings from all sensors.
- **Status:**
  - `GET /status`: Check the connection status of the sensor unit.
  - `POST /status`: Update the status text displayed on the sensor unit's LCD.
    - Body: `{"message": "your status text"}`

### Simulator (`/simulator`)

The simulator mimics endpoints from other services for testing purposes.

- **Simulated SensorUnit API:**
  Base Path: `/sensors/api/v1`
  - **Temperature:**
    - `GET /temperatures`: Get readings from all simulated temperature sensors.
    - `GET /temperatures/:sensor`: Get reading from a specific simulated sensor.
  - **Status:**
    - `GET /status`: Get the simulated connection status (always returns "connected").
    - `POST /status`: Log a status message (simulates updating an LCD).
      - Body: `{"message": "your status text"}`

- **Simulated Shelly Switch Control (RPC style):**
  Base Path: `/rpc`
  - `GET /Switch.GetStatus`: Get the status of simulated Shelly switches.
    - Query Params: `id=<switch_id>` (e.g., `id=0`)
  - `GET /Switch.Set`: Set the state of simulated Shelly switches.
    - Query Params: `id=<switch_id>&on=<true|false>`

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

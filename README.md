# Halko

Halko is a distributed system for controlling and monitoring wood drying kilns.
It consists of multiple components that work together to provide temperature
control, power management, and program execution capabilities.

## Overview

The system is built with a microservices architecture with these main components:

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

#### `/executor`

The Executor is the core service that executes drying programs. It manages the
state machine for program execution, interacts with the PowerUnit to control
heating elements, and with the SensorUnit (or Simulator) to monitor
temperatures. It also provides a REST API to manage and monitor program
execution.

The Executor includes a heartbeat service that periodically reports its IP
address to a configured status endpoint. This allows monitoring systems to
track the location and availability of the executor service in distributed
deployments.

#### `/powerunit`

The PowerUnit interfaces with Shelly smart switches to control power to
heaters, fans, and humidifiers. It provides a REST API for direct power
control operations.

#### `/sensorunit`

The SensorUnit component includes:

- Arduino firmware (`sensorunit/arduino/sensorunit/sensorunit.ino`) for a
  physical unit that reads from MAX6675 thermocouples and can display status
  on an LCD.
- A Go service (`sensorunit/main.go`) that communicates with the Arduino via
  USB serial and exposes a REST API for temperature and status.

#### `/simulator`

The Simulator emulates the physical components of the kiln, such as
temperature sensors and Shelly power controls. This is useful for development
and testing without requiring actual hardware. It mimics the REST APIs of the
SensorUnit and parts of the PowerUnit (Shelly devices).

#### `/webapp`

A React-based frontend application that serves as the user interface for the
Halko system. It allows users to create and modify drying programs, monitor
active drying sessions, and control the overall system.

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

The system includes an Arduino-based sensor unit for temperature monitoring
in the kiln and a service that provides a REST API for integration with the
executor component.

### Hardware Components

The sensor unit is based on an Arduino and uses MAX6675 thermocouple sensors
to measure temperatures. It can display status messages on an LCD.

### Arduino Firmware

The Arduino firmware for the sensor unit is located at
`/sensorunit/arduino/sensorunit/sensorunit.ino`. This firmware handles:

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

The sensor unit's service provides an endpoint to check the connection status
and another to update the status message displayed on the LCD.

### Integration

The Executor component communicates with the SensorUnit's REST API to retrieve
temperature data during drying programs. The SensorUnit continues to display
temperatures locally even when disconnected from the main system. The service
part of the SensorUnit handles the serial communication with the Arduino and
exposes the data via HTTP.

#### Configuration

The SensorUnit configuration is stored in a JSON file
(`/etc/opt/halko/sensorunit.json`) and includes parameters like `SerialPort`
and `BaudRate`.

#### Systemd Service

The SensorUnit service file (`sensorunit.service`) is located in the
`templates` directory and installed to `/etc/systemd/system/`. It can be
enabled and started with:

```bash
sudo systemctl enable --now sensorunit
```

## API Endpoints

For detailed API documentation including request/response formats
and endpoint specifications, see [API.md](API.md).

## System Configuration

The system uses JSON configuration files to define connection endpoints,
behavior parameters, and hardware settings.

### Main Configuration File (`halko.cfg`)

The main configuration file contains settings for all components. Here's an
example configuration:

```json
{
  "executor": {
    "base_path": "/var/opt/halko",
    "port": 8089,
    "tick_length": 6000,
    "sensor_unit_url": "http://localhost:8089/sensors",
    "power_unit_url": "http://localhost:8090/powers",
    "status_message_url": "http://localhost:8089/status",
    "network_interface": "eth0",
    "pid_settings": {
      "acclimate": {"kp": 2.0, "ki": 1.0, "kd": 0.5},
      "cooling": null,
      "heating": null
    },
    "max_delta_heating": 10.0,
    "min_delta_heating": 5.0
  },
  "power_unit": {
    "shelly_address": "http://localhost:8091",
    "cycle_length": 60,
    "max_idle_time": 70,
    "power_mapping": {
      "heater": 0,
      "humidifier": 1,
      "fan": 2
    }
  },
  "sensorunit": {
    "serial_device": "/dev/ttyUSB0",
    "baud_rate": 9600
  }
}
```

### Executor Configuration Options

- **`base_path`**: Directory for storing program data and execution logs
- **`port`**: HTTP server port for the executor API
- **`tick_length`**: Execution tick duration in milliseconds
- **`sensor_unit_url`**: Base URL for sensor unit API calls
- **`power_unit_url`**: Base URL for power unit API calls
- **`status_message_url`**: URL endpoint for heartbeat status messages
- **`network_interface`**: Network interface name for IP address reporting
  (e.g., "eth0", "wlan0")
- **`pid_settings`**: PID controller parameters for different program phases
- **`max_delta_heating`** / **`min_delta_heating`**: Temperature control
  limits

### Heartbeat Service

The executor includes an automatic heartbeat service that:

- Reports the executor's IP address every 30 seconds
- Uses the configured `network_interface` to determine the IP address
- Sends status messages to the `status_message_url` endpoint
- Helps monitor executor availability in distributed deployments
- Starts automatically when the executor service starts

The heartbeat sends a JSON payload in the following format:

```json
{"message": "192.168.1.100"}
```

Where the message contains the IPv4 address of the configured network
interface.

## Deployment

The system components are designed to run as systemd services. After building,
use `make install` to install the binaries and `make systemd-units` to set up
the systemd services.

Each component can be controlled independently:

```bash
# Start a specific component
sudo systemctl start halko@executor

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

# Halko

Halko is a distributed system for controlling and monitoring wood drying kilns.
It consists of multiple components that work together to provide temperature
control, power management, and program execution capabilities.

## Overview

The system is built with a microservices architecture with these main components:

- **ControlUnit**: Runs drying programs and controls the kiln.
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
# Build all Go backend components
make all

# Clean and rebuild all components from scratch
make build

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

# WebApp development and deployment
make webapp-install-node # Install Node.js 18 using nvm (if needed)
make webapp-install      # Install webapp dependencies
make webapp-dev          # Start development server
make webapp-build        # Build for production
make webapp-clean        # Clean webapp artifacts
make webapp-docker-build # Build webapp Docker image
```

## Project Structure

### Components

#### `/controlunit`

The ControlUnit is the core service that executes drying programs. It manages the
state machine for program execution, interacts with the PowerUnit to control
heating elements, and with the SensorUnit (or Simulator) to monitor
temperatures. It also provides a REST API to manage and monitor program
execution.

The ControlUnit includes a heartbeat service that periodically reports its IP
address to a configured status endpoint. This allows monitoring systems to
track the location and availability of the controlunit service in distributed
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

For detailed information about hardware setup, serial communication, and
configuration, see [sensorunit/README.md](sensorunit/README.md).

#### `/simulator`

The Simulator emulates the physical components of the kiln, such as
temperature sensors and Shelly power controls. This is useful for development
and testing without requiring actual hardware. It mimics the REST APIs of the
SensorUnit and parts of the PowerUnit (Shelly devices).

#### `/webapp`

A React-based frontend application that serves as the user interface for the
Halko system. It allows users to create and modify drying programs, monitor
active drying sessions, and control the overall system. Built with React 18,
TypeScript, Material-UI, and Redux Toolkit. See [webapp/README.md](webapp/README.md)
for detailed development and deployment instructions.

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

## API Endpoints

For detailed API documentation including request/response formats
and endpoint specifications, see [API.md](API.md).

## Program Structure

For detailed information about kiln drying program structure, step types,
power control methods, and validation rules, see [PROGRAM.md](PROGRAM.md).

## System Configuration

The system uses JSON configuration files to define connection endpoints,
behavior parameters, and hardware settings.

### Main Configuration File (`halko.cfg`)

The main configuration file contains settings for all components. Here's an
example configuration:

```json
{
  "controlunit": {
    "base_path": "/var/opt/halko",
    "tick_length": 6000,
    "network_interface": "eth0",
    "defaults": {
      "pid_settings": {
        "acclimate": {"kp": 2.0, "ki": 1.0, "kd": 0.5}
      },
      "max_delta_heating": 10.0,
      "min_delta_heating": 5.0
    }
  },
  "power_unit": {
    "shelly_address": "http://localhost:8088",
    "cycle_length": 60,
    "max_idle_time": 70,
    "power_mapping": {
      "heater": 0,
      "humidifier": 1,
      "fan": 2
    }
  },
  "storage": {
    "base_path": "/var/opt/halko"
  },
  "sensorunit": {
    "serial_device": "/dev/ttyUSB0",
    "baud_rate": 9600
  },
  "api_endpoints": {
    "controlunit": {
      "url": "http://localhost:8090",
      "status": "/status",
      "programs": "/programs",
      "running": "/running"
    },
    "sensorunit": {
      "url": "http://localhost:8093",
      "status": "/status",
      "temperatures": "/temperatures",
      "display": "/display"
    },
    "powerunit": {
      "url": "http://localhost:8092",
      "status": "/status",
      "power": "/power"
    },
    "storage": {
      "url": "http://localhost:8091",
      "status": "/status",
      "programs": "/programs",
      "execution_log": "/log"
    }
  }
}
```

### ControlUnit Configuration Options

- **`base_path`**: Directory for storing program data and execution logs
- **`tick_length`**: Execution tick duration in milliseconds
- **`network_interface`**: Network interface name for IP address reporting
  (e.g., "eth0", "wlan0")
- **`defaults`**: Default configuration settings
  - **`pid_settings`**: PID controller parameters for different program phases
  - **`max_delta_heating`** / **`min_delta_heating`**: Temperature control
    limits

### Heartbeat Service

The controlunit includes an automatic heartbeat service that:

- Reports the controlunit's IP address every 30 seconds
- Uses the configured `network_interface` to determine the IP address
- Sends status messages to the sensorunit display endpoint (configured in `api_endpoints.sensorunit`)
- Helps monitor controlunit availability in distributed deployments
- Starts automatically when the controlunit service starts

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
sudo systemctl start halko@controlunit

# Stop a component
sudo systemctl stop halko@controlunit

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

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

# Build and install webapp to /var/www/halko with nginx config
make install-webapp

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
execution, as well as storage management for drying programs and execution logs.

The ControlUnit uses a file-based storage system with separate directories for active
and completed program executions:

- `{base_path}/running/` - Files for currently executing programs
- `{base_path}/history/` - Completed program executions and their logs

When a program starts, execution files are created in the `running/` directory. Upon
completion, these files are automatically moved to the appropriate `history/` subdirectories.
On startup, any orphaned files in `running/` from previous crashes are cleaned up.

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

The simulator uses physics-based simulation engines (simple, differential, thermodynamic)
configured via `simulator.conf`. See [SIMULATOR.md](SIMULATOR.md) for detailed
configuration options and physics engine descriptions.

#### `/webapp`

A React-based frontend application that serves as the user interface for the
Halko system. It allows users to create and modify drying programs, monitor
active drying sessions, and control the overall system. Built with React 18,
TypeScript, Material-UI, and Redux Toolkit. See [webapp/README.md](webapp/README.md)
for detailed development and deployment instructions.

### Supporting Directories

#### `/bin`

Contains built executables for all components.

#### `/templates`

Contains systemd service templates and configuration samples. See [templates/README.md](templates/README.md) for configuration guidance.

#### `/tests`

Integration tests for the system components. Tests validate configuration loading, program validation with defaults, and Shelly API compatibility.

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
  "sensorunit": {
    "serial_device": "/dev/ttyUSB0",
    "baud_rate": 9600
  },
  "api_endpoints": {
    "controlunit": {
      "url": "http://localhost:8090",
      "engine": "/engine",
      "programs": "/programs",
      "status": "/status"
    },
    "sensorunit": {
      "url": "http://localhost:8093",
      "temperatures": "/temperatures",
      "display": "/display",
      "status": "/status"
    },
    "powerunit": {
      "url": "http://localhost:8092",
      "status": "/status",
      "power": "/power"
    }
  }
}
```

**Note**: Storage endpoints are served by the controlunit service at `/programs` (stored program templates)
and `/engine` (execution management). There is no separate storage service.

### ControlUnit Configuration Options

- **`base_path`**: Base directory for program storage. Contains:
  - `programs/` - Stored program templates
  - `running/` - Active program executions (auto-created)
  - `history/` - Completed executions with `logs/` and `status/` subdirectories (auto-created)
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

### Production Installation (Bare-Metal)

The system is designed for bare-metal production deployment with systemd services.

#### 1. Install Backend Services

```bash
# Install binaries to /opt/halko and config to /etc/opt/halko.cfg
make install

# Install and enable systemd services for controlunit, powerunit, sensorunit
make systemd-units
```

**Important**: Before starting services, edit `/etc/opt/halko.cfg` to configure:

- `network_interface`: Your system's network interface name (see [Network Interface Configuration](#network-interface-configuration))
- `serial_device`: Your Arduino device path (typically `/dev/ttyUSB0`)
- `shelly_address`: Your Shelly device IP address or hostname

Each component can be controlled independently:

```bash
# Start a specific component
sudo systemctl start halko@controlunit

# Stop a component
sudo systemctl stop halko@controlunit

# Check status
sudo systemctl status halko@powerunit

# View logs
sudo journalctl -u halko@controlunit -f
```

#### 2. Install WebApp

```bash
# Build and install webapp to /var/www/halko
make install-webapp

# Enable nginx site
sudo ln -s /etc/nginx/sites-available/halko /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

Access the webapp at `http://your-server-ip/`

The webapp proxies API requests to localhost backend services (ports 8090, 8092, 8093).

#### Network Interface Configuration

The `network_interface` setting in `/etc/opt/halko.cfg` must match your system's actual interface name for the heartbeat service to work correctly.

To find your interface name:

```bash
ip addr show
```

Common interface names:

- `eth0` - Traditional Ethernet naming
- `enp0s3`, `ens33`, `enp1s0` - Predictable PCI Ethernet names
- `wlan0` - WiFi interfaces
- `lo` - Loopback (do not use for heartbeat)

Update the configuration before starting services:

```bash
sudo nano /etc/opt/halko.cfg
# Change "network_interface": "eth0" to your interface name
```

### Docker Deployment

The system can also be deployed using Docker Compose. The containers are
configured to run as the host user to ensure proper file ownership.

#### Prerequisites

```bash
# Build all binaries and Docker images
make images
```

This target automatically:

- Rebuilds all Go binaries
- Builds webapp for production with nginx
- Generates `webapp/nginx-docker.conf` with WebSocket support
- Creates Docker images for all services

#### Configuration Files

Docker deployment uses specific configuration files:

- `halko-docker.cfg` - Service endpoints using Docker service names
- `simulator.conf` - Simulator physics engine configuration (see [SIMULATOR.md](SIMULATOR.md))

The simulator requires both configuration files mounted in the container.

#### File Ownership Configuration

By default, containers run as UID:GID 1000:1000. To use your current user's
UID/GID for proper file ownership on the host:

```bash
# Set environment variables before starting containers
export UID=$(id -u)
export GID=$(id -g)
```

#### Starting the Services

```bash
# Build and start all services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop all services
docker-compose down
```

Files created by the containers in the `fsdb/` directory will be owned by the
specified UID:GID on the host system.

#### Service Architecture

Docker deployment includes:

- **controlunit:8090** - Engine and storage endpoints
- **powerunit:8092** - Shelly device interface (proxies to simulator)
- **simulator:8088/8093** - Emulated hardware (Shelly + sensors)
- **webapp:8080** - React UI with nginx reverse proxy

The webapp nginx configuration includes:

- WebSocket upgrade support for live log streaming (`/api/v1/controlunit/engine/running/logws`)
- API proxying to backend services at `/api/v1/*`
- CORS headers for all endpoints

Access the webapp at <http://localhost:8080>

## Development

For development, you can run the simulator instead of connecting to real hardware:

```bash
./bin/simulator -c halko.cfg -s simulator.conf
```

The webapp can be run in development mode from the `/webapp` directory:

```bash
cd webapp
npm install
npm start
```

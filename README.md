# Halko

Halko is a distributed system for controlling and monitoring wood drying kilns.
It consists of multiple components that work together to provide temperature
control, power management, and program execution capabilities.

## Overview

The system is built with a microservices architecture with these main components:

- **ControlUnit**: Runs drying programs and controls the kiln.
- **PowerUnit**: Interfaces with Shelly devices to control power.
- **SensorUnit**: Reads temperature data from physical sensors and provides an API.
- **DBusUnit**: Systemd D-Bus integration for VPN and host power control.
- **Simulator**: Provides a simulated environment for testing.
- **WebApp**: User interface for controlling and monitoring the system.
- **halkoctl**: Command-line tool for managing the system.

For the project's history and credits, see [HISTORY.md](HISTORY.md).

## Building the Project

### Prerequisites

- Go 1.26 or higher
- Node.js 18+ and npm (for the webapp; if not present, the Makefile installs a
  local copy under `.nodejs/` automatically)
- golangci-lint (for linting)
- Arduino CLI (only for building/flashing the ESP32 firmware; installed by
  `make prepare-esp32`)

### Build Commands

The project uses a Makefile to simplify building and deployment:

```bash
# Build all Go backend components
make all

# Clean and rebuild all components from scratch
make build

# Remove all built binaries
make clean

# Run all test suites
make test

# Run linters (Go, markdown, webapp)
make lint

# Update Go module dependencies
make update-modules

# Run go mod tidy on all modules
make go-tidy

# Install binaries to /opt/halko and config to /etc/opt
make install

# Install and enable systemd service units
make systemd-units

# Build and install webapp to /var/www/halko with nginx config
make install-webapp

# Reformat changed Go files
make fmt

# WebApp development and deployment
make run-webapp          # Start development server on :1234 with hot reload
make build-webapp        # Build production bundle to webapp/dist/
make clean-webapp        # Clean webapp artifacts

# ESP32 firmware
make prepare-esp32       # Install Arduino CLI, ESP32 core, and libraries
make build-esp32         # Compile firmware
make upload-esp32        # Flash to device
make monitor-esp32       # Serial monitor
```

### Tests

There is no separate test module — tests live in the same package as the code
they cover, the standard Go layout. `make test` runs them across every module,
`make test-<module>` runs one, and `make test-race` runs the lot under the race
detector. Tests that drive a service end to end live with that service and are
self-contained: `simulator/shelly_api_test.go` builds the simulator binary,
runs it as a real process and drives its HTTP API, so no test depends on a
service you are expected to have running.

`make test` and `make lint` deliberately report every module rather than
stopping at the first failure, so they always exit 0. CI uses the gating
counterparts `make test-ci`, `make lint-ci` and `make lint-markdown-ci`;
`make ci-check` runs the whole suite the way CI does, which is what to run
before pushing.

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
heaters, fans, and steam injectors. It provides a REST API for direct power
control operations.

#### `/sensorunit`

The SensorUnit component includes:

- ESP32 firmware (`sensorunit/esp32/sensorunit/sensorunit.ino`) for a
  physical unit that reads from MAX31855 thermocouples and displays status
  on an I2C OLED display.
- A Go service (`sensorunit/main.go`) that communicates with the ESP32 via
  USB serial and exposes a REST API for temperature and status.

For detailed information about hardware setup, serial communication, and
configuration, see [sensorunit/README.md](sensorunit/README.md) and
[sensorunit/esp32/README.md](sensorunit/esp32/README.md).

#### `/simulator`

The Simulator emulates the physical components of the kiln, such as
temperature sensors and Shelly power controls. This is useful for development
and testing without requiring actual hardware. It emulates the Shelly devices
over HTTP, and the ESP32 sensor unit at the serial level over a
pseudo-terminal, so the real `sensorunit` service runs against it unmodified.

The simulator uses physics-based simulation engines (simple, differential, thermodynamic)
configured via `simulator.conf`. See [SIMULATOR.md](SIMULATOR.md) for detailed
configuration options and physics engine descriptions.

#### `/dbusunit`

The DBusUnit provides systemd D-Bus integration for the host system. It
exposes a REST API for listing and controlling VPN connections (systemd
units) and for shutting down or rebooting the host. It runs as root under
its own systemd unit (`halko-dbusunit.service`) because it needs access to
the system D-Bus.

#### `/halkoctl`

A command-line tool for managing the system: sending programs, querying
status and temperatures, and updating the sensor unit display. See
[halkoctl/README.md](halkoctl/README.md).

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

#### `/example`

An example drying program that passes validation and can be sent with
`halkoctl send`.

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
    "tick_length": "6s",
    "network_interface": "eth0",
    "defaults": {
      "deltas": {
        "heating": {"min_delta": 5.0, "max_delta": 10.0},
        "acclimate": {"min_delta": -1.0, "max_delta": 3.0}
      },
      "fan_power": 0,
      "steam_power": 0,
      "preheat_fan_power": 50,
      "max_target_temperature": 200,
      "steam_ceiling": 100,
      "sensor_timeout": "120s",
      "execution_log_interval": "60s"
    }
  },
  "power_unit": {
    "shelly_address": "http://localhost:8088",
    "cycle_length": "60s",
    "max_idle_time": "70s",
    "power_mapping": {
      "heater": 0,
      "steam": 1,
      "fan": 2
    }
  },
  "sensorunit": {
    "serial_device": "/dev/ttyUSB0",
    "baud_rate": 9600
  },
  "dbusunit": {
    "system_bus_socket": "/var/run/dbus/system_bus_socket"
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
    },
    "dbusunit": {
      "url": "http://localhost:8094",
      "vpn": "/vpn",
      "power": "/power",
      "status": "/status"
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
- **`tick_length`**: Execution tick duration (Go duration format: "6s", "100ms", etc.)
- **`network_interface`**: Network interface name for IP address reporting
  (e.g., "eth0", "wlan0")
- **`defaults`**: Everything the control unit would otherwise have to invent.
  All of it is required; a missing entry fails at startup rather than becoming a
  zero somewhere downstream. The webapp reads the same block from
  `GET /engine/defaults`, so the two cannot drift.
  - **`deltas.heating`**: the band a heating step holds the kiln in, measured
    from the **material**. Both values positive.
  - **`deltas.acclimate`**: the band an acclimate step holds the kiln in,
    measured from the **target**. `min_delta` negative (how far the kiln may
    sag before the heater fires), `max_delta` positive (how far the kiln may
    lead the material).
  - **`fan_power`** / **`steam_power`**: the power a step gets for a component
    it does not name.
  - **`preheat_fan_power`**: the fan power used while preheating, before the
    first step begins.
  - **`max_target_temperature`**: the highest target any step may ask for.
  - **`steam_ceiling`**: the temperature steam cannot heat the kiln past. Above
    it steam is thermally neutral; below it steam outruns the heater, which is
    what the steam rules exist to prevent.
  - **`sensor_timeout`**: how long a probe may go without a valid reading before
    the running program is failed and all power switched off.
  - **`execution_log_interval`**: how often a running program appends a line to
    its execution log.

### PowerUnit Configuration Options

- **`shelly_address`**: Base URL of the Shelly device (or the simulator's
  emulated Shelly during development)
- **`cycle_length`**: Duty-cycle period (Go duration format). Power percentages
  are realized by switching relays on for that fraction of each cycle
- **`max_idle_time`**: How long the PowerUnit will keep applying the last
  commanded percentages without hearing from the ControlUnit. Past it, all
  percentages are reset to 0 and the relays go off on the next cycle. This is a
  safety watchdog: only an incoming command refreshes the timer, so neither the
  running duty cycle nor status polling (the webapp does it every few seconds)
  can hold it off. Keep it slightly longer than `cycle_length`
- **`power_mapping`**: Maps channel names (`heater`, `steam`, `fan`) to Shelly
  switch IDs

### SensorUnit Configuration Options

- **`serial_device`**: Path to the ESP32 serial device. In production this is
  real hardware (typically `/dev/ttyUSB0`); during development it must instead
  name a path the simulator may create, such as `/tmp/esp32-halko`
- **`baud_rate`**: Serial speed, 9600 to match the firmware

### DBusUnit Configuration Options

- **`system_bus_socket`**: Path to the D-Bus system bus socket. The whole
  section is optional — when omitted the standard system location is used.
  Override it when the socket is elsewhere, for example inside a
  distrobox/toolbox container where the host bus is exposed at
  `/run/host/run/dbus/system_bus_socket`

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

**For Raspberry Pi deployment:** See [RASPBERRY_PI.md](RASPBERRY_PI.md) for detailed instructions on memory-optimized builds, USB boot setup, and dual network interface configuration.

#### 1. Install Backend Services

```bash
# Install binaries to /opt/halko and config to /etc/opt/halko.cfg
make install

# Install and enable systemd services for controlunit, powerunit, sensorunit
# (as halko@<name>) and dbusunit (as halko-dbusunit, running as root)
make systemd-units
```

**Important**: Before starting services, edit `/etc/opt/halko.cfg` to configure:

- `network_interface`: Your system's network interface name (see [Network Interface Configuration](#network-interface-configuration))
- `serial_device`: Your ESP32 device path (typically `/dev/ttyUSB0`)
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

# The dbusunit uses its own (non-templated) unit
sudo systemctl status halko-dbusunit
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

The webapp proxies API requests to localhost backend services (ports 8090,
8092, 8093, 8094). The nginx configuration is generated from `halko.cfg` by
`halkoctl nginx` during `make build-webapp`.

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

### Memory Monitoring

Monitor process memory usage to detect potential memory leaks during development or testing:

```bash
# Console output only (default)
./scripts/monitor-memory.py

# Save detailed CSV log
./scripts/monitor-memory.py -o memory-test.csv

# Monitor specific processes for 1 hour
./scripts/monitor-memory.py -p controlunit simulator -i 5 -t 3600

# Via Makefile
make monitor-memory MONITOR_ARGS="-o test.csv -t 7200"
```

Run `./scripts/monitor-memory.py --help` for all options.

### Tmux-Based Development Workflow

For rapid development and debugging, use the tmux environment to run all services locally:

```bash
# Start all services in tmux
make tmux-debug-run

# Same, but the simulator injects escalating sensor failures (see SIMULATOR.md)
make tmux-debug-fail-run

# With custom log level (0=ERROR, 1=WARN, 2=INFO, 3=DEBUG, 4=TRACE)
LOGLEVEL=4 make tmux-debug-run

# With specific simulator engine
SIMULATOR=thermodynamic make tmux-debug-run

# Write the log files somewhere other than logs/
LOG_DIR=/tmp/halko-logs make tmux-debug-run

# Combined options
LOGLEVEL=4 SIMULATOR=differential make tmux-debug-run

# Attach to running session
tmux attach -t halko-debug

# Stop all services
make tmux-debug-stop
```

The tmux session includes separate windows for:

- **simulator** - Hardware emulation (Shelly over HTTP, ESP32 over a pseudo-terminal)
- **powerunit** - Power control service
- **sensorunit** - Serial bridge, running against the simulator's emulated ESP32
- **controlunit** - Main control logic
- **dbusunit** - Systemd D-Bus integration (runs with sudo)
- **webapp** - Development server with hot reload
- **shell** - Command line for halkoctl and testing

Every window pipes its output through `tee`, so logs are on the console *and*
in `logs/<window>.log` under the workspace root — `logs/controlunit.log`,
`logs/simulator.log` and so on. That directory is git-ignored and overwritten
on every start, so it always holds the latest run; override it with `LOG_DIR`
(the start script aborts if it cannot create or write there). Together with
`fsdb/running/` and `fsdb/history/` these logs are the primary evidence for
what happened during a run.

Note that `sensorunit.serial_device` in `halko.cfg` must name a path the
simulator may create (for example `/tmp/esp32-halko`) rather than real
hardware, or the simulator refuses to start.

This workflow provides fast iteration with real-time log viewing and easy service restarts.

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
Arduino code for the physical temperature sensor unit.

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

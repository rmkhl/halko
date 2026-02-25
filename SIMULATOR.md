# Halko Simulator Documentation

The Halko Simulator emulates the physical behavior of a wood drying kiln for development and testing without requiring actual hardware. It provides realistic temperature dynamics and responds to control commands just like real hardware would.

## Overview

The simulator consists of two main components:

1. **Shelly Device Emulator** (port 8088) - Emulates Shelly smart switches for power control
2. **SensorUnit Emulator** (port 8093) - Emulates temperature sensors and LCD display

The simulator uses physics-based models to calculate how temperatures change over time based on power settings, making it suitable for testing control algorithms and drying programs.

## Configuration

The simulator requires two configuration files:

1. **halko.cfg** - Standard Halko configuration for API endpoints
2. **simulator.conf** - Simulator-specific physics engine configuration

### Running the Simulator

```bash
# With explicit config files
./bin/simulator -c halko.cfg -s simulator.conf

# Auto-discovery (looks for simulator.conf in standard locations)
./bin/simulator -c halko.cfg

# In Docker (automatic configuration)
docker-compose up simulator
```

### Configuration File Search Order

The simulator searches for `simulator.conf` in:

1. Path specified with `-s` flag
2. `SIMULATOR_CONFIG` environment variable
3. `./simulator.conf` (current directory)
4. `~/.simulator.conf` (user home)
5. `/etc/opt/simulator.conf` (system-wide)

## Simulator Configuration Structure

```json
{
  "tick_duration": "6s",
  "status_interval": 10,
  "initial_oven_temp": 20.0,
  "initial_material_temp": 20.0,
  "environment_temp": 20.0,
  "simulation_engine": "simple|differential|thermodynamic",
  "engine_config": { /* engine-specific parameters */ }
}
```

### Common Parameters

- **tick_duration** (string): Time between simulation updates (Go duration format: "6s", "1m", etc.)
- **status_interval** (int): Log status every N ticks (0 = disabled)
- **initial_oven_temp** (float): Starting oven temperature in °C
- **initial_material_temp** (float): Starting wood temperature in °C
- **environment_temp** (float): Ambient temperature in °C
- **simulation_engine** (string): Physics engine to use (`simple`, `differential`, or `thermodynamic`)
- **engine_config** (object): Engine-specific configuration parameters

## Physics Engines

The simulator supports three physics engines with increasing levels of realism:

### 1. Simple Engine

Basic rate-based temperature model suitable for quick testing.

**Configuration (`simulator.conf`):**

```json
{
  "tick_duration": "6s",
  "status_interval": 10,
  "initial_oven_temp": 20.0,
  "initial_material_temp": 20.0,
  "environment_temp": 20.0,
  "simulation_engine": "simple",
  "engine_config": {
    "heating_rate": 0.1,
    "cooling_rate": 0.01,
    "transfer_rate": 0.01
  }
}
```

**Engine Parameters:**

- **heating_rate** (float): Temperature increase per tick per % heater power
- **cooling_rate** (float): Temperature decrease per tick toward environment temperature
- **transfer_rate** (float): Heat transfer rate from oven to material per tick

**Behavior:**

- Linear heating based on heater power
- Exponential cooling toward ambient temperature
- Simple heat transfer between oven and material
- Fast simulation, low accuracy
- Good for UI testing and basic program validation

### 2. Differential Engine

Thermal mass-based model using differential equations for more realistic behavior.

**Configuration (`simulator-differential.conf`):**

```json
{
  "tick_duration": "6s",
  "status_interval": 10,
  "initial_oven_temp": 20.0,
  "initial_material_temp": 20.0,
  "environment_temp": 20.0,
  "simulation_engine": "differential",
  "engine_config": {
    "heater_power": 1.0,
    "heat_loss_coefficient": 0.05,
    "heat_transfer_coefficient": 0.15,
    "oven_thermal_mass": 1.0,
    "material_thermal_mass": 3.0
  }
}
```

**Engine Parameters:**

- **heater_power** (float): Maximum heating power (watts equivalent) at 100%
- **heat_loss_coefficient** (float): Rate of heat loss to environment
- **heat_transfer_coefficient** (float): Rate of heat transfer oven→material
- **oven_thermal_mass** (float): Thermal inertia of oven (higher = slower temperature changes)
- **material_thermal_mass** (float): Thermal inertia of wood (typically 2-4× oven mass)

**Behavior:**

- Differential equations model heat flow
- Realistic thermal inertia effects
- Material heats slower than oven (realistic lag)
- Good balance of accuracy and performance
- Suitable for control algorithm development

### 3. Thermodynamic Engine

High-fidelity physics simulation using thermodynamic principles.

**Configuration (`simulator-thermodynamic.conf`):**

```json
{
  "tick_duration": "6s",
  "status_interval": 10,
  "initial_oven_temp": 20.0,
  "initial_material_temp": 20.0,
  "environment_temp": 20.0,
  "simulation_engine": "thermodynamic",
  "engine_config": {
    "oven": {
      "mass": 140.0,
      "specific_heat": 500.0,
      "surface_area": 4.5,
      "wall_u_value": 0.5,
      "emissivity": 0.9
    },
    "air": {
      "volume": 0.5,
      "specific_heat": 1005.0
    },
    "material": {
      "mass": 100.0,
      "specific_heat": 1700.0,
      "surface_area": 3.0
    },
    "heater": {
      "wattage": 2000.0,
      "efficiency": 0.95
    },
    "convection": {
      "natural": 10.0,
      "forced": 80.0,
      "fan_waste_heat": 50.0
    },
    "environment": {
      "temperature": 20.0
    },
    "physics": {
      "stefan_boltzmann": 5.67e-8,
      "time_step": 6.0
    }
  }
}
```

**Engine Parameters:**

**Oven Properties:**

- **mass** (kg): Mass of kiln structure
- **specific_heat** (J/kg·K): Heat capacity of kiln walls
- **surface_area** (m²): External surface area
- **wall_u_value** (W/m²·K): Thermal conductance of walls
- **emissivity** (0-1): Radiant heat emission coefficient

**Air Properties:**

- **volume** (m³): Interior air volume
- **specific_heat** (J/kg·K): Heat capacity of air (typically 1005)

**Material (Wood) Properties:**

- **mass** (kg): Mass of wood being dried
- **specific_heat** (J/kg·K): Heat capacity of wood (typically 1700)
- **surface_area** (m²): Wood surface area for heat transfer

**Heater Properties:**

- **wattage** (W): Maximum heater power output
- **efficiency** (0-1): Electrical to thermal conversion efficiency

**Convection Properties:**

- **natural** (W/m²·K): Natural convection coefficient (fan off)
- **forced** (W/m²·K): Forced convection coefficient (fan on)
- **fan_waste_heat** (W): Heat generated by fan motor

**Physics Constants:**

- **stefan_boltzmann** (W/m²·K⁴): Stefan-Boltzmann constant (5.67e-8)
- **time_step** (s): Simulation time step (should match tick_duration)

**Behavior:**

- Models conduction, convection, and radiation
- Accounts for thermal mass and specific heat
- Simulates fan effects on heat transfer
- Heat losses through walls to environment
- Fan motor waste heat contribution
- Most accurate simulation
- Best for validating control strategies before deployment

## Choosing a Physics Engine

| Engine | Accuracy | Speed | Use Case |
|--------|----------|-------|----------|
| **Simple** | Low | Fast | UI testing, rapid iteration |
| **Differential** | Medium | Moderate | Control algorithm development |
| **Thermodynamic** | High | Slower | Final validation, characterization |

**Recommendations:**

- **Development & UI Testing**: Use `simple` for fast feedback
- **Control Tuning**: Use `differential` for realistic thermal behavior
- **Pre-Deployment Validation**: Use `thermodynamic` for accurate predictions

## Emulated APIs

### Shelly Device API (Port 8088)

Emulates Shelly smart switch RPC endpoints:

- `GET /rpc/Switch.GetStatus?id=N` - Get switch state
- `GET /rpc/Switch.Set?id=N&on=true|false` - Set switch state

Switch mapping (from halko.cfg):

- 0 = heater
- 1 = humidifier
- 2 = fan

### SensorUnit API (Port 8093)

Emulates temperature sensor endpoints:

- `GET /temperatures` - Get current simulated temperatures
- `POST /display` - Receive display messages (logged)
- `GET /status` - Service health status

**Response Format:**

```json
{
  "data": {
    "oven": 45.2,
    "wood": 42.5,
    "ovenPrimary": 45.2,
    "ovenSecondary": 44.8
  }
}
```

## Docker Integration

The simulator is automatically configured in Docker Compose environments:

```yaml
simulator:
  volumes:
    - ./halko-docker.cfg:/etc/halko/halko.cfg:ro
    - ./simulator.conf:/etc/halko/simulator.conf:ro
  ports:
    - "8088:8088"  # Shelly emulation
    - "8093:8093"  # SensorUnit emulation
```

Both configuration files are mounted and used automatically.

## Status Logging

When `status_interval > 0`, the simulator logs internal state periodically:

```
[INFO] Tick 10: Oven=45.2°C Material=42.5°C Heater=75% Fan=50%
```

Set to 0 to disable status logging (recommended for production-like testing).

## Development Tips

1. **Start with Simple**: Use `simple` engine first to verify program logic
2. **Tune with Differential**: Use `differential` to tune PID/delta parameters
3. **Validate with Thermodynamic**: Final testing with `thermodynamic` before hardware deployment
4. **Match Tick Duration**: Set `tick_duration` to match `controlunit.tick_length` for realistic timing
5. **Realistic Initial Conditions**: Set `initial_*_temp` to room temperature (20°C) for realistic startup
6. **Status Interval**: Use status logging during development, disable for performance testing

## Limitations

- No moisture content modeling (only temperature)
- No humidity effects on heat transfer
- Fan effects are simplified (on/off, no speed variation)
- No wood shrinkage or cracking simulation
- Humidifier has no effect on physics (placeholder only)

These limitations make the simulator suitable for temperature control development but not for predicting actual drying outcomes.

## See Also

- [API.md](API.md) - Complete API endpoint documentation
- [PROGRAM.md](PROGRAM.md) - Program structure and validation rules
- [README.md](README.md#docker-deployment) - Docker deployment guide

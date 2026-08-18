# Halko Simulator Documentation

The Halko Simulator emulates the physical behavior of a wood drying kiln for development and testing without requiring actual hardware. It provides realistic temperature dynamics and responds to control commands just like real hardware would.

## Overview

The simulator consists of two main components:

1. **Shelly Device Emulator** (port 8088) - Emulates Shelly smart switches for power control
2. **ESP32 Emulator** (pseudo-terminal) - Emulates the sensor unit hardware at
   the serial level, so the real `sensorunit` service runs against it unmodified

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
```

### Configuration File Search Order

The simulator searches for `simulator.conf` in:

1. Path specified with `-s` flag
2. `SIMULATOR_CONFIG` environment variable
3. `./simulator.conf` (current directory)
4. `~/.simulator.conf` or `~/.config/simulator.conf` (user home)
5. `/etc/halko/simulator.conf`, `/etc/opt/halko/simulator.conf`, or
   `/etc/opt/simulator.conf` (system-wide)
6. Next to the simulator executable

## Simulator Configuration Structure

```json
{
  "status_interval": 10,
  "initial_kiln_temp": 20.0,
  "initial_material_temp": 20.0,
  "environment_temp": 20.0,
  "simulation_engine": "simple|differential|thermodynamic",
  "engine_config": { /* engine-specific parameters */ }
}
```

**Note**: The simulator uses the tick duration from `controlunit.tick_length` in the main `halko.cfg` configuration file. This ensures the simulator always runs with the same timing as the control unit.

### Common Parameters

- **status_interval** (int): Log status every N ticks (0 = disabled)
- **initial_kiln_temp** (float): Starting kiln temperature in °C
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
  "status_interval": 10,
  "initial_kiln_temp": 20.0,
  "initial_material_temp": 20.0,
  "environment_temp": 20.0,
  "simulation_engine": "simple",
  "engine_config": {
    "heating_rate": 0.1,
    "cooling_rate": 0.03,
    "transfer_rate": 0.01
  }
}
```

**Engine Parameters:** (all are required, and all must be positive)

- **heating_rate** (float): Temperature increase per tick per % heater power
- **cooling_rate** (float): Temperature decrease per tick toward environment temperature
- **transfer_rate** (float): Heat transfer rate from kiln to material per tick

**Behavior:**

- Linear heating based on heater power
- Exponential cooling toward ambient temperature
- Simple heat transfer between kiln and material
- Fast simulation, low accuracy
- Good for UI testing and basic program validation

### 2. Differential Engine

Thermal mass-based model using differential equations for more realistic behavior.

**Configuration (`simulator-differential.conf`):**

```json
{
  "status_interval": 10,
  "initial_kiln_temp": 20.0,
  "initial_material_temp": 20.0,
  "environment_temp": 20.0,
  "simulation_engine": "differential",
  "engine_config": {
    "heater_power": 5.0,
    "heat_loss_coefficient": 0.02,
    "heat_transfer_coefficient": 0.15,
    "kiln_thermal_mass": 1.0,
    "material_thermal_mass": 3.0
  }
}
```

**Engine Parameters:** (all are required, and all must be positive)

- **heater_power** (float): Maximum heating power (watts equivalent) at 100%
- **heat_loss_coefficient** (float): Rate of heat loss to environment
- **heat_transfer_coefficient** (float): Rate of heat transfer kiln→material
- **kiln_thermal_mass** (float): Thermal inertia of kiln (higher = slower temperature changes)
- **material_thermal_mass** (float): Thermal inertia of wood (typically 2-4× kiln mass)

**Behavior:**

- Differential equations model heat flow
- Realistic thermal inertia effects
- Material heats slower than the kiln (realistic lag)
- Good balance of accuracy and performance
- Suitable for control algorithm development

### 3. Thermodynamic Engine

High-fidelity physics simulation using thermodynamic principles.

**Configuration (`simulator-thermodynamic.conf`):**

```json
{
  "status_interval": 10,
  "initial_kiln_temp": 20.0,
  "initial_material_temp": 20.0,
  "environment_temp": 20.0,
  "simulation_engine": "thermodynamic",
  "engine_config": {
    "kiln": {
      "mass": 140.0,
      "specific_heat": 500.0,
      "surface_area": 4.5,
      "wall_u_value": 0.08,
      "emissivity": 0.08
    },
    "air": {
      "volume": 0.5,
      "specific_heat": 1005.0
    },
    "material": {
      "mass": 20.0,
      "specific_heat": 1700.0,
      "surface_area": 1.5
    },
    "heater": {
      "wattage": 2000.0,
      "efficiency": 0.95
    },
    "convection": {
      "natural": 5.0,
      "forced": 15.0,
      "fan_waste_heat": 50.0
    },
    "environment": {
      "temperature": 20.0
    },
    "physics": {
      "stefan_boltzmann": 5.67e-8
    }
  }
}
```

**Note**: The `time_step` parameter is automatically injected by the simulator based on `controlunit.tick_length` from the main configuration.

**Engine Parameters:**

**Kiln Properties:**

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

**Note**: The `time_step` parameter (simulation time step in seconds) is automatically calculated from `controlunit.tick_length` and injected into the physics configuration at runtime.

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
- 1 = steam
- 2 = fan

### ESP32 Sensor Unit (pseudo-terminal)

The simulator does not serve a SensorUnit HTTP API. It emulates the sensor
hardware instead: it allocates a pseudo-terminal, links it at the path named
by `sensorunit.serial_device` in `halko.cfg`, and answers the ESP32's serial
protocol on it.

- `helo;` - Handshake, answers `helo`
- `read;` - Answers `KilnPrimary=XX.XC,KilnSecondary=XX.XC,Wood=XX.XC`
  (a failed sensor reports `NaN` in place of a value)
- `show TEXT;` - Display message (logged)

The real `sensorunit` service opens that device and serves `/temperatures`,
`/status` and `/display` itself, so the whole stack above the serial line is
the production code path.

**`serial_device` must name a path the simulator may create** — for example
`/tmp/esp32-halko`. A path under `/dev/` but outside `/dev/pts/` is refused at
startup, naming the offending path, because it looks like real hardware that
simply is not plugged in.

### Sensor Failure Injection

Passing `-fail-sensors` makes the simulator inject escalating sensor failures
during a run, so the control unit's failsafe can be exercised without
unplugging anything:

```bash
./bin/simulator -c halko.cfg -s simulator.conf -fail-sensors

# Or via the tmux workflow, which passes the flag for you
make tmux-debug-fail-run
```

Failures are injected into the three physical probes on a four-stage schedule,
measured from the moment the injector arms: intermittent dropouts at +60 s, one
probe permanently lost at +120 s, a second at +180 s, the last at +240 s. A
failed probe reports `NaN`, which the sensor unit forwards as an invalid
reading.

The control unit keeps the last valid value per sensor and fails the run —
switching all power off — once a sensor has gone 120 seconds without a valid
one. Note that the kiln reading survives on either kiln probe alone, so it only
goes invalid when the second one dies. The trip therefore lands somewhere
between ~240 s and ~300 s depending on which probe was lost first; a varying
trip time is expected, not a bug.

## Status Logging

When `status_interval > 0`, the simulator logs internal state periodically:

```text
[INFO] Simulation status - Tick #10: Kiln=45.2°C, Material=42.5°C, Heater=true, Fan=true, Steam=false
```

The heater, fan and steam values are the relay states for that tick, not
percentages — a duty-cycled channel shows as `true` for the on portion of its
cycle and `false` for the rest.

Set to 0 to disable status logging (recommended for production-like testing).

## Development Tips

1. **Start with Simple**: Use `simple` engine first to verify program logic
2. **Tune with Differential**: Use `differential` to tune delta parameters
3. **Validate with Thermodynamic**: Final testing with `thermodynamic` before hardware deployment
4. **Automatic Timing**: Simulator automatically uses `controlunit.tick_length` from `halko.cfg` for realistic timing
5. **Realistic Initial Conditions**: Set `initial_*_temp` to room temperature (20°C) for realistic startup
6. **Status Interval**: Use status logging during development, disable for performance testing

## Limitations

- No moisture content modeling (only temperature)
- No humidity effects on heat transfer
- Fan effects are simplified (on/off, no speed variation)
- No wood shrinkage or cracking simulation
- Steam has no effect on physics (placeholder only)

These limitations make the simulator suitable for temperature control development but not for predicting actual drying outcomes.

## See Also

- [API.md](API.md) - Complete API endpoint documentation
- [PROGRAM.md](PROGRAM.md) - Program structure and validation rules
- [README.md](README.md) - System overview and deployment guide

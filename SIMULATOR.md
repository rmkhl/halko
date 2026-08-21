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
2. **simulator-\*.conf** - Simulator-specific physics configuration

The shipped simulator configurations are named for the job they do, not for the
physics engine behind them, because two of them run the same engine at very
different settings:

| Config | `SIMULATOR=` | Job |
| ------ | ------------ | --- |
| `simulator-fast.conf` | `fast` (default) | Verifying FSM and program flow. An exaggerated kiln that runs a whole program in about an hour. |
| `simulator-tuning.conf` | `tuning` | Tuning program targets and deltas. Coefficients fitted to a real run. |
| `simulator-equalize.conf` | `equalize` | The equalization step. Same physics as `fast`, but the kiln starts hot so the step has work to do. |
| `simulator-thermodynamic.conf` | `thermodynamic` | The detailed model. |

There is deliberately no plain `simulator.conf`: the config has to be named, so
that a run always states which kiln it is simulating.

### Running the Simulator

```bash
./bin/simulator -c halko.cfg -s simulator-fast.conf
```

### Configuration File Search Order

Pass the config with `-s`. Without it the simulator falls back to searching for
a file literally named `simulator.conf`, in this order, and exits with an error
if it finds none:

1. `SIMULATOR_CONFIG` environment variable
2. `./simulator.conf` (current directory)
3. `~/.simulator.conf` or `~/.config/simulator.conf` (user home)
4. `/etc/halko/simulator.conf`, `/etc/opt/halko/simulator.conf`, or
   `/etc/opt/simulator.conf` (system-wide)
5. Next to the simulator executable

None of the shipped configs carry that name, so in this workspace `-s` is
effectively required.

## Simulator Configuration Structure

```jsonc
{
  "status_interval": 10,
  "initial_kiln_temp": 20.0,
  "initial_material_temp": 20.0,
  "environment_temp": 20.0,
  "simulation_engine": "differential|thermodynamic",
  "engine_config": { /* engine-specific parameters */ }
}
```

**Note**: The simulator uses the tick duration from `controlunit.tick_length` in the main `halko.cfg` configuration file. This ensures the simulator always runs with the same timing as the control unit.

### Common Parameters

- **status_interval** (int): Log status every N ticks (0 = disabled)
- **initial_kiln_temp** (float): Starting kiln temperature in °C
- **initial_material_temp** (float): Starting wood temperature in °C
- **environment_temp** (float): Ambient temperature in °C
- **simulation_engine** (string): Physics engine to use (`differential` or `thermodynamic`)
- **engine_config** (object): Engine-specific configuration parameters

## Physics Engines

The simulator has two physics engines. `differential` runs at two very
different settings - see the config table above - so the engine you pick and
the job you are doing are not the same choice.

### 1. Differential Engine

Thermal mass-based model using differential equations for more realistic behavior.

**Configuration (`simulator-tuning.conf`, fitted coefficients):**

```json
{
  "status_interval": 10,
  "initial_kiln_temp": 20.0,
  "initial_material_temp": 20.0,
  "environment_temp": 20.0,
  "simulation_engine": "differential",
  "engine_config": {
    "heater_power": 0.11433,
    "heat_loss_coefficient": 0.00017282,
    "heat_transfer_coefficient": 0.002,
    "kiln_thermal_mass": 1.0,
    "material_thermal_mass": 0.18438,
    "steam_power": 0.0
  }
}
```

**Engine Parameters:** (all are required; all must be positive except
`steam_power`, which may be zero)

- **heater_power** (float): Energy added per tick while the heater relay is on
- **heat_loss_coefficient** (float): Rate of heat loss to environment
- **heat_transfer_coefficient** (float): Rate of heat transfer kiln→material
- **kiln_thermal_mass** (float): Thermal inertia of kiln (higher = slower temperature changes)
- **material_thermal_mass** (float): Thermal inertia of wood
- **steam_power** (float): Energy added per tick while the steam relay is on
  **and** the kiln is below the steam cutoff. Zero means steam is inert.

**steam_cutoff_temp** is not written in the file. The simulator injects it from
`controlunit.defaults.steam_ceiling` in `halko.cfg`, the same way it injects
`time_step` from `controlunit.tick_length`. The rule and the physics are one
belief - steam stops heating once the kiln is hot enough that the steam must
itself be raised to kiln temperature - so they read one number and cannot
drift apart.

**Behavior:**

- Differential equations model heat flow
- Realistic thermal inertia effects
- Material heats slower than the kiln (realistic lag)
- Steam heats only below the cutoff; the fan is not modelled at all

### 2. Thermodynamic Engine

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
- The most detailed model of the three

Detail is not the same as accuracy: none of its coefficients have been fitted
against a real kiln, so a `thermodynamic` run is a richer model, not a
better-calibrated one.

## Choosing a Physics Engine

All engines advance one simulated tick per real `controlunit.tick_length`, so
none of them is faster than another in wall-clock terms. What differs is how
long a *program* takes, because that is set by how quickly the modelled
material reaches each step's target.

| Config | Wall clock for a full program | Choose it to |
| ------ | ----------------------------- | ------------ |
| `fast` | ~90 min | Verify the FSM: step order, transitions, completion conditions, failure handling |
| `tuning` | as long as the real kiln takes | Tune program targets, deltas and runtimes against realistic behaviour |
| `thermodynamic` | as long as the real kiln takes | Exercise the detailed model |

`fast` is not a physically plausible kiln and is not meant to be: it is an
extremely responsive one, chosen so a whole program fits in a working session.
Never read a tuning conclusion off a `fast` run.

Only the acclimate regime of `simulator-tuning.conf` has been fitted against a
real run. Its ramp rate and heater ceiling are still unconstrained guesses, so
treat ramp timings from it as indicative until a full real run is available to
refit them.

### Why `fast` is tuned the way it is

A relay only adopts a new state on a cycle boundary - every ten simulator
ticks, matching the power unit's `cycle_length`, so 60s at the shipped 6s tick.
Worse, two such quantizations sit in series: the power unit only switches a
device *on* at the start of its own 60s cycle (`processTick`, tick 0), and the
emulated relay then latches on its own boundary. An ON command can therefore
take up to 120s to reach the physics.
That quantum is real, not a simulator artifact: the deployed power unit
switches on the same cycle. It puts a floor under how tightly any delta band
can be held, because the kiln keeps moving for up to a minute after the
controller has asked it to stop.

That floor is what sets `fast`'s coefficients, by way of the acclimate step.
The acclimate controller does not hold the kiln in one fixed window. It bands
the kiln in `[target+minDelta, target]` while the material is at or above the
step target, and in `[material, material+maxDelta]` while the material is below
it, so the band jumps by most of its own width the moment the wood crosses the
target. That jump is the mechanism: wood reaches target, heater stops, wood
sags slightly, heater starts. The step is a test of that cycle.

For the cycle to happen at all the kiln has to be able to follow the band when
it jumps, and to still be moving slowly enough that up to 120s of actuation
delay does not carry it clean past. With `heat_loss_coefficient` at 0.0025 against
`kiln_thermal_mass` 1.0 the kiln drifts 2.5 C/min at 120 C, and
`heater_power` 0.5 makes the heating half-cycle the same 2.5 C/min. The
equilibrium ceiling is 220 C, well clear of the program's 120 C target.

Brisk is what you want here. `example/example-program-fast.json` holds for 30
minutes with a realistic `-1 .. +5` band, and a measured run cycles the relay
about eight times over that hold, the kiln sawing between roughly 116 C and
126 C with a period near eight minutes. Both branches of the band get exercised
three or four times each. The point of the step is the number of decisions the
controller is made to take, not how steady the temperature looks - a slower
kiln holds a tidier line and tests less.

The swing is wide because of that 120s delay, not in spite of it: the kiln runs
about 5 C past each threshold before the relay catches up. Anything that shrinks
the delay would tighten the sawtooth and raise the cycle count together.

An earlier and much brisker tuning made this step useless: the kiln bled
6.4 C/min against the same 5 C band, so the actuation delay carried it clean
across and out the other side. A 5 minute hold ended with the kiln 21 C below target and still
falling, having managed exactly one state change whose effect was never
observed before the step ran out. If you retune `fast`, check that the acclimate
step still cycles: a kiln that drifts much faster than about 3 C/min at the hold
temperature will outrun the band between relay boundaries.

The trade-off is that a slow-bleeding kiln also cools slowly. `fast` is matched
to `example/example-program-fast.json`, whose cooling step only drops the
material from 120 C to 110 C. Pointing a cooling step at, say, 40 C with these
coefficients takes well over an hour - the cooling step is there to show that
the program ends when the target is reached, not to model cooling.

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
./bin/simulator -c halko.cfg -s simulator-fast.conf -fail-sensors

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

1. **Start with `fast`**: verify the program's logic and step flow in about an hour
2. **Then `tuning`**: tune delta parameters and runtimes against fitted behaviour
3. **`example/example-program-fast.json`**: the four-step program the `fast` timings assume. Copy it into `fsdb/programs/` by hand; nothing copies it for you, and no example program is installed into a production environment
4. **Automatic Timing**: Simulator automatically uses `controlunit.tick_length` from `halko.cfg` for realistic timing
5. **Realistic Initial Conditions**: Set `initial_*_temp` to room temperature (20°C) for realistic startup
6. **Status Interval**: Use status logging during development, disable for performance testing

## Limitations

- No moisture content modeling (only temperature)
- No humidity effects on heat transfer
- No wood shrinkage or cracking simulation
- The fan is modelled only by `thermodynamic`, as forced convection and motor
  waste heat. `differential` ignores it entirely, so a step that commands the
  fan - a cooling step, or the equalization step's kiln-hotter branch - shows
  the state machine and the commanding, not the fan's thermal effect
- Steam is modelled only by `differential`, as a fixed energy input below the
  steam cutoff, and only where a config gives it a non-zero `steam_power`.
  `simulator-tuning.conf` sets it to zero: nothing has been fitted against a
  real run, so the fitted config stays honest about knowing nothing here.
  `simulator-fast.conf` sets 0.15, which is a guess - roughly a third of its
  heater. The steam warm-up step therefore *runs*, but how long it takes is an
  artifact of that guess and means nothing about real hardware. Only a real run
  can say what the generator needs
- A consequence of steam being a fixed input while the kiln's loss to ambient
  grows with temperature: the gap steam can open above the wood shrinks as the
  kiln warms. At `fast`'s 0.15 it is 3.1 C with the kiln at 22 C but only 1.6 C
  at 50 C, so a steam warm-up entered above roughly 46 C can never reach the
  2 C it waits for and will time out. Starting a program with a warm kiln - or
  running `simulator-equalize.conf`, which hands over near 33 C - is where that
  bites. Real steam carries far more latent heat than this model gives it

These limitations make the simulator suitable for temperature control development but not for predicting actual drying outcomes.

## See Also

- [API.md](API.md) - Complete API endpoint documentation
- [PROGRAM.md](PROGRAM.md) - Program structure and validation rules
- [README.md](README.md) - System overview and deployment guide

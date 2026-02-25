# Kiln Drying Programs

This document explains the structure and behavior of kiln drying programs in the
Halko system. Programs define the complete drying process, including temperature
targets, timing, and power control strategies.

## Program Structure

A kiln drying program is defined as a JSON file with the following structure:

```json
{
  "name": "Program Name",
  "steps": [
    {
      "name": "Step Name",
      "type": "heating|acclimate|cooling",
      "temperature_target": 100,
      "runtime": "6h",
      "heater": { /* power control settings */ },
      "fan": { /* power control settings */ },
      "humidifier": { /* power control settings */ }
    }
  ]
}
```

## Step Types

### Heating Steps

- **Purpose**: Raise the kiln temperature to target levels
- **Behavior**:
  - No fixed runtime - continues until target temperature is reached
  - Can use any power control method for the heater
  - Progresses to next step when oven temperature reaches target
- **Validation**: Runtime must not be specified

### Acclimate Steps

- **Purpose**: Maintain stable conditions for wood moisture equilibration
- **Behavior**:
  - Fixed duration specified by `runtime`
  - Maintains target temperature using specified control method
  - Progresses to next step when runtime expires
- **Validation**: Runtime is required

### Cooling Steps

- **Purpose**: Reduce kiln temperature in controlled manner
- **Behavior**:
  - Can have both target temperature and runtime specified
  - Progresses when either target temperature is reached OR runtime expires
    (whichever comes first)
  - Heater must use simple power control (typically 0% power)
  - Typically the final step in a program
- **Validation**: Runtime is optional, heater must use simple power

## Power Control Methods

Each component (heater, fan, humidifier) uses one of three power control methods:

### Simple Power Control

Maintains constant power output.

```json
{
  "power": 75
}
```

- **power**: Percentage (0-100) of maximum power
- **Usage**: Required for fan and humidifier, optional for heater
- **Behavior**: Outputs constant power regardless of temperature

### Delta Control

Maintains temperature difference between oven and wood.

```json
{
  "min_delta": 5.0,
  "max_delta": 15.0
}
```

- **max_delta**: Maximum temperature difference (oven - wood) in degrees
- **min_delta**: Minimum temperature difference (oven - wood) in degrees
- **Usage**: Heater only, primarily for heating steps
- **Behavior**:
  - Full power (100%) when oven temperature is below calculated target
  - Zero power (0%) when oven temperature is above calculated target
  - Target oven temperature = min(program_target, wood_temp + max_delta,
    max(wood_temp + min_delta))

### PID Control

Uses PID algorithm for precise temperature control.

```json
{
  "pid": {
    "kp": 2.0,
    "ki": 1.0,
    "kd": 0.5
  }
}
```

- **kp**: Proportional gain coefficient
- **ki**: Integral gain coefficient
- **kd**: Derivative gain coefficient
- **Usage**: Heater only, typically for acclimate steps
- **Behavior**: Calculates power adjustments based on temperature error

## Runtime Format

The `runtime` field uses Go's duration string format:

- `"6h"` - 6 hours
- `"30m"` - 30 minutes
- `"2h30m"` - 2 hours and 30 minutes
- `"45s"` - 45 seconds

## Program Validation Rules

### Step Order Requirements

1. **First step** must be a heating step
2. **Last step** must be a cooling step
3. **Minimum 2 steps** required

### Temperature Progression

- **Heating steps**: Target temperature must be higher than previous step
- **Acclimate steps**: Target temperature must be greater than or equal to
  previous step
- **Cooling steps**: Target temperature must be lower than previous step
- **Maximum temperature**: 200°C limit for all steps

### Component Restrictions

- **Fan**: Must always use simple power control
- **Humidifier**: Must always use simple power control
- **Heater in cooling steps**: Must use simple power control

## Example Programs

### Basic Delta Control Program

```json
{
  "name": "Basic Delta Drying",
  "steps": [
    {
      "name": "Initial Heating",
      "type": "heating",
      "temperature_target": 100,
      "heater": {
        "min_delta": 5.0,
        "max_delta": 15.0
      },
      "fan": {"power": 100},
      "humidifier": {"power": 50}
    },
    {
      "name": "Acclimation",
      "type": "acclimate",
      "temperature_target": 100,
      "runtime": "6h",
      "heater": {
        "min_delta": 2.0,
        "max_delta": 5.0
      },
      "fan": {"power": 75},
      "humidifier": {"power": 25}
    },
    {
      "name": "Cool Down",
      "type": "cooling",
      "temperature_target": 25,
      "runtime": "12h",
      "heater": {"power": 0},
      "fan": {"power": 100},
      "humidifier": {"power": 0}
    }
  ]
}
```

### PID Control Program

```json
{
  "name": "PID Controlled Drying",
  "steps": [
    {
      "name": "Heat Up",
      "type": "heating",
      "temperature_target": 80,
      "heater": {
        "min_delta": 10.0,
        "max_delta": 20.0
      },
      "fan": {"power": 100},
      "humidifier": {"power": 75}
    },
    {
      "name": "PID Acclimation",
      "type": "acclimate",
      "temperature_target": 80,
      "runtime": "8h",
      "heater": {
        "pid": {
          "kp": 2.0,
          "ki": 1.0,
          "kd": 0.5
        }
      },
      "fan": {"power": 50},
      "humidifier": {"power": 25}
    },
    {
      "name": "Cool Down",
      "type": "cooling",
      "temperature_target": 30,
      "runtime": "8h",
      "heater": {"power": 0},
      "fan": {"power": 100},
      "humidifier": {"power": 0}
    }
  ]
}
```

## Control Logic Details

### Delta Control Algorithm

The delta control method maintains the temperature difference between oven and
wood within specified bounds:

1. **Calculate target oven temperature**:
   - Start with program step target temperature
   - Apply max delta constraint: `min(target, wood_temp + max_delta)`
   - Apply min delta constraint: `max(result, wood_temp + min_delta)`

2. **Power decision**:
   - If oven temperature < target: 100% power
   - If oven temperature ≥ target: 0% power

### PID Control Algorithm

The PID controller calculates power adjustments based on temperature error:

1. **Error calculation**: `error = target_temp - actual_temp`
2. **PID terms**:
  - Proportional: `kp * error`
  - Integral: `ki * accumulated_error`
  - Derivative: `kd * error_rate_of_change`
3. **Power adjustment**: `current_power + (proportional + integral + derivative)`
4. **Clamping**: Final power limited to 0-100% range

## Default Settings

The system applies default settings for components when not specified in the program:

- **Fan power**: 0% (off)
- **Humidifier power**: 0% (off)
- **Heater settings**: Based on step type
  - Heating steps: Use configured delta defaults
  - Acclimate steps: Use configured PID defaults
  - Cooling steps: 0% power

These defaults are defined in the main configuration file under `controlunit.defaults`.

## Program Execution Flow

1. **Load program**: Parse JSON and validate structure
2. **Apply defaults**: Fill in missing component settings
3. **Validate program**: Check all rules and constraints
4. **Start execution**: FSM initializes and waits for initial sensor readings
5. **Pre-heat phase** (automatic):
  - If material temperature > oven temperature:
    - Heater set to 100% power
    - Fan set to 50% power
    - Continues until oven reaches material temperature
  - If oven temperature ≥ material temperature:
    - Pre-heat phase is skipped
    - Proceeds directly to first program step
  - **Purpose**: Prevents thermal shock by ensuring oven doesn't start colder than wood
6. **Execute steps sequentially**:
  - Initialize power controllers for current step
  - Monitor temperatures continuously
  - Update power outputs based on control algorithms
  - Check step completion conditions
  - Progress to next step when conditions met
7. **Complete**: Program ends when the final step (typically cooling) completes

### FSM State Machine

The controlunit uses a finite state machine (FSM) with the following states:

1. **start** - Initial state, sets timestamps and initializes
2. **waiting** - Waits for temperature and PSU status updates before proceeding
3. **preheat** - Automatic pre-heat if material warmer than oven (see above)
4. **next_program_step** - Increments step counter, determines next state from step type
5. **heat_up** - Execute heating step logic
6. **acclimate** - Execute acclimation step logic
7. **cool_down** - Execute cooling step logic
8. **idle** - Program completed successfully
9. **failed** - Error state

The FSM operates on a tick-based system with the update frequency controlled by
`controlunit.tick_length` in the configuration file (default 6 seconds).

### Execution Logging

The controlunit maintains detailed logs of temperature readings, power outputs, and
step transitions throughout the execution process. The program completes when the
last step finishes executing, either by reaching its target temperature, runtime
expiring, or both conditions being met (whichever comes first).

Execution logs are stored in CSV format at `{base_path}/running/` during execution
and moved to `{base_path}/history/logs/` upon completion.

# Kiln Drying Programs

This document explains the structure and behavior of kiln drying programs in the
Halko system. Programs define the complete drying process, including temperature
targets, timing, and power control strategies.

## Program Structure

A kiln drying program is defined as a JSON file with the following structure:

```jsonc
{
  "name": "Program Name",
  "description": "Optional free text about the program",
  "steps": [
    {
      "name": "Step Name",
      "type": "heating|acclimate|cooling",
      "temperature_target": 100,
      "runtime": "6h",
      "heater": { /* power control settings */ },
      "fan": { /* power control settings */ },
      "steam": { /* power control settings */ }
    }
  ]
}
```

The optional `description` is free text the operator writes about the program —
what charge it suits, why the steps are shaped the way they are. Nothing in the
system interprets it: it is carried into a run and its history alongside the
program, and shown in the program list, the `validate --verbose` output and the
PDF run report. Programs without one are unaffected.

## Step Types

### Heating Steps

- **Purpose**: Raise the material temperature to target levels
- **Behavior**:
  - No fixed runtime - continues until target temperature is reached
  - Progresses to next step when **material** temperature reaches target
- **Validation**:
  - Runtime must not be specified
  - Heater **must** use delta control (see [Why the heater must be
    delta](#why-the-heater-must-be-delta))
  - Steam must use simple or delta control, and simple non-zero steam is only
    allowed above the steam ceiling (see [The steam
    ceiling](#the-steam-ceiling))

### Acclimate Steps

- **Purpose**: Maintain stable conditions for wood moisture equilibration
- **Behavior**:
  - Fixed duration specified by `runtime`
  - Maintains target temperature using specified control method
  - Progresses to next step when runtime expires
- **Validation**:
  - Runtime is required
  - Heater may use any control method
  - Steam must use simple control, and must be 0% when the step's target is
    below the steam ceiling

### Cooling Steps

- **Purpose**: Reduce kiln temperature in controlled manner
- **Behavior**:
  - Optional runtime - can specify duration, target, or both
  - Progresses when material temperature reaches target OR runtime expires (whichever comes first)
  - Typically the final step in a program
- **Validation**:
  - Runtime is optional
  - Heater must use simple power control (typically 0% power)
  - Steam must be switched off (simple control at 0%)

## Power Control Methods

Each component (heater, fan, steam) uses one of three power control methods:

### Simple Power Control

Maintains constant power output.

```json
{
  "power": 75
}
```

- **power**: Percentage (0-100) of maximum power
- **Usage**: Required for the fan; required for the heater in cooling steps;
  required for steam except in heating steps, where delta is also allowed
- **Behavior**: Outputs constant power regardless of temperature

### Delta Control

Bands the kiln temperature. What the band is measured against depends on the
step type.

```json
{
  "min_delta": 5.0,
  "max_delta": 15.0
}
```

**Heating steps** band the kiln against the **wood**: the heater runs while the
kiln is at or below `wood + min_delta` and stops at `wood + max_delta`. Both
values are positive. This is the bound that keeps the kiln from running far
enough ahead of the wood to cause checking and case-hardening.

**Acclimate steps** band the kiln against whichever job the heater is doing, so
their deltas straddle zero:

| Condition | Kiln banded in | Job |
|-----------|----------------|-----|
| `wood >= target` | `[target + min_delta, target]` | holding the kiln at target |
| `wood < target` | `[wood, wood + max_delta]` | driving heat into the wood |

Here `min_delta` is negative — how far the kiln may sag below target before the
heater fires — and `max_delta` is positive, how far the kiln may lead the wood.
The two bands meet at the target, so when the wood arrives a kiln still up in
the heating band is above its new one and the heater stops.

- **Usage**: heater in heating and acclimate steps; steam in heating steps only
- **Behavior**: full power (100%) at or below the band's lower bound, zero
  power (0%) at or above its upper bound, and the previous state in between

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
3. **Minimum 3 steps** required (e.g., heating, acclimate, cooling)

### Runtime Requirements

- **Heating steps**: Cannot have runtime (progresses on temperature)
- **Acclimate steps**: Must have runtime (fixed duration)
- **Cooling steps**: Runtime is optional (progresses on temperature, timeout, or both)

### Temperature Progression

- **Heating steps**: Target temperature must be higher than previous step
- **Acclimate steps**: Target temperature must be greater than or equal to
  previous step; when directly following a heating step, it must equal the
  heating step's target temperature
- **Cooling steps**: Target temperature must be lower than previous step
- **Maximum temperature**: `defaults.max_target_temperature` limit for all steps

### Component Restrictions

Each component must define exactly one control method per step.

| Step | Heater | Fan | Steam |
|------|--------|-----|-------|
| heating | delta (required) | simple | simple or delta; simple must be 0% below the steam ceiling |
| acclimate | delta (required) | simple | simple; must be 0% when the target is below the steam ceiling |
| cooling | simple (required) | simple | simple, and must be 0% |

### Why the Heater Must Be Delta

Heating and acclimate steps both restrict the heater to delta control, because
only delta bounds the gap between the kiln and the wood. Simple control runs the
heater open-loop and lets the kiln run arbitrarily far ahead of the wood, which
is what causes checking and case-hardening.

### The Steam Ceiling

Steam cannot heat the kiln past the configured **steam ceiling**
(`defaults.steam_ceiling`, 100 °C as shipped): above that it is thermally neutral,
since it must itself be raised to kiln temperature. Below it, steam outruns the
heater and becomes the most immediate contributor to kiln temperature — opening
a kiln/wood delta that the heater cannot close by backing off, because it is
not the thing supplying the heat.

Steam may therefore only run open-loop (constant non-zero power) where the kiln
is already at or above the ceiling:

- **Heating steps**: allowed only if the kiln is already at or above 100 °C as
  the step *begins*. Each step is taken to enter where its predecessor handed
  over; the first step is taken to start cold, since nothing constrains the
  temperature a charge is loaded at. A heating step that enters below the
  ceiling must either switch steam off or put it under delta control.
- **Acclimate steps**: an acclimate settles the kiln back onto the material, so
  a step whose target is below 100 °C must switch steam off.
- **Cooling steps**: the kiln descends back through the ceiling, so steam must
  be off entirely.

A step's handover temperature is the lowest kiln temperature it can end at: for
a delta-controlled heating step that is `target + min_delta`, since the step
ends once the material reaches target and the controller holds the kiln at
least `min_delta` above it. For an acclimate it is the step target.

### Sensor Failsafe

Independently of program validation, the control unit fails a running program
and switches all power off if either the kiln or the material temperature goes
**120 seconds** without a valid reading. A failed probe does not immediately
disturb control — the last valid value per sensor is held, so a brief dropout
is ridden out — but the run is abandoned rather than flown blind past that
point.

## Example Programs

Runnable copies of both live in [`example/`](example/); validate any program
with `halkoctl validate <file>` before sending it.

### Basic Delta Control Program

Crosses the steam ceiling: the first step starts cold, so its steam runs under
delta control, while the second step enters at 105 °C and may hold steam at a
constant 50%.

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
      "steam": {
        "min_delta": 5.0,
        "max_delta": 10.0
      }
    },
    {
      "name": "Secondary Heating",
      "type": "heating",
      "temperature_target": 160,
      "heater": {
        "min_delta": 5.0,
        "max_delta": 15.0
      },
      "fan": {"power": 100},
      "steam": {"power": 50}
    },
    {
      "name": "Acclimation",
      "type": "acclimate",
      "temperature_target": 160,
      "runtime": "6h",
      "heater": {
        "min_delta": 2.0,
        "max_delta": 5.0
      },
      "fan": {"power": 75},
      "steam": {"power": 25}
    },
    {
      "name": "Cool Down",
      "type": "cooling",
      "temperature_target": 25,
      "runtime": "12h",
      "heater": {"power": 0},
      "fan": {"power": 100},
      "steam": {"power": 0}
    }
  ]
}
```

### Delta Control Algorithm

The delta control method maintains the temperature difference between kiln and
wood within specified bounds:

1. **Calculate target kiln temperature**:
   - Start with program step target temperature
   - Apply max delta constraint: `min(target, wood_temp + max_delta)`
   - Apply min delta constraint: `max(result, wood_temp + min_delta)`

2. **Power decision**:
   - If kiln temperature < target: 100% power
   - If kiln temperature ≥ target: 0% power

## Default Settings

The system applies default settings for components when not specified in the program:

- **Fan power**: `defaults.fan_power`
- **Steam power**: `defaults.steam_power`
- **Heater settings**: Based on step type
  - Heating steps: `defaults.deltas.heating`
  - Acclimate steps: `defaults.deltas.acclimate`
  - Cooling steps: 0% power — a cooling step never drives the heater, so
    there is nothing to configure

These defaults are defined in the main configuration file under `controlunit.defaults`.

## Program Execution Flow

1. **Load program**: Parse JSON and validate structure
2. **Apply defaults**: Fill in missing component settings
3. **Validate program**: Check all rules and constraints
4. **Start execution**: FSM initializes and waits for initial sensor readings
5. **Startup steps** (added by the control unit, configured by the program's
   `equalize` block):
   - **Equalize** — holds until the kiln and the material are within
     `equalize.delta` of each other, so the program starts from a known state.
     A kiln above the band is cooled by running the fan, which exhausts warm
     air and draws in ambient. A kiln below the band is waited out: the wood is
     the larger thermal mass and warms the air up to meet it. The step never
     drives the heater and has no timeout — it adds no heat, so waiting is safe.
   - **Steam warm-up** — only when `equalize.steam_prewarm` is set. Runs the
     steam generator with the heater and the fan off. The reservoir produces
     nothing until it boils, and nothing measures it, so the step infers it
     from the kiln: with the wood unable to push the air past its own
     temperature, a kiln climbing past the top of the band is being heated by
     the steam. Fails the run on `equalize.steam_prewarm_timeout`, which catches an
     empty reservoir or a dead element before the program depends on steam.
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
3. **next_program_step** - Increments step counter, determines next state from step type
4. **equalize** - Startup step, holds until the kiln and material are within the band
5. **steam_prewarm** - Optional startup step, proves the steam generator is producing
6. **heat_up** - Execute heating step logic
7. **acclimate** - Execute acclimation step logic
8. **cool_down** - Execute cooling step logic
9. **idle** - Program completed successfully
10. **failed** - Error state

The FSM operates on a tick-based system with the update frequency controlled by
`controlunit.tick_length` in the configuration file (e.g., "6s").

### Execution Logging

The controlunit maintains detailed logs of temperature readings, power outputs, and
step transitions throughout the execution process. The program completes when the
last step finishes executing, either by reaching its target temperature, runtime
expiring, or both conditions being met (whichever comes first).

Execution logs are stored in CSV format at `{base_path}/running/` during execution
and moved to `{base_path}/history/logs/` upon completion.

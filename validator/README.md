# Halko Program Validator

A command-line tool to validate Halko program.json files against the schema and business rules.

## Overview

This tool validates program files by:

1. Loading the Halko configuration (including defaults)
2. Parsing the program JSON file
3. Applying default values according to the configuration
4. Running comprehensive validation checks including:
   - JSON schema validation
   - Business rule validation (step order, temperature progression, etc.)
   - Power control method validation
   - Runtime requirements validation

## Usage

```bash
# Basic usage with default config
./validator -program example/example-program-delta.json

# Specify custom config file
./validator -program my-program.json -config my-halko.cfg

# Verbose output
./validator -program example/example-program-delta.json -verbose

# Show help
./validator -help
```

## Options

- `-program`: Path to the program.json file to validate (required)
- `-config`: Path to the halko.cfg file (optional, searches in order: /etc/opt/halko/halko.cfg, templates/halko.cfg, ../templates/halko.cfg)
- `-verbose`: Enable verbose output showing validation steps
- `-help`: Show help message

## Building

From the validator directory:

```bash
go build -o validator main.go
```

Or from the project root:

```bash
go build -o bin/validator ./validator
```

## Validation Rules

The validator checks:

### Program Structure

- Must have at least 2 steps
- First step must be a heating step
- Last step must be a cooling step
- Temperature targets must not exceed 200°C

### Step Types and Rules

- **Heating steps**: Cannot have runtime, must increase temperature
- **Acclimate steps**: Must have runtime, temperature ≥ previous step
- **Cooling steps**: Cannot have runtime, must decrease temperature, heater must use simple power control

### Power Control

- Each component (heater/fan/humidifier) must use exactly one control method:
  - Simple power (0-100%)
  - Delta control (min/max delta values)
  - PID control (kp, ki, kd values)
- Fan and humidifier must always use simple power control
- Heater can use any control method depending on step type

### Temperature Progression

- Heating steps must have higher temperature than previous step
- Acclimate steps must have temperature ≥ previous step
- Cooling steps must have lower temperature than previous step

## Exit Codes

- `0`: Validation successful
- `1`: Validation failed or error occurred

## Example Output

```text
$ ./validator -program example/example-program-delta.json -verbose
Validating program: example/example-program-delta.json
Using config: templates/halko.cfg

Loading configuration from: templates/halko.cfg
✓ Configuration loaded successfully
Loading program from: example/example-program-delta.json
✓ Program file loaded successfully
Parsing JSON...
✓ JSON parsed successfully - Program: 'Four-Stage Kiln Drying Program using delta acclimation' with 4 steps
Creating program copy for validation...
✓ Program copy created
Applying defaults...
✓ Defaults applied successfully
Running validation...
✓ Program validation completed successfully

Program structure:
  Name: Four-Stage Kiln Drying Program using delta acclimation
  Steps: 4
    1. Initial Heating (heating) - Target: 100°C
    2. Secondary Heating (heating) - Target: 150°C
    3. Acclimation Phase (acclimate) - Target: 160°C
       Runtime: 6h0m0s
    4. Cooling Phase (cooling) - Target: 30°C

✓ Program validation successful!
```

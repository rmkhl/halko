# halkoctl - Halko Control Tool

A command-line tool for interacting with the Halko wood drying kiln control system.

## Usage

```bash
halkoctl [-c|--config <config-file>] <command> [options]
```

## Global Options

- `-c, --config string`: Path to halko.cfg configuration file (optional, auto-discovers if not specified)

## Commands

### send

Sends a program.json file to the Halko executor to start program execution.

```bash
halkoctl send <program-file> [options]
```

#### Send Arguments

- `program-file`: Path to the program.json file to send (required)

#### Send Options

- `-v, --verbose`: Enable verbose output
- `-h, --help`: Show help for send command

#### Send Examples

Send a program using default config:

```bash
halkoctl send example/example-program-delta.json
```

Send a program with custom config and verbose output:

```bash
halkoctl --config /path/to/halko.cfg send my-program.json -v
```

### status

Gets the status of the currently running program from the Halko executor.

```bash
halkoctl status [options]
```

#### Status Options

- `-v, --verbose`: Enable verbose output
- `-h, --help`: Show help for status command

#### Status Examples

Get status using default config:

```bash
halkoctl status
```

Get status with custom config and verbose output:

```bash
halkoctl --config /path/to/halko.cfg status -v
```

### validate

Validates a program.json file against the Halko program schema and business rules.

```bash
halkoctl validate <program-file> [options]
```

#### Validate Arguments

- `program-file`: Path to the program.json file to validate (required)

#### Validate Options

- `-v, --verbose`: Enable verbose output
- `-h, --help`: Show help for validate command

#### Validate Examples

Validate a program using default config:

```bash
halkoctl validate example/example-program-delta.json
```

Validate a program with custom config and verbose output:

```bash
halkoctl --config /path/to/halko.cfg validate my-program.json --verbose
```

## API Endpoints

### send command

The `send` command sends a POST request to the executor's `/engine/api/v1/running` endpoint with the program definition in the request body:

```json
{
  "program": {
    "name": "Program Name",
    "steps": [...]
  }
}
```

### status command

The `status` command sends a GET request to the executor's `/engine/api/v1/running` endpoint and displays the response in a user-friendly format.

Example response when a program is running:

```json
{
  "data": {
    "status": "running",
    "program": {
      "name": "Standard Drying",
      "currentPhase": "Initial Heating",
      "elapsedTime": 1200,
      "currentTemperature": 42.5,
      "targetTemperature": 45,
      "remainingTime": 2400
    }
  }
}
```

Example response when no program is running:

```json
{
  "data": {
    "status": "idle"
  }
}
```

## Configuration

halkoctl uses the Halko configuration file (`halko.cfg`) to determine API endpoints. The tool automatically searches for the config file in these locations:

1. `halko.cfg` (current directory)
2. `templates/halko.cfg`
3. `/etc/halko/halko.cfg`
4. `/var/opt/halko/halko.cfg`
5. `../templates/halko.cfg` (relative to executable)

You can also specify a custom config file using the `-config` flag.

## Notes

- The `validate` command performs local validation using the configuration defaults
- The `send` command sends the program to the executor which also performs validation
- The executor must be running and accessible at the URL specified in the config file
- Programs start execution immediately upon successful submission via `send`
- Use the verbose flag to see detailed HTTP request/response information and validation details

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

Sends a program.json file to the ControlUnit to start program execution.

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

---

### status

Gets the overall status of the ControlUnit service (health and running program info).

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

---

### running

Shows information about the currently running program.

```bash
halkoctl running [options]
```

Displays program name, current step, elapsed time, temperatures, and power status.

#### Running Options

- `-v, --verbose`: Enable verbose output for HTTP requests
- `-h, --help`: Show help for running command

#### Running Examples

```bash
halkoctl running                    # Show current program
halkoctl --verbose running          # Show with verbose HTTP output
```

---

### stream

Connects to the live execution log WebSocket and displays messages in real-time.

```bash
halkoctl stream
```

Useful for debugging to see exactly what data the WebSocket sends. Press Ctrl+C to stop.

#### Stream Examples

```bash
halkoctl stream
halkoctl --config /path/to/halko.cfg stream
```

---

### history

Manages program execution history.

```bash
halkoctl history <subcommand> [arguments]
```

#### History Subcommands

- `list` - List all executed programs
- `show <program-name>` - Show detailed information about a specific program run
- `log <program-name> [-o output-file]` - Display the execution log for a program run

#### History Options

- `-o, --output string` - (log subcommand only) Write log output to specified file
- `-h, --help` - Show help message

#### History Examples

```bash
halkoctl history list
halkoctl history show "My Program@2024-01-15T10:30:00Z"
halkoctl history log "My Program@2024-01-15T10:30:00Z"
halkoctl history log "My Program@2024-01-15T10:30:00Z" -o logfile.csv
```

---

### temperatures

Gets the current temperature readings from the sensor unit.

```bash
halkoctl temperatures [options]
```

Retrieves readings from the sensor unit's GET `/temperatures` endpoint.

#### Temperatures Options

- `-v, --verbose`: Enable verbose output
- `-h, --help`: Show help message

#### Temperatures Examples

```bash
halkoctl temperatures
halkoctl --config /path/to/halko.cfg temperatures
halkoctl --verbose temperatures
```

---

### display

Sends a text message to the sensor unit LCD display.

```bash
halkoctl display <message> [options]
```

#### Display Arguments

- `message` - Text message to display on the sensor unit LCD (required)

#### Display Options

- `-v, --verbose`: Enable verbose output
- `-h, --help`: Show help message

#### Display Examples

```bash
halkoctl display "Hello World"
halkoctl --config /path/to/halko.cfg display "Temperature: 25°C"
halkoctl --verbose display "System Ready"
```

---

### programs

Manages programs stored in the ControlUnit storage.

```bash
halkoctl programs <subcommand> [arguments]
```

#### Programs Subcommands

- `list` - List all stored programs
- `get <program-name>` - Get a specific program
- `create <program-file>` - Create a new program from JSON file
- `update <name> <file>` - Update existing program with new content
- `delete <program-name>` - Delete a stored program

#### Programs Examples

```bash
halkoctl programs list
halkoctl programs get my-program
halkoctl programs create example/example-program-delta.json
halkoctl programs update my-program updated-program.json
halkoctl programs delete old-program
```

**Notes:**

- Program files must be valid JSON
- Program names are derived from filenames if not specified in JSON
- Use `--verbose` for detailed operation information

---

### nginx

Generates an nginx configuration file for proxying Halko services.

```bash
halkoctl nginx [options]
```

#### Nginx Options

- `-port int` - Port for nginx to listen on (default: 80)
- `-output string` - Output file path (default: stdout)
- `-h, --help` - Show help message

#### Nginx Examples

```bash
halkoctl nginx -port 8080
halkoctl nginx -port 80 -output /etc/nginx/sites-available/halko
halkoctl -c halko-docker.cfg nginx -port 80 -output webapp/nginx-docker.conf
```

The generated configuration includes:

- API proxy routes for all services
- WebSocket support for live log streaming
- CORS headers
- Static file serving for the webapp

---

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

halkoctl interacts with various ControlUnit and SensorUnit endpoints:

### send command

Sends a POST request to the ControlUnit's `/engine/running` endpoint with the program definition in the request body:

```json
{
  "program": {
    "name": "Program Name",
    "steps": [...]
  }
}
```

### running command

Sends a GET request to `/engine/running` and displays the current execution status.

**Response when program is running:**

```json
{
  "data": {
    "program": { /* program definition */ },
    "started_at": 1734007854,
    "current_step": "Initial Heating",
    "current_step_started_at": 1734007890,
    "temperatures": {
      "material": 42.5,
      "oven": 45.2
    },
    "power_status": {
      "heater": 75,
      "fan": 50,
      "humidifier": 0
    }
  }
}
```

**Response when no program running:** HTTP 204 No Content

### stream command

Connects to WebSocket endpoint `/engine/running/logws` for real-time log streaming.

### history command

Interacts with:

- `GET /engine/history` - List all runs
- `GET /engine/history/{name}` - Get specific run details
- `GET /engine/history/{name}/log` - Get execution log CSV

### programs command

Manages stored programs via:

- `GET /programs` - List all stored programs
- `GET /programs/{name}` - Get specific program
- `POST /programs` - Create new program
- `POST /programs/{name}` - Update existing program
- `DELETE /programs/{name}` - Delete program

### status command

Gets service health from `GET /status` endpoint.

### temperatures command

Gets sensor readings from SensorUnit's `GET /temperatures` endpoint.

### display command

Sends message to SensorUnit's `POST /display` endpoint.

## Configuration

halkoctl uses the Halko configuration file (`halko.cfg`) to determine API endpoints. The tool automatically searches for the config file in these locations (in order):

1. Path specified with `-c` or `--config` flag
2. `HALKO_CONFIG` environment variable
3. `./halko.cfg` (current directory)
4. `~/.halko.cfg` (user home directory)
5. `/etc/opt/halko.cfg` (system-wide)

You can also specify a custom config file using the `-c` or `--config` flag.

## Notes

- The `validate` command performs local validation using the configuration defaults
- The `send` command sends the program to the ControlUnit which also performs validation
- Programs start execution immediately upon successful submission via `send`
- The ControlUnit must be running and accessible at the URL specified in the config file
- Use the `--verbose` flag to see detailed HTTP request/response information
- The `stream` command connects via WebSocket for real-time log updates
- History command uses timestamps in program names (e.g., "Program@2024-01-15T10:30:00Z")

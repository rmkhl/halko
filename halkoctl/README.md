# halkoctl - Halko Control Tool

A command-line tool for interacting with the Halko wood drying kiln control system.

## Usage

```bash
halkoctl <command> [options]
```

## Commands

### send

Sends a program.json file to the Halko executor to start program execution.

```bash
halkoctl send -program <path-to-program.json> [options]
```

#### Options

- `-program string`: Path to the program.json file to send (required)
- `-host string`: Executor host (default: localhost)
- `-port string`: Executor port (default: 8080)
- `-verbose`: Enable verbose output
- `-help`: Show help for send command

#### Examples

Send a program to the local executor:
```bash
halkoctl send -program example/example-program-delta.json
```

Send a program to a remote executor with verbose output:

```bash
halkoctl send -program my-program.json -host 192.168.1.100 -port 8080 -verbose
```

### status

Gets the status of the currently running program from the Halko executor.

```bash
halkoctl status [options]
```

#### Options

- `-host string`: Executor host (default: localhost)
- `-port string`: Executor port (default: 8080)
- `-verbose`: Enable verbose output
- `-help`: Show help for status command

#### Examples

Get status from the local executor:
```bash
halkoctl status
```

Get status from a remote executor with verbose output:
```bash
halkoctl status -host 192.168.1.100 -port 8080 -verbose
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

## Notes

- The tool does not validate the program locally - the executor will perform validation
- The executor must be running and accessible at the specified host:port
- The program will start execution immediately upon successful submission
- Use the verbose flag to see detailed HTTP request/response information

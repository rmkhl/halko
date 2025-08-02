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

## API Endpoint

The `send` command sends a POST request to the executor's `/engine/api/v1/running` endpoint with the program definition in the request body:

```json
{
  "program": {
    "name": "Program Name",
    "steps": [...]
  }
}
```

## Notes

- The tool does not validate the program locally - the executor will perform validation
- The executor must be running and accessible at the specified host:port
- The program will start execution immediately upon successful submission
- Use the verbose flag to see detailed HTTP request/response information

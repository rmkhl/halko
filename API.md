# Halko API Documentation

This document provides detailed information about the REST APIs exposed by each
component of the Halko wood drying kiln control system.

## Common Response Patterns

All Halko services use a consistent JSON response structure:

```json
{
  "data": {
    // Service-specific data
  }
}
```

### Standardized Status Endpoint

All services implement a `/status` endpoint that returns health information using this format:

```json
{
  "data": {
    "status": "healthy",
    "service": "controlunit",
    "details": {
      // Service-specific status details
    }
  }
}
```

**Status Values:**

- `healthy`: Service is operating normally
- `degraded`: Service is operational but experiencing issues
- `unavailable`: Service is not operational

**Service-Specific Details:**

- **ControlUnit**: `program_running` (bool), `current_step` (int), `started_at` (string)
- **PowerUnit**: `controller_initialized` (bool)
- **Storage**: `accessible` (bool), `error` (string if not accessible)
- **SensorUnit**: `arduino_connected` (bool)

## 1. SensorUnit API

Base Path: `/sensors`

### Temperature Endpoints

#### GET `/temperatures`

Fetches current temperature readings from all sensors.

**Response Format:**

```json
{
  "data": {
    "oven": 45.2,
    "wood": 32.1,
    "ovenPrimary": 45.2,
    "ovenSecondary": 44.8
  }
}
```

The response includes:

- `oven`: The highest of the two oven temperatures or a single oven
  temperature if one sensor is unavailable
- `wood`: The current wood temperature
- `ovenPrimary`: The primary oven temperature sensor reading
- `ovenSecondary`: The secondary oven temperature sensor reading

### Status Endpoints

#### GET `/status`

Checks the connection status of the sensor unit. Follows the standard status endpoint format (see Common Response Patterns).

**Response Format:**

```json
{
  "data": {
    "status": "healthy",
    "service": "sensorunit",
    "details": {
      "arduino_connected": true
    }
  }
}
```

**Details:**

- `arduino_connected`: Boolean indicating if the Arduino device is connected via USB serial
- `disconnected`: The sensor unit is not connected or not responding

#### POST `/display`

Updates the text displayed on the sensor unit's LCD.

**Request Format:**

```json
{
  "message": "Drying in progress"
}
```

**Response Format:**

```json
{
  "data": {
    "status": "ok"
  }
}
```

## 2. PowerUnit API

Base Path: `/powers`

### Status Endpoints

#### GET `/status`

Gets the health status of the PowerUnit service. Follows the standard status endpoint format (see Common Response Patterns).

**Response Format:**

```json
{
  "data": {
    "status": "healthy",
    "service": "powerunit",
    "details": {
      "controller_initialized": true
    }
  }
}
```

**Details:**

- `controller_initialized`: Boolean indicating if the power controller (Shelly device interface) is properly initialized

### Power Control Endpoints

### GET `/`

Gets the status of all power channels.

**Response Format:**

```json
{
  "data": {
    "heater": {
      "percent": 100
    },
    "fan": {
      "percent": 100
    },
    "humidifier": {
      "percent": 0
    }
  }
}
```

### GET `/:power`

Gets the status of a specific power channel.

**Path Parameters:**

- `power`: The name of the power channel (e.g., `heater`, `fan`,
  `humidifier`)

**Response Format:**

```json
{
  "data": {
    "percent": 100
  }
}
```

### POST `/:power` (also supports PUT and PATCH)

Operates a specific power channel.

**Path Parameters:**

- `power`: The name of the power channel (e.g., `heater`, `fan`,
  `humidifier`)

**Request Format:**

```json
{
  "percent": 75
}
```

**Response Format:**

```json
{
  "data": {
    "percent": 75
  }
}
```

## 3. ControlUnit API

Base Path: `/engine` (execution management), `/programs` (stored program templates)
Default Port: `8090`

The ControlUnit serves both execution management and program storage endpoints:

- `/engine/*` - Program execution, history, and live monitoring
- `/programs/*` - Stored program template management (CRUD operations)
- `/status` - Service health status

**Note:** There is no separate "storage" service - all storage functionality is integrated into the ControlUnit.

### Status Endpoints

#### GET `/status`

Gets the health status of the ControlUnit service. Follows the standard status endpoint format (see Common Response Patterns).

**Response Format:**

```json
{
  "data": {
    "status": "healthy",
    "service": "controlunit",
    "details": {
      "program_running": true,
      "current_step": 3,
      "started_at": "2024-01-15T10:30:00Z"
    }
  }
}
```

**Details:**

- `program_running`: Boolean indicating if a program is currently executing
- `current_step`: Current step number in the program (only present if program is running)
- `started_at`: ISO 8601 timestamp of program start (only present if program is running)

### Program Execution Endpoints

These endpoints manage running programs and execution history.

#### GET `/engine/history`

Lists all completed program executions.

**Response Format:**

```json
{
  "data": [
    {
      "name": "Standard Drying@2024-01-15T10:30:00Z",
      "started_at": "2024-01-15T10:30:00Z",
      "completed_at": "2024-01-15T18:45:00Z"
    },
    {
      "name": "Quick Drying@2024-01-14T08:00:00Z",
      "started_at": "2024-01-14T08:00:00Z",
      "completed_at": "2024-01-14T14:30:00Z"
    }
  ]
}
```

#### GET `/engine/history/{name}`

Gets details of a specific completed program execution.

**Path Parameters:**

- `name`: The full name of the execution (includes timestamp, e.g., "Program@2024-01-15T10:30:00Z")

**Response Format:**

```json
{
  "data": {
    "program": { /* full program definition */ },
    "started_at": 1734007854,
    "completed_at": 1734037854,
    "status": "completed"
  }
}
```

#### GET `/engine/history/{name}/log`

Gets the execution log CSV for a completed program.

**Path Parameters:**

- `name`: The full name of the execution

**Response:** Returns CSV data directly (not JSON-wrapped)

```csv
timestamp,step_name,oven_temp,material_temp,heater_power,fan_power,humidifier_power
1734007890,Initial Heating,45.2,42.5,75,50,0
...
```

#### DELETE `/engine/history/{name}`

Deletes a completed program execution and its logs.

**Path Parameters:**

- `name`: The full name of the execution to delete

**Response:**

- Status 200 OK on success
- Status 404 Not Found if execution doesn't exist

#### GET `/engine/running`

Gets the status of the currently running program.

**Response Format:**

When a program is running:

```json
{
  "data": {
    "program": {
      "program_name": "Four-Stage Kiln Drying Program using delta acclimation",
      "program_steps": [
        {
          "name": "Initial Heating",
          "target_temperature": 50,
          "runtime": 7200
        }
      ]
    },
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

**Response Fields:**

- `program`: The complete program definition being executed
- `started_at`: Unix timestamp when program execution began
- `current_step`: Name of the currently executing step
- `current_step_started_at`: Unix timestamp when current step began
- `temperatures.material`: Current material (wood) temperature in °C
- `temperatures.oven`: Current oven temperature in °C
- `power_status.heater`: Heater power level (0-100%)
- `power_status.fan`: Fan power level (0-100%)
- `power_status.humidifier`: Humidifier power level (0-100%)

If no program is running, returns HTTP 204 No Content with error message:

```json
{
  "error": "No program running"
}
```

#### POST `/engine/running`

Starts a new program (by providing its definition or name).

**Request Format:**

```json
{
  "programName": "Standard Drying"
}
```

OR

```json
{
  "program": {
    "name": "Custom Program",
    "description": "One-time custom program",
    "phases": []
  }
}
```

**Response Format:**

```json
{
  "data": {
    "status": "started",
    "program": {
      "name": "Standard Drying"
    }
  }
}
```

#### DELETE `/engine/running`

Cancels the currently running program.

**Response:**

- Status 204 No Content on success

#### GET `/engine/running/log`

Fetches the accumulated execution log as CSV data for the currently running program.

**Response Format:**

When a program is running, returns CSV data:

```csv
timestamp,step_name,oven_temp,material_temp,heater_power,fan_power,humidifier_power
1734007890,Initial Heating,45.2,42.5,75,50,0
1734007896,Initial Heating,46.1,42.8,75,50,0
...
```

When no program is running, returns HTTP 204 No Content with error message:

```json
{
  "error": "No program running"
}
```

**Usage:** This endpoint is used by the webapp to fetch historical log data before connecting to the WebSocket for real-time updates.

#### WebSocket `/engine/running/logws`

WebSocket endpoint for real-time execution log streaming.

**Connection:** Upgrade HTTP connection to WebSocket at `ws://host:port/engine/running/logws`

**Message Format:**

The server sends CSV lines as text messages:

```
timestamp,step_name,oven_temp,material_temp,heater_power,fan_power,humidifier_power
```

**Behavior:**

- Sends CSV header line on connection
- Sends new log entries as they're generated (every tick)
- Closes connection when program completes or is cancelled
- Returns 204 error if no program is running

**Example using JavaScript:**

```javascript
const ws = new WebSocket('ws://localhost:8090/engine/running/logws');

ws.onmessage = (event) => {
  console.log('Log entry:', event.data);
};

ws.onclose = () => {
  console.log('Program execution ended');
};
```

**Note:** nginx proxy configurations must include WebSocket upgrade headers:

```nginx
proxy_set_header Upgrade $http_upgrade;
proxy_set_header Connection "upgrade";
```

#### GET `/engine/defaults`

Gets the default program settings configured in the ControlUnit.

**Response Format:**

```json
{
  "data": {
    "pid_settings": {
      "acclimate": {
        "kp": 2.0,
        "ki": 1.0,
        "kd": 0.5
      }
    },
    "max_delta_heating": 10.0,
    "min_delta_heating": 5.0
  }
}
```

**Usage:** These defaults are automatically applied to programs that don't specify complete power control settings for all steps.

### File-Based Storage

The ControlUnit maintains a file-based storage system with the following structure:

- `{base_path}/programs/` - Stored program templates (managed via `/programs` endpoints)
- `{base_path}/running/` - Active program execution files (JSON + TXT status + CSV log)
- `{base_path}/history/` - Completed program executions (JSON)
- `{base_path}/history/logs/` - Completed execution logs (CSV)
- `{base_path}/history/status/` - Completed program status files (TXT)

**Automatic File Management:**

When a program starts executing, files are created in `running/`. Upon completion (whether successful, failed, or canceled), these files are automatically moved to the appropriate `history/` subdirectories. On startup, the ControlUnit performs cleanup of any orphaned files in `running/` from previous crashes.

### Stored Program Template Endpoints

These endpoints manage stored program templates (not executions). Templates are stored in `{base_path}/programs/` and can be used to start new executions.

#### GET `/programs`

Lists all stored program templates with their last modification times.

**Response Format:**

```json
{
  "data": [
    {
      "name": "Standard Drying",
      "last_modified": "2023-12-17T14:30:00Z"
    },
    {
      "name": "Quick Drying",
      "last_modified": "2023-12-16T10:15:00Z"
    },
    {
      "name": "Pine Program",
      "last_modified": "2023-12-15T08:00:00Z"
    }
  ]
}
```

**Fields:**

- `name`: The name of the stored program template
- `last_modified`: ISO 8601 formatted timestamp of the last modification time

#### GET `/programs/{name}`

Gets a specific stored program template by name.

**Path Parameters:**

- `name`: The name of the program template

**Response Format:**

```json
{
  "data": {
    "name": "Standard Drying",
    "steps": [
      {
        "name": "Initial Heating",
        "type": "heating",
        "temperature_target": 50,
        "heater": { /* power control */ }
      }
    ]
  }
}
```

#### POST `/programs`

Creates a new stored program template.

**Request Body:**

```json
{
  "name": "New Program",
  "steps": [ /* program steps */ ]
}
```

**Response:**

- Status 201 Created on success
- Status 409 Conflict if program already exists

#### POST `/programs/{name}`

Updates an existing stored program template.

**Path Parameters:**

- `name`: The name of the program template to update

**Request Body:**

```json
{
  "name": "Updated Program",
  "steps": [ /* updated program steps */ ]
}
```

**Response:**

- Status 200 OK on success
- Status 404 Not Found if program doesn't exist

#### DELETE `/programs/{name}`

Deletes a stored program template.

**Path Parameters:**

- `name`: The name of the program template to delete

**Response:**

- Status 200 OK on success
- Status 404 Not Found if program doesn't exist

## 4. Simulator API

The simulator mimics endpoints from the SensorUnit and Shelly devices.

### Simulated SensorUnit API

Base Path: `/sensors`

#### Simulated GET `/temperatures`

Gets readings from all simulated temperature sensors.

**Response Format:**

```json
{
  "data": {
    "oven": 45.2,
    "wood": 32.1,
    "ovenPrimary": 45.2,
    "ovenSecondary": 44.8
  }
}
```

#### Simulated GET `/status`

Gets the simulated connection status.

**Response Format:**

```json
{
  "data": {
    "status": "connected"
  }
}
```

#### Simulated POST `/display`

Logs a display message (simulates updating an LCD).

**Request Format:**

```json
{
  "message": "Simulation in progress"
}
```

**Response Format:**

```json
{
  "data": {
    "status": "ok"
  }
}
```

### Simulated Shelly Switch Control

Base Path: `/rpc`

#### GET `/Switch.GetStatus`

Gets the status of simulated Shelly switches.

**Query Parameters:**

- `id`: The ID of the switch (e.g., `0`, `1`, `2`)

**Response Format:**

```json
{
  "id": 0,
  "source": "HTTP",
  "output": true,
  "timer_started": 0,
  "timer_duration": 0,
  "timer_remaining": 0
}
```

#### GET `/Switch.Set`

Sets the state of simulated Shelly switches.

**Query Parameters:**

- `id`: The ID of the switch (e.g., `0`, `1`, `2`)
- `on`: Boolean value to set the switch state (`true` or `false`)

**Response Format:**

```json
{
  "id": 0,
  "source": "HTTP",
  "output": true,
  "timer_started": 0,
  "timer_duration": 0,
  "timer_remaining": 0
}
```

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

Base Path: `/engine` and `/storage`
Default Port: `8090`

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

### Program Execution History Endpoints

#### GET `/programs`

Lists all available programs (definitions loaded by controlunit).

**Response Format:**

```json
{
  "data": [
    {
      "name": "Standard Drying",
      "description": "Standard drying program for pine",
      "phases": []
    },
    {
      "name": "Quick Drying",
      "description": "Quick drying program for thinner woods",
      "phases": []
    }
  ]
}
```

#### GET `/programs/:name`

Gets a specific program definition by name.

**Path Parameters:**

- `name`: The name of the program

**Response Format:**

```json
{
  "data": {
    "name": "Standard Drying",
    "description": "Standard drying program for pine",
    "phases": [
      {
        "name": "Initial Heating",
        "duration": 3600,
        "targetTemperature": 45,
        "minTemperature": 40,
        "maxTemperature": 50,
        "fanPower": 100,
        "humidifierPower": 0
      }
    ]
  }
}
```

#### DELETE `/programs/:name`

Deletes/unloads a specific program definition.

**Path Parameters:**

- `name`: The name of the program to delete

**Response:**

- Status 204 No Content on success

### Engine Control Endpoints

#### GET `/running`

Gets the status of the currently running program.

**Response Format:**

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

If no program is running:

```json
{
  "data": {
    "status": "idle"
  }
}
```

#### POST `/running`

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

#### DELETE `/running`

Cancels the currently running program.

**Response:**

- Status 204 No Content on success

### Program Storage Endpoints

The ControlUnit also provides program storage management at `/storage/*` endpoints.

#### GET `/storage/status`

Gets the health status of the storage subsystem. Follows the standard status endpoint format (see Common Response Patterns).

**Response Format:**

```json
{
  "data": {
    "status": "healthy",
    "service": "storage",
    "details": {
      "accessible": true
    }
  }
}
```

If storage is not accessible:

```json
{
  "data": {
    "status": "degraded",
    "service": "storage",
    "details": {
      "accessible": false,
      "error": "storage directory not accessible: /path/to/storage"
    }
  }
}
```

**Details:**

- `accessible`: Boolean indicating if the storage directory is accessible
- `error`: Error message if storage is not accessible (only present when accessible is false)

#### GET `/storage/programs`

Lists all stored programs.

**Response Format:**

```json
{
  "data": [
    "Standard Drying",
    "Quick Drying",
    "Pine Program"
  ]
}
```

#### GET `/storage/programs/{name}`

Gets a specific stored program by name.

**Path Parameters:**

- `name`: The name of the program

**Response Format:**

```json
{
  "data": {
    "programName": "Standard Drying",
    "description": "Standard drying program for pine",
    "phases": [
      {
        "name": "Initial Heating",
        "duration": 3600,
        "targetTemperature": 45,
        "steps": []
      }
    ]
  }
}
```

#### POST `/storage/programs`

Creates a new stored program.

**Request Body:**

```json
{
  "programName": "New Program",
  "description": "Description of the program",
  "phases": []
}
```

**Response:**

- Status 201 Created on success
- Status 409 Conflict if program already exists

#### POST `/storage/programs/{name}`

Updates an existing stored program.

**Path Parameters:**

- `name`: The name of the program to update

**Request Body:**

```json
{
  "programName": "Updated Program",
  "description": "Updated description",
  "phases": []
}
```

**Response:**

- Status 200 OK on success
- Status 404 Not Found if program doesn't exist

#### DELETE `/storage/programs/{name}`

Deletes a stored program.

**Path Parameters:**

- `name`: The name of the program to delete

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

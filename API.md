# Halko API Documentation

This document provides detailed information about the REST APIs exposed by each component of the Halko wood drying kiln control system.

## 1. SensorUnit API

Base Path: `/sensors/api/v1`

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
- `oven`: The highest of the two oven temperatures or a single oven temperature if one sensor is unavailable
- `wood`: The current wood temperature
- `ovenPrimary`: The primary oven temperature sensor reading
- `ovenSecondary`: The secondary oven temperature sensor reading

### Status Endpoints

#### GET `/status`

Checks the connection status of the sensor unit.

**Response Format:**
```json
{
  "data": {
    "status": "connected"
  }
}
```

Possible status values:
- `connected`: The sensor unit is connected and responding
- `disconnected`: The sensor unit is not connected or not responding

#### POST `/status`

Updates the status text displayed on the sensor unit's LCD.

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

Base Path: `/powers/api/v1`

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
- `power`: The name of the power channel (e.g., `heater`, `fan`, `humidifier`)

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
- `power`: The name of the power channel (e.g., `heater`, `fan`, `humidifier`)

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

## 3. Executor API

Base Path: `/engine/api/v1`

### Program Storage Endpoints

#### GET `/programs`

Lists all available programs (definitions loaded by executor).

**Response Format:**
```json
{
  "data": [
    {
      "name": "Standard Drying",
      "description": "Standard drying program for pine",
      "phases": [
        // Phase details...
      ]
    },
    {
      "name": "Quick Drying",
      "description": "Quick drying program for thinner woods",
      "phases": [
        // Phase details...
      ]
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
      },
      // More phases...
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
    "phases": [
      // Phase details...
    ]
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

## 4. Simulator API

The simulator mimics endpoints from the SensorUnit and Shelly devices.

### Simulated SensorUnit API

Base Path: `/sensors/api/v1`

#### GET `/temperatures`

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

#### GET `/status`

Gets the simulated connection status.

**Response Format:**
```json
{
  "data": {
    "status": "connected"
  }
}
```

#### POST `/status`

Logs a status message (simulates updating an LCD).

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

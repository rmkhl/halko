# Storage Service

The storage service is an independent microservice that handles program storage operations for the Halko system.

## Overview

This service provides REST API endpoints for managing stored programs. It was extracted from the executor service to create a more modular architecture.

## Endpoints

The service provides the following endpoints:

- `GET /storage/programs` - List all stored programs
- `GET /storage/programs/{name}` - Get a specific program by name
- `POST /storage/programs` - Create a new program
- `POST /storage/programs/{name}` - Update an existing program
- `DELETE /storage/programs/{name}` - Delete a program

## Configuration

The service can be configured through the main `halko.cfg` configuration file using the `storage` section:

```json
{
  "storage": {
    "base_path": "/path/to/storage",
    "port": 8091
  }
}
```

If no storage configuration is provided, it falls back to:
- Base path: the executor's base path, or `/tmp/halko` as a last resort
- Port: 8091

## Running

### Standalone
```bash
go run main.go
```

### With Docker Compose
The service is included in the main docker-compose.yml file and will start automatically when running the full Halko system.

## Dependencies

The service depends on:
- `github.com/rmkhl/halko/types` - Shared types and configuration
- Standard Go libraries for HTTP handling and file operations

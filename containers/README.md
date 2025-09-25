# Container Configuration

This directory contains all Docker-related files for the Halko project.

## Files

- `docker-compose.yml` - Main Docker Compose configuration file
- `Dockerfile.executor` - Dockerfile for the executor service
- `Dockerfile.powerunit` - Dockerfile for the powerunit service
- `Dockerfile.simulator` - Dockerfile for the simulator service
- `Dockerfile.storage` - Dockerfile for the storage service

## Usage

To run the services with Docker Compose from the project root:

```bash
cd containers
docker-compose up -d
```

Or from the project root:

```bash
docker-compose -f containers/docker-compose.yml up -d
```

## Build Contexts

The Docker Compose file uses build contexts pointing to the respective service directories in the parent directory, while the Dockerfiles are centralized here for better organization.

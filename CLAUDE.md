# Halko — Claude Code Guide

Halko is a distributed system for controlling and monitoring wood drying kilns. It runs bare-metal on Raspberry Pi with multiple Go microservices, a React frontend, and ESP32 firmware for temperature sensing.

## Tech Stack

- **Backend**: Go 1.26+ workspace (`go.work`) with 8 modules
- **Frontend**: React 18 + TypeScript + Material-UI + Redux Toolkit, bundled by Parcel
- **Firmware**: Arduino/ESP32 (C/Arduino), compiled with Arduino CLI
- **Deployment target**: Bare-metal Linux / Raspberry Pi

## Go Module Structure

```
types/        — shared types, imported by all other modules
controlunit/  — kiln control logic and program execution
powerunit/    — Shelly smart switch control
sensorunit/   — ESP32/Arduino serial bridge
simulator/    — hardware emulator (physics engines: simple/differential/thermodynamic)
dbusunit/     — systemd D-Bus integration
halkoctl/     — CLI management tool
tests/        — integration test suite (separate module)
```

When modifying types shared across services, change `types/` and then update all consumers.

## Build & Test

Always use the Makefile, not `go` commands directly from the root.

```bash
make all              # build all Go binaries → bin/
make test             # run all test suites
make lint             # golangci-lint + markdown + ESLint
make fmt-changed      # gofmt/goimports on changed files only
make go-tidy          # go mod tidy on all modules

# Webapp
make webapp-build     # production build
make webapp-dev       # dev server on :1234 with hot reload

# Development environment
make tmux-debug-run   # all services + simulator in tmux
LOGLEVEL=4 make tmux-debug-run  # verbose logging (0=ERROR … 4=TRACE)
SIMULATOR=thermodynamic make tmux-debug-run

# ESP32
make build-esp32      # compile firmware
make upload-esp32     # flash to device
make monitor-esp32    # serial monitor
```

Running tests directly: `cd tests && go test ./...`

Running linter on a single module: `cd <module> && golangci-lint run`

## Linting Rules

Config: `.golangci.yaml`. Enabled linters include `bodyclose`, `errchkjson`, `gocritic`, `nestif`, `nilerr`, `prealloc`, `revive`, `whitespace`, and others. The `receiver-naming` revive rule is disabled. Always run `make lint` before finishing a task.

## Architecture Conventions

- Each service exposes a REST+JSON HTTP API; endpoints are configured in `halko.cfg`
- Services share types via the `github.com/rmkhl/halko/types` module
- File-based storage under `/var/opt/halko/` (running programs, history)
- Configuration loaded from `/etc/opt/halko.cfg` at startup
- CORS headers are enabled on all services
- No mocks — tests use real logic (see `tests/`)
- Serial communication to ESP32 at 9600 baud

## Frontend Conventions

- Uses **Parcel** (not Webpack/Vite) — HMR works out of the box
- Redux Toolkit slices in `webapp/src/store/`
- RTK Query for all API calls (no raw fetch)
- MUI v5 with custom theme in `webapp/src/material-ui/`
- i18next for all user-visible strings
- API base URL configured in `webapp/src/store/` (dev vs. prod differ)

## Development Configuration

During development, the services read config from `halko.cfg` in the **workspace root** (not `/etc/opt/halko.cfg`). File storage (running programs, history) is written to `fsdb/` in the **workspace root** (not `/var/opt/halko/`). Both are git-ignored and must be created manually if missing.

## Deployment

```bash
OPTIMIZED=yes make all     # smaller binaries for Raspberry Pi (-s -w -trimpath)
make install               # copy binaries to /opt/halko
make systemd-units         # install + enable systemd services
make install-webapp        # build frontend + install to /var/www/halko
```

## What Not to Commit

`.gitignore` excludes: binaries (`bin/`), `.nodejs/`, `.arduino-*`, `fsdb`, `halko.cfg`, `node_modules/`.

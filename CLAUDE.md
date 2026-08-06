# Halko — Claude Code Guide

Halko is a distributed system for controlling and monitoring wood drying kilns. It runs bare-metal on Raspberry Pi with multiple Go microservices, a React frontend, and ESP32 firmware for temperature sensing.

## Tech Stack

- **Backend**: Go 1.26+ workspace (`go.work`) with 8 modules
- **Frontend**: React 18 + TypeScript + Material-UI + Redux Toolkit, bundled by Parcel
- **Firmware**: Arduino/ESP32 (C/Arduino), compiled with Arduino CLI
- **Deployment target**: Bare-metal Linux / Raspberry Pi

## Go Module Structure

```text
types/        — shared types, imported by all other modules
types/log/    — logging (nested module, imported by types and every service)
controlunit/  — kiln control logic and program execution
powerunit/    — Shelly smart switch control
sensorunit/   — ESP32/Arduino serial bridge
simulator/    — hardware emulator (physics engines: simple/differential/thermodynamic)
dbusunit/     — systemd D-Bus integration
halkoctl/     — CLI management tool
```

When modifying types shared across services, change `types/` and then update all consumers.

The module list lives in the `Makefile` as two variables: `MODULES` (the
binary-producing services, used by `all`/`build`/`install`/`systemd-units`) and
`GO_MODULES` (all of them, adding `types` and `types/log`, used by
`fmt-changed`/`go-tidy`/`prepare`, and the list the per-module `test-<module>`
and `lint-<module>` targets are generated from). Adding a module means updating
`GO_MODULES` and `go.work` — `make prepare` regenerates `go.work` from
`GO_MODULES`, so a module missing there silently drops out of the workspace.

## Build & Test

Always use the Makefile, not `go` commands directly from the root.

```bash
make all              # build all Go binaries → bin/
make test             # go test ./... in every Go module
make test-controlunit # one module only — one target per entry in GO_MODULES
make test-race        # all modules, with the race detector
make lint             # golangci-lint + markdown + ESLint
make fmt-changed      # gofmt/goimports on changed files only
make go-tidy          # go mod tidy on all modules

# Webapp
make build-webapp     # production build
make run-webapp       # dev server on :1234 with hot reload

# Development environment
make tmux-debug-run   # all services + simulator in tmux
LOGLEVEL=4 make tmux-debug-run  # verbose logging (0=ERROR … 4=TRACE)
SIMULATOR=thermodynamic make tmux-debug-run
make tmux-debug-fail-run  # same, but the simulator injects escalating sensor failures
LOG_DIR=/tmp/halko-logs make tmux-debug-run  # write the log files elsewhere
```

Every tmux window pipes its output through `tee`, so the logs are on the
console and in `logs/<window>.log` under the workspace root (git-ignored,
overwritten on every start). Override the directory with `LOG_DIR`; the start
script aborts if it cannot create or write to it.

**When debugging a run, read `logs/` first — no need to ask.** It always holds
the *latest* run: `logs/controlunit.log`, `logs/simulator.log`,
`logs/sensorunit.log`, `logs/powerunit.log`, `logs/dbusunit.log`,
`logs/webapp.log`. Together with `fsdb/running/` and `fsdb/history/` (the run
that just ended) they are the primary evidence for any "what happened during
that run" question.

The simulator emulates the ESP32 over a pseudo-terminal, so the real
`sensorunit` service runs against it. It creates that device at the path named
by `sensorunit.serial_device` in `halko.cfg`, which during development must
name a path the simulator may create (for example `/tmp/halko-esp32`) rather
than real hardware. Leaving it pointing at a real device path such as
`/dev/ttyUSB1` makes the simulator refuse to start, naming that path in the
error.

```bash
# ESP32
make build-esp32      # compile firmware
make upload-esp32     # flash to device
make monitor-esp32    # serial monitor
```

Running tests for a single module: `make test-<module>` (or `cd <module> && go
test ./...`). Adding a module to `GO_MODULES` generates its `test-` target
automatically.

Running the linter on a single module: `make lint-<module>` (or `cd <module> &&
golangci-lint run`). Like the test targets, these are generated from
`GO_MODULES`.

The `lint-*` targets are deliberately non-gating: `make lint` runs every linter
and reports all of their results in one pass rather than stopping at the first
one with issues. The `test-*` targets behave the same way. Do not add gating to
either.

CI needs the opposite, so `test-ci`, `lint-ci` and `lint-markdown-ci` exist as
gating counterparts. They run the same commands over the same `GO_MODULES`
list and still report every module, but exit non-zero at the end if any of them
failed. GitHub Actions calls these, never `make test` or `make lint` — those
always exit 0 and would make a green build meaningless.

## Test Layout

Tests live in the same package as the code they cover, the standard Go layout —
there is no separate test module. A change to `controlunit/engine` is tested by
`controlunit/engine/*_test.go`, `types.LoadConfig` by `types/config_test.go`,
and so on. Put a new test next to its subject; `make test` picks it up with no
wiring.

Tests that drive a service end to end belong with that service. `simulator/
shelly_api_test.go` is the model: it builds the simulator binary, runs it as a
real process and drives its HTTP API. Such a test must stay self-contained —
build what it needs, write its config to `t.TempDir()`, and never depend on a
service the developer is expected to have running.

## Linting Rules

Config: `.golangci.yaml`. Enabled linters include `bodyclose`, `errchkjson`, `gocritic`, `nestif`, `nilerr`, `prealloc`, `revive`, `whitespace`, and others. The `receiver-naming` revive rule is disabled. Always run `make lint` before finishing a task.

## Architecture Conventions

- Each service exposes a REST+JSON HTTP API; endpoints are configured in `halko.cfg`
- Services share types via the `github.com/rmkhl/halko/types` module
- File-based storage under `/var/opt/halko/` (running programs, history)
- Configuration loaded from `/etc/opt/halko.cfg` at startup
- CORS headers are enabled on all services
- No mocks — tests use real logic
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

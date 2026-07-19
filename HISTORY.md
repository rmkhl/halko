# Halko — Project History and Credits

Halko is a distributed system for controlling and monitoring wood drying
kilns, built to run on a Raspberry Pi with ESP32-based temperature
sensing. This file summarizes how the project got to release
1.0 and credits the people who made it happen.

## History

### 2024 — Foundations

- **March–April 2024** — Project started: initial program/phase schemas, a
  Go service skeleton with a filesystem database, the first version of the
  React webapp, and a hardware simulator for development without a kiln.
- **May–June 2024** — First program executor: program configuration APIs,
  program execution with step changes, execution CSV logging, and the
  programs UI with sortable steps. Initial sensor unit (Arduino-based
  serial bridge) added.
- **October–November 2024** — Shared `types` module introduced across
  services. Program control switched to PID-based power control, and the
  executor was rebuilt around a finite state machine (heating, acclimate,
  cooling).
- **December 2024** — First version of the power unit for controlling
  Shelly smart switches.

### 2025 — From prototype to system

- **February–April 2025** — Power unit hardened: duty-cycle power
  controller, error handling, and startup fixes. The simulator learned to
  emulate Shelly switches directly.
- **June 2025** — Major consolidation: Makefile-driven builds, systemd
  service templates and daemonization, graceful shutdown across services,
  program storage moved into the executor, the standalone configurator
  service retired, and the sensor unit reworked. First integration tests.
- **July–August 2025** — Heartbeat reporting, configuration defaults with
  human-readable durations, unified power-control settings, program
  validation tests, and the `halkoctl` command-line tool.
- **September–October 2025** — API simplification: migration from gin to
  `net/http`, all endpoints defined in configuration, leveled logging
  across services, expanded `halkoctl` (programs, status, temperatures,
  display), and resilient serial handling for sensor unplug/replug.
- **December 2025** — The executor became the **control unit** with
  storage rolled in. The webapp was rewritten: new program editor, status
  page, run log endpoint with live WebSocket streaming, and per-run
  history.

### 2026 — Road to 1.0

- **February 2026** — Live execution charts in the webapp, CORS support,
  and a pluggable simulator physics layer with simple, differential, and
  thermodynamic engines.
- **March 2026** — Development environment moved from Docker to a
  tmux-based workflow, Raspberry Pi setup guide and optimized builds,
  system status page, `dbusunit` for systemd D-Bus integration (VPN and
  power control), and ESP32 firmware build/upload tooling.
- **May 2026 — Switch from Arduino to ESP32** — The sensor unit firmware
  was rewritten for the ESP32, replacing the original Arduino Nano
  bridge: more capable hardware, OLED display support, and verified
  serial, display, and sensor code.
- **April–June 2026** — Terminology unified from "oven" to "kiln", delta
  power control fixes, repo-wide lint cleanup, and deployment hardening
  (startup ordering, restart behavior, serial robustness).
- **July 2026** — Final polish for 1.0: TypeScript strictness and webapp
  crash fixes, configurable D-Bus socket, refreshed integration tests,
  temperature-focused execution charts, and CSV and PDF export of run
  reports.

## Credits

### Initiator

- **Ruokangas Guitars** — for initiating the project and defining the
  real-world need: precise, repeatable drying of tonewood.

### Hardware

- **Santeri Airasmaa** — for building and testing the controller
  hardware.

### Development

- **Marko Teiste** — architecture and development across all services,
  firmware, and tooling.
- **Petri Kallio** — UX design and initial prototype implementation.

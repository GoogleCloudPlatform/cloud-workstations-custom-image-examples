<!--
Copyright 2026 Google LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
-->

# Developer Guide

Welcome to the Cloud Workstations Custom Image blueprint! This guide provides an overview of the development practices, testing frameworks, and architectural guidelines required to contribute to this project.

## Core Architectural Concepts

This project builds a complex, multi-service environment utilizing systemd, Docker-in-Docker, Wayland, and remote desktop protocols within a single container. Understanding the lifecycle is critical:
* **Entrypoint**: `entrypoint.sh` runs first, generating persistent identifiers.
* **Systemd Handoff**: The entrypoint hands over to `/sbin/init` (systemd) which orchestrates the rest of the startup process.
* **Build-time Hooks**: The `Dockerfile` executes scripts in `assets/build-hooks.d/` at the end of the build. This is the preferred way to add custom build logic.
* **Dynamic Rendering**: `config-rendering.service` executes early in the boot to read runtime `ENV` variables and render configuration templates (Guacamole, Nginx, Frontend JS).

## Architecture Deep-Dives

### Traffic Readiness Handover (Firewall-based)
1. **`entrypoint.sh` Hard Silence**: Executes `traffic_control.sh block` to drop all external traffic via `nftables`.
2. **`permit-traffic.service`**: Flushes firewall rules only after Nginx returns a successful response.

## Developer Tools & Frameworks

*   **[Bats-core](https://github.com/bats-core/bats-core)**: For Bash unit and integration testing.

### Management Scripts

| Script | Purpose |
| :--- | :--- |
| `run_all_tests.sh` | Orchestrates all local and (optionally) integration tests. |

## Git Pre-commit Hooks
Automatically runs unit tests, linters, and secret checks before commits using the `pre-commit` framework.
```bash
pre-commit install
```

## AI Agents Instructions
Agent mandates and constraints: 👉 **[AGENTS.md](../AGENTS.md)**

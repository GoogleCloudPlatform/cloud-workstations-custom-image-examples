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

---
name: manage-cws-lifecycle
description: Handles Cloud Workstation lifecycle (start, stop, restart), SSH connectivity, and monitoring long-running operations (LROs) or Cloud Builds. Use when the user asks to manage the workstation state or needs to check build progress.
---

# Manage Cloud Workstations Lifecycle

This skill provides the standardized workflow for interacting with the active Cloud Workstation instance using the Google Cloud CLI (`gcloud`) and the provided management scripts.

## Connecting via SSH
To run commands directly inside the user's workspace, use the `gcloud workstations ssh` command.

## Standardized Management Scripts
**MANDATE:** Always use these scripts instead of raw `gcloud` commands for long-running operations.

*   **Monitor Cloud Builds**: Use `./scripts/monitor_build.sh [BUILD_ID] [--latest]` to poll until a build finishes.
*   **Manage Workstation Lifecycle**: Use `./scripts/manage_workstation.sh [start|stop|restart|wait|tunnel] [WORKSTATION_NAME]` to handle state transitions, monitoring, or establishing secure tunnels.

## SSH Connectivity and Tunneling
The `tunnel` action automates the process of starting the workstation, opening the workspace URL in your browser, and establishing a secure SSH tunnel.

```bash
# Start workstation and establish a tunnel (default local port 2222)
./scripts/manage_workstation.sh tunnel $WORKSTATION

# Establish a tunnel on a custom local port and skip opening the browser
./scripts/manage_workstation.sh tunnel $WORKSTATION --local-port 8888 --browser ""
```
This action also ensures a `ws` host entry exists in your `~/.ssh/config`, allowing you to connect simply by running `ssh ws`.

## Hot-Patching
**MANDATE:** After troubleshooting and resolving an issue, apply the changes in the local codebase and then hot-patch the live instance whenever possible to avoid waiting for a full rebuild and deployment. Use `gcloud workstations ssh` to transfer built assets or modify configuration files directly on the running instance.

## Monitoring Long Running Operations (LROs)

### Workstation State Monitoring
Use the provided script to poll the workstation's state:
```bash
./scripts/manage_workstation.sh wait $WORKSTATION --target-state STATE_RUNNING
```

### Cloud Build State Monitoring
Monitor Cloud Builds using `gcloud builds describe` and filter for terminal statuses: `SUCCESS`, `FAILURE`, `CANCELLED`, or `TIMEOUT`.

When using `./scripts/monitor_build.sh`, you can specify discovery flags:
```bash
# Monitor the absolute latest build in the current project/region
./scripts/monitor_build.sh --latest

# Monitor a build with a custom lookback window
./scripts/monitor_build.sh --lookback 5
```

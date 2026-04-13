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

# Preflight Layer Architecture

The Preflight layer is responsible for managing the user's initial connection experience while background services are still warming up. It provides a highly optimized loading interface and orchestrates the transition from "Booting" to "Ready."

## Layer Responsibility

This layer sits between the **Base OS** and the **Graphical Desktop**. Its primary job is to hide "Connection Refused" errors from the user and provide actionable feedback during the boot process.

| Feature | Responsibility | Technology |
| :--- | :--- | :--- |
| **Frontend UI** | Visual status and precision timer | Vite + Vanilla TypeScript |
| **Readiness Probing** | Automated health checking of RDP/SSH ports | XHR / Fetch polling |
| **Traffic Control** | Firewall management during early boot | Nginx + `iptables` |
| **Interception** | Catching early traffic before the desktop is ready | Nginx `error_page` mapping |

## The "Hard Silence" Boot Sequence

Preflight implements a "Silent Retry" strategy to ensure a seamless transition:

1.  **Block Stage**: At container entry, all external traffic is dropped at the kernel level.
2.  **Internal Readiness**: Nginx starts and serves the Preflight UI on the loopback interface.
3.  **Permit Stage**: Once Nginx is locally reachable, the firewall is flushed. The user's browser now sees the loading screen instead of a timeout.
4.  **Backend Probing**: The UI (running in the user's browser) polls the workstation's `/healthz` endpoint.
5.  **Handover**: Once the desktop RDP port responds, the UI triggers an automatic redirect to the workstation's main interface.

## Frontend State Model (SPA)

The UI operates as a stateful Single Page Application, ensuring high performance and a seamless user experience:

*   **Atomic State Singleton**: The application state is managed in a central `state` object (`types.ts`), ensuring a single source of truth.
*   **Transient UI Staging**: To prevent data loss during navigation (e.g., from Settings to Language), the UI utilizes a `uiTransient` staging object. Unsaved changes are "snapshotted" during modal transitions and re-applied upon return.
*   **Zero-Reload Configuration**: Savings settings updates the in-memory state and the browser history (`replaceState`) without triggering a page reload.

## Phased Health Monitoring

The health check cycle (`health_module.ts`) operates in two distinct phases:

1.  **Nominal Phase**: Polling occurs at a high frequency (customizable via the Retry Interval slider) to ensure rapid connection once ready.
2.  **Timeout Phase**: If the `timeoutMs` threshold is exceeded, the system transitions to an **Exponential Backoff** strategy (current * 1.5, max 30s) to minimize backend load during extended startup periods.

## Frontend Module Structure

The UI is built as a set of decoupled TypeScript modules:
*   `health_module.ts`: Phased polling logic and exponential backoff.
*   `config_module.ts`: Management of settings, logarithmic mappings, and transient snapshots.
*   `i18n_module.ts`: Hash-validated translation loading and DOM application.
*   `ui_module.ts`: Modal stack management, adaptive layout logic, and event orchestration.
*   `sbom_module.ts`: Dynamic license and component metadata rendering.
*   `window_utils.ts`: Browser location and navigation abstraction layer to support testability under strict JSDOM security constraints.

## Internal Documentation Architecture
The Preflight dashboard utilizes a pre-rendered documentation pipeline to ensure a high-performance, self-contained experience. During the build phase, Markdown sources (e.g., `examples/preflight/docs/privacy_notice.md`) are converted into HTML fragments. This allows the Single Page Application (SPA) to fetch and render professional documentation internally without external dependencies or heavy client-side Markdown libraries.

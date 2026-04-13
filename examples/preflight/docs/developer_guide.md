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

# Preflight Layer Developer Guide

This guide explains how to contribute to and modify the **Custom Image Preflight** layer.

## Development Environment

The Preflight UI is built using **Vite** and **Vanilla TypeScript**.

*   **Source Code**: `web/`
*   **Build Output**: `web/dist/`
*   **Documentation Source**: `docs/`
*   **Nginx Templates**: `assets/etc/nginx/`

### Local UI Development

1.  Navigate to the `web/` directory.
2.  Install dependencies: `npm install`.
3.  Start the development server: `npm run dev`.
4.  Run tests: `npm test`.

## Build Pipeline

The Preflight layer is built as a standalone Docker image. The build process includes several specialized automation steps:

1.  **Document Rendering**: The `scripts/build/generate_fragments.js` utility uses the **marked** library to convert Markdown files in `examples/preflight/docs/` into professional HTML fragments (with Tailwind styling) for the internal Help system.
2.  **Security (SRI)**: The `inject-sri.cjs` script automatically generates and injects **Subresource Integrity** hashes for all JavaScript and localization assets, ensuring technical session integrity.
3.  **Artifact Generation**: Vite compiles the TypeScript assets and prepares the final production distribution.
4.  **Asset Export**: Compiled assets are stored in `/var/www/html/` within the Docker image.

## Hot-patching

To rapidly test UI changes on a running workstation:

1.  Use the `./examples/preflight/skills/update-preflight/scripts/hotpatch_frontend.sh` script from the project root.
2.  Use the `--wipe` flag to ensure a clean-slate deployment on the live instance:
    ```bash
    ./examples/preflight/skills/update-preflight/scripts/hotpatch_frontend.sh --wipe
    ```
3.  This script builds the local source, clears the target directory, and transfers the `dist/` assets via SSH.

## Customization

### Adding a Language
1.  Add the translation JSON file to `web/public/locales/[code].json`.
2.  Register the language in `web/i18n_module.ts` under `SUPPORTED_LANGS`.
3.  The SRI script will automatically detect and hash the new locale file during the next build.

### Modifying Health Checks
The health-check logic is centralized in `web/health_module.ts`. It utilizes a **Phased Monitoring** approach:
*   **Nominal Stage**: Polling at the user-defined Retry Interval.
*   **Backoff Stage**: Transitioning to exponential backoff (current * 1.5, max 30s) after the timeout threshold is reached.

### URL Configuration Overrides
The following parameters can be appended to the workstation URL to override default Preflight UI behavior for testing and development:

| Parameter | Default | Description |
| :--- | :--- | :--- |
| `autoRedirect` | `true` | Automatically navigate to the desktop when ready. |
| `simulateDelay`| `0`    | Hold the UI in "STARTING" state for X seconds. |
| `timeout`      | `250`  | Nominal polling phase (seconds) before backoff. |
| `debug`        | `false` | Force open the Technical Debug overlay. |
| `lang`         | `auto` | Override the interface language (e.g., `?lang=ar`). |

### i18n, Localization, & WCAG Accessibility
*all* user-facing text must be localized.
1. **`data-i18n` Attributes**: Standard for static HTML elements.
2. **Dynamic TypeScript Translation**: Use the `t('key', 'default')` helper for generated content.
3. **Resolution Hierarchy**: Language is resolved via (1) URL Parameter, (2) Local Storage, (3) Server Meta, (4) Browser Preference, (5) 'en' Fallback.

### SPA State Management
The Preflight dashboard follows a **Single Page Application** state model for its configuration:

1.  **URL Persistence**: All technical parameters (e.g., `?autoRedirect=true&debug=false`) are synchronized with the browser address bar via `history.replaceState` upon saving.
2.  **Session Re-simulation**: To support iterative testing of startup scenarios, clicking **Save** in the UI executes a total functional reset of the `AppState`, resetting the session timer and restarting the health check sequence without a page reload.
3.  **Reporting Consistency**: The **Copy URL** link captures the real-time, unsaved state of the UI elements for peer sharing. Conversely, the **Bug Report** strictly captures the persistent, saved `state.config` for support baseline stability.

### Technical Configuration Constraints
The frontend enforces strict architectural fallbacks and intuitive control mapping:
*   `DEFAULT_TIMEOUT_MS`: 200s nominal polling threshold.
*   **Timeout Mapping**: The UI uses a logarithmic curve to map the slider (0-100) to an exponential range of **1s to 500s**, providing high precision at lower values.
*   `DEFAULT_RETRY_INTERVAL_MS`: 1s architectural default polling frequency.
*   `MAX_BACKOFF_INTERVAL_MS`: 30s maximum interval after timeout exhaustion.

### Developer Tools & Frameworks
*   **[Jest](https://jestjs.io/)**: For testing TypeScript frontend components.
    *   **Routing Tests**: JSDOM v30+ strictly enforces `window.location` immutability. All navigation testing MUST mock and assert against the abstractions in `window_utils.ts`.
    *   **Event Dispatching**: To prevent JSDOM `TypeError` crashes during rapid teardowns, dispatch synthesized UI events (e.g., `KeyboardEvent`) on `document.body` with `{ bubbles: true }` instead of directly on the global `document` or `window` objects.

### Management Scripts
| Script | Purpose |
| :--- | :--- |
| `hotpatch_frontend.sh` | Builds Vite assets and syncs them to the live workstation. |

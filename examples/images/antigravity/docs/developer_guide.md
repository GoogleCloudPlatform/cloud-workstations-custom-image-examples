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

# Antigravity Developer Guide

This guide explains how to customize and extend the **Antigravity Custom Image for Cloud Workstations**. It is intended for platform engineers and developers building specialized workstation images.

## Architecture Recap

The Antigravity Custom Image for Cloud Workstations uses a **declarative asset injection** pattern. To modify the image, you primarily interact with the `assets/` directory rather than modifying the `Dockerfile` directly.

## Customization Patterns

### 1. Adding New Tools via APT
To add a new tool that is available via a public or private APT repository:

1.  **Add Repo**: Place a `.list` file in `assets/etc/apt/sources.list.d/`.
2.  **Add Hook**: Create a new script in `assets/build-hooks.d/` (e.g., `20_install_my_tool.sh`):
    ```bash
    #!/bin/bash
    apt-get update
    apt-get install -y my-tool
    ```
3.  **Permissions**: Ensure the script is executable (`chmod +x`).

### 2. Injecting Desktop Configurations
To pre-configure the GNOME environment for users:

1.  **GSettings**: Use a build hook to apply system-wide GSettings overrides in `/usr/share/glib-2.0/schemas/`.
2.  **Autostart**: Place `.desktop` files in `assets/etc/xdg/autostart/` to launch applications when the session begins.
3.  **Default Icons**: Add custom `.desktop` files to `assets/usr/share/applications/` to add them to the GNOME "Activities" view.

### 3. Modifying the Autostart Logic
Antigravity uses XDG Autostart. If you need to change how the main Antigravity dashboard launches:
*   Edit `assets/etc/xdg/autostart/antigravity.desktop`.
*   You can use the `Exec=` line to point to a wrapper script that performs environment checks before launching the UI.

## Build and Validation

### Local Testing
Since Antigravity is a layered image, it is best tested using **Skaffold**:

```bash
# From the examples/images/antigravity directory
skaffold build --file-output=build.json
```

### Validation Workflow
Follow the mandatory [Agent Validation Lifecycle](../../../../AGENTS.md#2-agent-validation-lifecycle):
1.  Run local tests: `skills/validate-image-updates/scripts/run_local_tests.sh`.
2.  Verify the image successfully installs the `antigravity` package.
3.  Confirm the `.desktop` file is correctly placed in `/etc/xdg/autostart/`.

## Best Practices

*   **Keep Hooks Small**: Each hook should do one thing (e.g., install one toolchain).
*   **Idempotency**: Ensure hooks can run multiple times without failure (though they usually run once during build).
*   **Layer Agnosticism**: Do not assume files from the GNOME layer are in a specific state; use the hook mechanism to ensure your requirements are met.

---
Brought to you by the Google Cloud Professional Services Organization Apps team (PSO AppΣ)

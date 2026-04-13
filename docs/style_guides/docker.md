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

# Dockerfile Style Guide

## Build-time (ARG) vs. Runtime (ENV) Configuration

To maximize the flexibility of the container image across different environments, strictly distinguish between build-time and runtime parameters:

1. **`ARG` (Build-Time):** Use `ARG` strictly for parameters that physically alter the image structure or the packages installed. For example:
   * Package versions (`GUACAMOLE_VERSION`).
   * Toggles for heavy dependencies that should not be downloaded if unused (`INSTALL_TIGERVNC`).
2. **`ENV` (Runtime):** Use `ENV` for variables that dictate how the image behaves when the container is executed.
   * If a service is optional but its binaries are installed, use an `ENV` to toggle its startup (e.g., `ENABLE_TIGERVNC`, `ENABLE_AUDIO_INPUT`).
   * Pass build `ARG` defaults into the runtime environment by explicitly mapping them: `ENV ENABLE_TIGERVNC=${INSTALL_TIGERVNC}`.

This ensures a single, immutable image can be promoted through multiple environments while drastically altering its behavior simply by passing `-e` flags to `docker run`.

## Image Layer Optimization
1. **Combine RUN statements**: Chain `apt-get update`, `apt-get install`, and `rm -rf /var/lib/apt/lists/*` in a single `RUN` block to minimize layer size.
2. **Multi-stage Builds**: Use `FROM alpine AS fetcher` blocks for heavy network downloads (like `crane` pulls or `curl` downloads) so that build utilities and intermediate archives don't bloat the final production image.
3. **Build Hook Pattern**: To maintain a clean and modular `Dockerfile`, move complex tool installations or configurations into external scripts in an `assets/build-hooks.d/` directory.
   * **Mandate**: Every hook script must be executable and follow the [Bash Style Guide](bash.md).
   * **Implementation**: COPY the hooks into a temporary directory in the image, iterate through and execute them, then delete the directory in the same `RUN` layer.
   * **Example**:
     ```dockerfile
     COPY assets/build-hooks.d/ /build-hooks.d/
     RUN if [ -d "/build-hooks.d/" ]; then \
           for hook in /build-hooks.d/*.sh; do \
             [ -e "$hook" ] || continue; \
             /bin/bash "$hook"; \
           done; \
           rm -rf /build-hooks.d/; \
         fi
     ```

## Formatting and Syntax
1. **Empty Line After FROM**: Do not place an empty line immediately after a `FROM` instruction. Group the `FROM` instruction with its initial stage configuration (like `ARG`, `WORKDIR`, `USER`, or `ENV`). Use empty lines to separate distinct logical blocks further down in the stage.
2. **Sort Multi-Line Arguments**: Sort multi-line lists (like `apt-get install` packages or `apk add` dependencies) alphabetically. This improves readability, reduces duplicate entries, and simplifies git diffs.
3. **Consistent Formatting**: Use a backslash (`\`) at the end of a line followed by a newline and consistent indentation (strictly **2 spaces**) for multi-line commands to maintain readability.
4. **Avoid `latest`**: Prefer pinning base images and packages to specific versions or use `ARG`s to ensure reproducible builds and clear upgrade paths.
5. **Stage Header Blocks**: Every `FROM` instruction must be preceded by a distinct comment block clearly labeling the stage. The block must consist of a top line of exactly 63 hash characters (`#`), a middle line containing `# Stage: <name>` (or a short description), and a bottom line of 63 hash characters.
   ```dockerfile
   ###############################################################
   # Stage: <stage-name-or-description>
   ###############################################################
   FROM <image>
   ```
6. **Syntax 1.4 and Heredocs**: Dockerfile syntax version 1.4 (`# syntax=docker/dockerfile:1.4`) supports Heredocs, which can greatly improve the readability of complex multi-line `RUN` scripts or inline file generation. **However**, due to a current limitation with the Google Cloud Build `skaffold` container image's included Docker version, Heredocs cannot be used yet in this repository. In any other unconstrained environment, Heredocs should be the preferred standard.

## Testing Docker Containers

We do not test the `Dockerfile` syntax itself; we test the runtime integrity of the final built image.

### How to Write Tests
1. **Service Viability**: Use integration tests to verify the container image successfully boots systemd and all expected background daemons (`dockerd`, `nginx`, `guacd`) transition to the `active` state.
2. **Volume Mounts and Permissions**: Verify that runtime configurations (like dynamically generated `ephemeral.env` files) have the precise permissions and ownership required by the container's unprivileged users.
3. **Network Readiness**: Test that the container successfully binds to the expected host ports (e.g., 80, 22, 3389).

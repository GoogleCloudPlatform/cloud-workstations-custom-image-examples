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
name: troubleshoot-cws
description: Executes diagnostic procedures when integration tests fail, systemd cycles occur, or Wayland/RDP crashes are reported. Provides access to the official troubleshooting playbook.
---

# Troubleshoot Cloud Workstations

This skill provides the standardized workflow for diagnosing and resolving issues within the custom GNOME-based Cloud Workstations image.

## Basic Troubleshooting & Validation

**MANDATE: When an error is discovered and resolved, you MUST cover the case with additional tests (unit or integration) to prevent the issue from recurring.**

If the user reports an issue, or if you need to validate a fix, you must follow the official troubleshooting process defined in the playbook.

### The Playbook
See `../../docs/playbooks/troubleshooting.md` for the official Cloud Workstations (CWS) Troubleshooting Playbook. The playbook covers:
*   Automated Initial Assessment using `run_all_tests.sh`.
*   Targeted Diagnosis based on test failures (e.g., systemd cycles, GNOME shell crashes, Nginx/Guacamole issues, RDP NLA verification).
*   Known Issues & Architecture Gotchas.

**MANDATE:** Do not attempt manual, ad-hoc debugging of the live instance until you have executed the test suites as described in the playbook.

## Advanced Troubleshooting & Patching

### The D-Bus "Golden Command"
To interact with the user session (e.g., modifying `gsettings` or using `gnome-extensions`) via SSH, use the provided script to ensure the correct session bus address is used:
```bash
./scripts/dbus_session_cmd.sh "<command>"
```
*Note: Using `dbus-run-session` is discouraged for live patching as it creates a transient session that does not persist changes to the active UI.*

### Non-Clobbering Extension Patches
When unzipping GNOME extensions during the build process, the target directory is overwritten. To ensure custom patches (CSS/JS) are preserved:
1.  **Storage**: Place patch files in `assets/usr/share/gnome-shell/extension-patches/<extension-id>/`.
2.  **Application**: Apply the patch in the build hook (`assets/build-hooks.d/01_install_gnome_extensions.sh`) *after* the `unzip` command has completed.
3.  **Persistence**: Never place patches directly in the extension's functional directory in the `assets/` folder, as they will be clobbered by the download.

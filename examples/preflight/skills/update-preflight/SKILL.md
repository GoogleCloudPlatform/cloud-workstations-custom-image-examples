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
name: update-preflight
description: Manages preflight web app UX/UI constraints, layout stability, and the frontend hot-patching deployment process. Use when modifying HTML, CSS, or JS in the examples/preflight/web/ directory.
---

# Update Cloud Workstations Frontend

This skill manages the updates to the preflight web UI and ensures adherence to the repository's UX/UI standards.

## UX/UI Standards
When modifying the frontend, you MUST strictly adhere to the authoritative UX/UI standards:
👉 **[Authoritative UX/UI Standards](../../docs/ux_standards.md)**

## Frontend Hot-patching Workflow

When modifying the preflight page or other web assets in the `examples/preflight/web/` directory, use the following hot-patch script to provide immediate feedback on the live instance:

1.  **Execute**: Run `./scripts/test_and_hotpatch.sh` from the skill directory.
2.  **What it does**: This script automatically runs the frontend test suite via Jest (`npm test`). If successful, it builds the frontend assets via Vite (`npm run build`) and syncs the `examples/preflight/web/dist/` directory to `/var/www/html/` on your active workstation.
3.  **Validate**: Refresh the browser and verify changes immediately on the live instance before concluding the task.

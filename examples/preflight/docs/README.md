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

# Preflight Layer Documentation

This directory contains technical documentation for the **Custom Image Preflight** layer.

## Documentation Standards

To ensure consistency and compatibility across the project, adhere to the following standards:

### File Naming
*   **Convention**: Use `snake_case.md` for all technical documentation.
*   **Rationale**: Ensures maximum compatibility with Linux environments and consistent sorting.

### Structure & Redundancy
*   **Agnosticism**: Focus exclusively on the loading experience, health-check orchestration, and frontend implementation.
*   **Linking**: Refer to the **GNOME Layer** for details on RDP/Desktop handoff.


### README Strategy
*   **Top-Level `README.md`**: Reserved for external consumption. Must contain a high-level overview, configuration variables (ARGs/ENVs), quick-start guides, and a link pointing developers to the `docs/` folder.
*   **Internal `docs/README.md`**: Reserved for internal maintenance. Acts as the table of contents for technical deep-dives (`architecture.md`, `developer_guide.md`, etc.) and enforces layer-specific documentation rules.

## Documentation Map

| File | Purpose |
| :--- | :--- |
| [architecture.md](architecture.md) | Frontend architecture, health-check logic, and readiness probing. |
| [developer_guide.md](developer_guide.md) | How to build, test, and hot-patch the preflight UI. |

---
Brought to you by the Google Cloud Professional Services Organization Apps team (PSO AppΣ)

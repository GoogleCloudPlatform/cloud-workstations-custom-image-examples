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

# Google Cloud Workstations - GNOME Blueprint

This repository defines a highly customized, multi-service Docker image for Google Cloud Workstations (CWS), featuring a professional GNOME desktop environment and a cinematic "Preflight" loading experience.

## 🚀 Quick Start by Role

Select your role to find the most relevant documentation and workflows:

*   **[Developers](docs/developer_guide.md)**: Environment setup, local testing, and hot-patching.
*   **[Architects](docs/architecture.md)**: Deep-dives into systemd, networking, and security.
*   **[Operators / SREs](../../../docs/playbooks/troubleshooting.md)**: Monitoring, logging, and troubleshooting.
*   **[Security & Compliance](../../../docs/style_guides/docker.md)**: Hardening, ephemeral credentials, and SBOM.
*   **[Hackathon Participants](../../../docs/playbooks/hackathon_guide.md)**: Rapid prototyping and blueprint extension.

## 🏗️ Architectural Modules

This blueprint is organized into self-contained modules to ensure high maintainability and clarity:

### 1. GNOME Desktop Layer (`examples/images/gnome/`)
The core environment providing the desktop experience, systemd orchestration, and terminal tools.

### 2. Preflight Dashboard (`examples/preflight/`)
A cinematic loading interface that intercepts early traffic and provides technical telemetry.
*   👉 **[UX Standards](../../preflight/docs/ux_standards.md)**
*   👉 **[Language Priorities](../../preflight/docs/language_priorities.md)**
*   👉 **[Developer Guide](../../preflight/docs/developer_guide.md)**

## 🛠️ Global Tooling

*   **Gemini CLI**: Context-aware AI assistance directly in your terminal.
*   **Skaffold**: Automated container build and deployment.
*   **Bats & Jest**: Comprehensive Bash and TypeScript testing suites.

---
👉 **[Full Documentation Index](docs/README.md)** | 👉 **[AI Agents Instructions](AGENTS.md)**

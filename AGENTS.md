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

# AI Agents Instructions: Global Mandates

This repository defines a highly customized Google Cloud Workstation environment. Strict adherence to these global mandates is required for any AI agent.

## 1. Repository Management & Git Constraints

**CRITICAL MANDATE: Git is strictly READ-ONLY.**

As an AI agent, you must **NEVER** use `git` commands to manage the repository, branches, or commits, nor should you stage or push changes. Staging and committing are the exclusive responsibility of the human user.

You may use `git` exclusively for **read-only analysis** (`git status`, `git diff`, `git log`).

## 2. Agent Validation Lifecycle

Every technical change MUST follow this sequential workflow:

1.  **Stop Workstation**: Ensure the instance is in `STATE_STOPPED`.
2.  **Run Local Tests**: Execute `skills/validate-image-updates/scripts/run_local_tests.sh`.
3.  **Monitor Build**: Wait for the Cloud Build to reach `SUCCESS`.
4.  **Start Workstation**: Only after a successful build.
5.  **Run Integration Tests**: Execute integration suites on the live instance.
6.  **Persona Reviews**: Conduct in-depth persona reviews (UX, SEC, SRE, etc.) after validation.

## 3. Global Agent Skills (Foundation)

These global skills are available within the `skills/` directory:

*   **`manage-cws-lifecycle`**: interacting with the live CWS instance (start/stop).
*   **`troubleshoot-cws`**: Official diagnostic playbook for service crashes.
*   **`validate-image-updates`**: Enforces the mandatory 6-step validation workflow.
*   **`enforce-standards`**: Mandatory style guides for Bash, TypeScript, and Docker.
*   **`persona-swe`**: Software Engineering mandates and history refactoring.
*   **`persona-sre`**: System resilience and service orchestration.
*   **`persona-security`**: Vulnerability auditing and system hardening.
*   **`persona-readability`**: Maintainability and idiomatic code standards.
*   **`persona-ux`**: Design language, layout stability, and accessibility.
*   **`persona-legal`**: Licensing, copyrights, and SBOM compliance.
*   **`persona-oss`**: Upstream-first and community standards.
*   **`persona-privacy`**: Data handling and privacy regulation compliance.
*   **`persona-agent-manager`**: Skill lifecycle and agent orchestration.

## 4. Module-Specific Proximity Router

AI agents must follow local instructions when operating within a specific module:

*   👉 **[Preflight Dashboard (Frontend)](examples/preflight/AGENTS.md)**: SPA model, i18n standards, and local hot-patching.
*   👉 **[GNOME Desktop (Layer)](examples/images/gnome/AGENTS.md)**: Extension compatibility and systemd lifecycle.

---
👉 **[Full Documentation Index](README.md)**

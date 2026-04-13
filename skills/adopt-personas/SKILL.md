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
name: adopt-personas
description: Adopts specialized engineering roles (SRE, SEC, SWE, OSS, Legal, UX) for comprehensive code reviews, system audits, and multidimensional task execution. Use when asked to review code or ensure cross-domain quality.
---

# Adopt Persona-Driven Development

To ensure all tasks are handled with the appropriate depth and multi-dimensional focus, AI agents should dynamically adopt one or more of the following specialized personas.

## Available Personas

*   **SWE (Software Engineer)**: Focuses on functionality, behavioral correctness, and clean logic. (See `references/swe.md`)
*   **Readability (Code Reviewer)**: Enforces style guides, maintainability, and clear documentation. (See `references/readability.md`)
*   **SRE (Site-Reliability Engineer)**: Focuses on resilience, systemd orchestration, and health monitoring. (See `references/sre.md`)
*   **SEC (Security Expert)**: Audits for vulnerabilities, credential protection, and system hardening. (See `references/security.md`)
*   **OSS (Open-Source Expert)**: Ensures upstream-first standards and public codebase readiness. (See `references/oss.md`)
*   **Legal (Legal Expert)**: Verifies copyrights, authoritative license sourcing, and SBOM compliance. (See `references/legal.md`)
*   **UX (UX/UI Specialist)**: Maintains design language, layout stability, and accessibility. (See `references/ux.md`)

## How to Leverage Personas

1.  **Identify the Core Domain**: Analyze the user's request to determine which personas are most relevant.
2.  **Explicit Adoption**: At the start of a task, explicitly state which personas are being activated (e.g., *"Adopting SWE and Readability personas for this refactor"*).
3.  **Cross-Check Implementation**: For critical changes, perform a mental "Persona Cross-Check" (e.g., an SRE implementation reviewed through the lens of the SEC persona).
4.  **Consolidated Reasoning**: Use the persona descriptions in the `references/` directory to guide technical rationale and decision-making during the Plan and Act phases.

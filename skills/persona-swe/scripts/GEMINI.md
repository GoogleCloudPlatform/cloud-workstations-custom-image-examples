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

# GitSeep Architectural & Engineering Mandates

This document formalizes the architectural patterns and engineering standards for the `gitseep` engine and TUI dashboard.

## 1. Error Propagation & Troubleshooting

- **Mandate: Rich Failure Reporting**: All engine-level failures that occur during strata validation MUST be wrapped in a `models.RichPipelineError`.
- **Mandate: Metadata Carrying**: The error MUST carry a complete `models.PipelineEvent` containing:
  - Detailed `LogOutput`.
  - An automated `ReproductionCmd` (must use `xargs` and `SKIP=gitseep-check`).
  - The `WorktreeDir` where the failure occurred.
- **Mandate: Headless Parity**: Headless and TUI modes MUST use the same troubleshooting provider interface to deliver consistent guidance.

## 2. TUI Dashboard Architecture

- **Mandate: Navigation Priority**: The `DashboardModel.Update` loop MUST prioritize active overlays (Logs, Help, Failure) before global tab navigation. This ensures `Enter` and `Esc` reliably close dialogues.
- **Mandate: Permanent Auditability**: The Logs tab (Tab 1) MUST always be visible and accessible to ensure users can monitor streaming events at any time.
- **Mandate: Responsive Layout**: Sub-models (especially Validation) MUST dynamically scale their diagnostic areas based on `availableHeight`. The geological tree visualization MUST be prioritized over logs if vertical space is constrained (< 15 lines).
- **Mandate: Uniform Terminology**: All interactive nodes and selection lists MUST use the "toggle" terminology for the `Space` key.

## 3. Geological Tree Visualization

- **Mandate: Geometric Parity**: All tree visualizations across different tabs MUST be driven by the shared `TreeContext` coordinate system.
- **Mandate: Lane Integrity**: Tree rendering MUST enforce a strict 2-character segment grid (`● `, `│ `, `├─`, `└─`, `┴─`) to ensure pixel-perfect vertical alignment of hashes and subjects.
- **Mandate: Independent Columns**: sibling branches (independent geological paths) MUST occupy unique columns. The lane assignment algorithm must ensure next-column availability for every new branch.
- **Mandate: Anchor Convergence**: The bedrock anchor MUST be rendered as a first-class node using `●` and show upward connectivity (`┴`) for all currently active lanes.

## 4. Engineering Standards

- **Mandate: Sorted Lists**: Utilize `go/keep-sorted` for all model fields and configuration entries with multiple items. Do not use for single-element collections.
- **Mandate: i18n Integrity**: No hardcoded user-facing strings are permitted. All labels, including tab numbers, MUST be driven by the dynamic `i18n` locale engine.
- **Mandate: Context Efficiency**: Sub-models MUST be subscribed to the internal `EventBus` to receive pipeline updates, avoiding direct channel propagation across the dashboard.

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

# UX Standards: Custom Image Preflight

This document defines the user experience and visual standards for the "Preflight" loading interface.

## Cinematic Experience Principles

The Preflight UI is designed to feel like a high-end dashboard rather than a technical loading screen.

1.  **Immediate Visual feedback**: The UI must reveal its primary layout (background and structural containers) within 1500ms of initial load, even if technical assets (e.g., translation dictionaries) are still processing.
2.  **Immersive Motion**:
    *   **Background**: Uses a slow "starfield drift" animation to provide a sense of life.
    *   **Pulsing State**: The primary "STARTING" text pulses subtly to indicate active processing.
    *   **Phased Transitions**: The progress ring rotates in sync with the nominal timeout threshold.
3.  **Clean Typography**: Utilizes "Space Grotesk" for headlines to maintain a modern, technical aesthetic.

## Technical Data Priority

To ensure operational accuracy, the Preflight dashboard enforces a strict data priority hierarchy:

1.  **Authoritative Server Meta**: Technical metadata provided by the workstation environment (Hostname, Uplink, Client IP) always takes precedence over stored values.
2.  **Explicit URL Parameters**: Temporary developer overrides via the address bar (e.g., `?lang=ar&debug=true`).
3.  **Local Storage**: Persisted user preferences (Language, Debug Mode) that do not impact technical session integrity.

## Configuration & SPA Interaction

The dashboard follows a **Single Page Application** state model:

1.  **Zero-Reload saving**: Clicking **Save** updates the `state` singleton and the URL (`history.replaceState`) without refreshing the page.
2.  **Sticky Modal Architecture**: To ensure accessibility on small viewports, modals (e.g., Settings) utilize a fixed header/footer with internal scrolling content. Actions (Save/Reset) are always visible.
3.  **Adaptive Layout & Exclusive Expansion**:
    *   **Exclusive Sectioning**: Expanding "Basic Settings" (Connection/Language) automatically collapses "Advanced Settings," and vice versa, to maximize viewport utilization.
    *   **Adaptive Slider States**: The **Retry Interval** slider is permanently visible within Advanced Settings but is dynamically disabled and dimmed when "Auto Redirect" is unchecked.
4.  **Implicit Rollback & Transient Staging**:
    *   **Staging Model**: The UI captures a "snapshot" of unsaved changes when navigating between modals (e.g., from Settings to Language) to ensure no user input is lost.
    *   **Rollback**: Closing a modal (via 'X' or Escape) without saving discards UI changes and restores the active `state.config` values.
5.  **Precision Parameter Mapping**:
    *   **Timeout (1s to 500s)**: Logarithmic curve provides high precision for low-latency environments while supporting long startup thresholds.
    *   **Retry Interval (100ms to 10s)**: High-precision logarithmic control over backend polling frequency.
6.  **Keyboard Interaction & Shortcuts**:
    *   **Modal Gating**: Global shortcuts (except Escape) are automatically restricted when a modal is active to prevent unintentional state changes.
    *   **Aliased triggers**: Primary actions support intuitive aliases (e.g., both **H** and **?** trigger the Help system) to accommodate different keyboard layouts and user habits.
    *   **Grouped Reference**: The shortcut reference modal utilizes a responsive multi-column grid to maintain a compact, easily scannable list of operations.
7.  **Session Re-simulation**: Saving settings triggers a functional reset—clearing the timer to 00:00 and restarting the health check sequence.

## Documentation & Help Architecture

The documentation is organized for progressive disclosure:

1.  **Centralized Help**: The Help modal serves as the gateway to all non-operational information.
2.  **Internal Rendering**: Key documents (Privacy Notice) are pre-rendered into HTML fragments during the build process and served locally, ensuring high performance and offline availability.
3.  **Bridged Modals**: Links within the Help modal (Licenses, Shortcuts, Privacy) trigger respective specialized modals.
    *   **Stateless Entry**: Specialized modals explicitly reset their navigation stack upon entry, ensuring the user always starts at the primary index (e.g., the root license list).
    *   **Proactive Rendering**: To ensure technical consistency, modal content (such as the software list) is re-rendered upon each entry to synchronize with the latest application state.

## Global Standardization

*   **Protocol Identification**: Connection types must use professional, descriptive labels (e.g., "Remote Desktop Protocol (RDP)") rather than technical IDs.
*   **Units**: Time-based technical metrics must display localized units (**s** for seconds, **ms** for milliseconds).
*   **Attribution precision**: The software manifest must provide context-aware metadata (e.g., identifying creators as "Authors" for assets and "Suppliers" for technical packages).
*   **Accessibility**:
    *   Interactive elements must have clear focus states (`focus-ring`).
    *   Dynamic regions (Status, Timer, Debug Panel) utilize `aria-live` to ensure real-time screen reader announcements.

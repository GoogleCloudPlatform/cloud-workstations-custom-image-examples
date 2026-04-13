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
name: enforce-standards
description: Enforces coding standards, strict typing, and architectural conventions when modifying source code. Use when writing, refactoring, or reviewing code in Bash, TypeScript, Dockerfiles, Systemd units, or Python.
---

# Enforce Coding Standards & Style Guides

This skill provides the mandatory language-specific style guides and coding conventions for the repository.

**MANDATE:** When modifying code in this repository, you must strictly adhere to the appropriate language-specific style guides located in the `docs/style_guides/` directory.

## Core Mandates

*   **i18n:** Full i18n support is required. You MUST ensure that all translations for all text strings are in place across all supported languages when modifying user-facing text.
*   **Examples:** For all examples (code, configuration, or documentation), you MUST use `example.com` for companies/enterprises, `example.net` for ISPs, and `example.org` for NGOs/Non-profits.
*   **Sorted Lists:** Utilize `go/keep-sorted` directives ([google/keep-sorted](https://github.com/google/keep-sorted)) to maintain alphabetical order in lists of imports, metadata, CLI arguments, and other relevant collections.

## Language-Specific Guides

When working with specific file types, consult the appropriate authoritative reference:

*   **Bash (`.sh`, `.bash`, `.bats`)**: See [../../docs/style_guides/bash.md](../../docs/style_guides/bash.md).
*   **TypeScript (`.ts`, `.tsx`)**: See [../../docs/style_guides/typescript.md](../../docs/style_guides/typescript.md).
*   **Docker (`Dockerfile`)**: See [../../docs/style_guides/docker.md](../../docs/style_guides/docker.md).
*   **Systemd (`.service`, `.socket`, etc.)**: See [../../docs/style_guides/systemd.md](../../docs/style_guides/systemd.md).
*   **Python (`.py`)**: See [../../docs/style_guides/python.md](../../docs/style_guides/python.md).

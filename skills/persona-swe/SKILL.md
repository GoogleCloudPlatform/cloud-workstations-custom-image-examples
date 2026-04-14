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
name: persona-swe
description: Adopts the Software Engineer persona. Focuses on functional correctness, structural integrity, and architectural hygiene, including Source Code Versioning History Refactoring.
---

# Persona: Software Engineer (SWE)

## Mission
To implement robust, functional, and logically sound solutions that solve user problems efficiently. The SWE persona prioritizes behavioral correctness and structural integrity.

## Core Responsibilities
- **Functional Implementation**: Write code that fulfills the technical requirements of the task.
- **Architectural Alignment**: Consolidate logic into clean abstractions rather than threading state across unrelated layers.
- **Test-Driven Reliability**: Always accompany changes with comprehensive unit and integration tests.
- **Source Code Versioning History Refactoring**: Proactively clean up repository history to maintain a clean, logical progression of changes.
- **Dependency Management**: Use established project libraries and frameworks; avoid introducing redundant dependencies.

## Source Code Versioning History Refactoring

This skill provides specialized tooling for "Synthetic History Reconstruction." This technique is safer than interactive rebasing for large-scale architectural cleanups where changes need to be moved back into specific base commits without dealing with intermediate merge conflicts.

### Logic Visualization
```mermaid
graph TD
    subgraph Original History
    C1[Commit 1: Base for /foo] --> C2[Commit 2: Changes /foo & /bar]
    C2 --> C3[Commit 3: Changes /foo]
    end

    subgraph Refactored History
    R1[Commit 1': Final /foo state established] --> R2[Commit 2': Changes to /foo removed, /bar kept]
    R2 --> R3[Commit 3': Changes to /foo removed, Empty commit kept]
    end

    C1 -. "Force Final State" .-> R1
    C2 -. "Force Revert /foo" .-> R2
    C3 -. "Force Revert /foo" .-> R3
```

### The Refactoring Tool
The `git_refactor.py` script (located in `scripts/`) allows you to consolidate path-based changes into earlier commits while preserving the exact commit count and messages of the original branch.

**Workflow:**
1.  **Preparation**: Create a `refactor_rules.yaml` file (use `refactor_rules.yaml.example` as a template).
2.  **Stable IDs**: Mapping uses **Author Date** strings (unique within the branch) to identify base commits, making the rules resilient to hash changes.
3.  **Interactive Run**: Execute the script and follow the step-by-step prompts.
4.  **Finalization**: By default, the script updates your **current local branch** (after confirmation). Use `--branch` to target a different branch.

```bash
# Example: Refactor and update the CURRENT branch (default)
python3 scripts/git_refactor.py

# Example: Refactor and put results on a specific branch
python3 scripts/git_refactor.py --branch clean-history-v2

# Example: Specify a custom rules file
python3 scripts/git_refactor.py --rules my_rules.yaml
```

### Configuration (YAML)
The refactoring rules are defined in a YAML file using the following structure:
```yaml
"2026-04-15 10:00:00 +0000":
  - path/to/consolidate/
  - another/path/
```
All changes to these paths from later commits in the branch history will be moved back to the base commit matching that timestamp.

#### Retrieving the ISO Date
To see the full history with ISO dates:
```bash
git log --date=iso
```

To get the exact **Author Date** for a specific commit, run:
```bash
git log -1 --format=%ai <COMMIT_HASH>
```
Example Output: `2026-04-13 12:24:06 +0000` (Use this entire string as the key in your YAML).

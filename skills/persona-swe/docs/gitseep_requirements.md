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

# Product Requirements Document: GitSeep

## 1. Objective and Background

GitSeep is a highly opinionated Source Code History Orchestrator designed to solve the cognitive load and complexity of maintaining a pristine, structurally sound commit history. Specifically, it enables a "Flow-State Workflow" where developers and AI agents can develop linearly on `HEAD` and rely on GitSeep to autonomously sort changes back into their logical historical chapters ("strata") before publishing.

The primary goal is to provide a safe, automated alternative to complex `git rebase -i` operations, mathematically guaranteeing the preservation of code states while eliminating merge conflicts.

## 2. Target Audience

- **Human Developers:** Seeking to present a perfectly clean, logical narrative to reviewers without the toil of manual conflict resolution.
- **AI Agents (Agentic Workflows):** Agents struggle with multi-file conflict resolution during interactive rebases. GitSeep allows agents to commit linearly and defer structural organization to an automated process.
- **Reviewers:** Benefiting from atomic, self-contained commits that represent a clear architectural progression.

## 3. Core Requirements (Functional)

### 3.1 Declarative Configuration

- **REQ-FUNC-1:** The system MUST support a declarative `.gitseep.yaml` file to map file paths to specific historical "Bedrock" commits (identified by Author Date and branch).
- **REQ-FUNC-2:** Path matching MUST resolve conflicts by prioritizing the most specific (longest) path.

### 3.2 The Execution Pipeline

- **REQ-FUNC-3:** The system MUST be capable of parsing the `.gitseep.yaml` rules, analyzing the `HEAD` state, and mathematically projecting the target files backward in time into their designated commits.
- **REQ-FUNC-4:** The system MUST support a Terminal User Interface (TUI) phase to allow human users to selectively review, approve, or exclude identified migrations before any changes are written to the git database.
- **REQ-FUNC-5:** The system MUST autonomously generate and update a Directed Acyclic Graph (DAG) of feature branches (Stacked PRs) based on dependencies defined in the `.gitseep.yaml` file.

### 3.3 Zero-Loss Guarantee

- **REQ-FUNC-6:** The system MUST guarantee that any files present at the surface (`HEAD`) that are not managed by explicit rules are captured and injected into the final orchestration commit to prevent data loss.

### 3.4 CI/CD and Validation Integration

- **REQ-FUNC-7:** The system MUST provide a `check` subcommand (`gitseep check`) that validates stratigraphy without modifying state, returning standard exit codes for integration into pre-commit hooks and CI pipelines.
- **REQ-FUNC-8:** The system MUST natively execute an automated validation phase on all newly reconstructed commits and feature branches _before_ mutating the user's workspace. To ensure scalability, it MUST isolate execution using a lightweight `git worktree`, use `go-git` to calculate the precise file modifications for each commit, and strictly run `pre-commit` against those specific files (via `--files`) to avoid triggering manual-stage or full-repository hooks. It MUST deduplicate test runs by commit hash and immediately abort on failure. Users can opt-out via a `--skip-pre-commit` flag.
- **REQ-FUNC-9:** The system MUST provide real-time visual feedback during the validation phase via a dedicated TUI, displaying a global progress bar and a live-updating list of commit/branch statuses in a "Gravity" layout (newest first). The validation MUST follow a "Tips to Bedrock" order to provide immediate feedback on recent changes.
- **REQ-FUNC-10:** The system MUST provide a high-fidelity "Shell Escape" mechanism upon validation failure. This mechanism MUST automatically identify the failing files and execute a targeted `pre-commit run` within the sub-shell to facilitate immediate troubleshooting.

## 4. Opinionated Constraints (Non-Functional)

### 4.1 The Zero-Mutation Guarantee (Critical Mandate)

This is the foundational philosophical stance of GitSeep.

- **REQ-NF-1:** The system MUST NOT rewrite a commit hash if the mathematical structure (Tree + Parent + Message) of the reconstructed commit is identical to the original commit.
- **REQ-NF-2:** The system MUST achieve absolute diff parity across branches. Each feature branch commit MUST reflect the exact change set (patch) of its corresponding linear stratum, ensuring that `git show --name-only` produces identical results for both `main` and feature branches. While tree states remain consistent, commit hashes will correctly differ when the feature branch skips intermediate geological strata (e.g., intermediate linear commits not part of the branch hierarchy).

### 4.2 In-Memory Operation & State Integrity

- **REQ-NF-3:** Core reconstruction MUST occur entirely in-memory using the Git object database (e.g., via `go-git`). It MUST NOT rely on expensive or error-prone physical disk I/O or shell subprocesses for history generation.
- **REQ-NF-4:** The system MUST guarantee thread-safe operations during concurrent discovery of repository objects.

### 4.3 Workspace Safety

- **REQ-NF-5:** The system MUST NEVER corrupt, delete, or modify untracked files in the developer's working directory during its final synchronization phase.
- **REQ-NF-6:** Upon a validation failure during the Isolated Validation phase, the system MUST immediately halt execution, leave the primary workspace untouched, and deliberately preserve the isolated temporary `git worktree` to allow the user or agent to securely inspect the failure state. The TUI MUST provide descriptive technical output including the raw logs from the failing process.
- **REQ-NF-7 (Multi-Agent Safety):** To prevent race conditions in environments where multiple AI agents or automated processes may operate concurrently, GitSeep MUST NOT perform an implicit `git add` during its execution. The system always performs a mandatory staged-only `git commit --amend` before starting the reconstruction pipeline. This ensures that only modifications explicitly staged by the user or agent are folded into the geological history, maintaining a safe and deliberate "Flow-State" workflow.
- **REQ-NF-8 (Non-Destructive Finalization):** To ensure safety during concurrent workspace mutation, the system's final synchronization phase MUST NOT use destructive operations (like `git checkout -f`) that overwrite uncommitted worktree changes. Instead, it MUST use a "Safe Pointer Move" strategy (e.g., `git reset --mixed`) to update the branch pointer and staging area (index) to match the new history while leaving all unstaged and untracked modifications in the worktree untouched.

## 5. Metrics & Telemetry

- The system should produce a summary report upon completion detailing the number of strata processed, files lithified (squashed), and branches successfully sedimented.

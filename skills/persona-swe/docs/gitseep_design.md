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

# Design Document: GitSeep

## 1. Context and Scope

GitSeep is a CLI tool written in Go (1.24+). It orchestrates git history by reading declarative rules and projecting code changes backward in time into specific historical commits ("bedrock"). Unlike tools that rely on patch application (`git rebase`), GitSeep operates mathematically on the Git Object Database, enabling conflict-free, high-speed organization of complex architectures.

This document describes the internal architecture of GitSeep, focusing on the Go implementation, its 8-phase pipeline, and the core opinionated design decisions that govern its behavior.

## 2. System Architecture: The 8-Phase Pipeline

GitSeep utilizes a decoupled architecture, separating data discovery from user interaction and final execution. It heavily leverages `github.com/go-git/go-git/v5`.

### Phase 0: Concurrent Discovery & Topology Validation

The engine reads the target `Commit` and `Tree` objects directly from `.git/objects` into memory.

- **Memory Footprint:** To ensure scalability, the engine limits the scope of its search explicitly to the commit range between the defined `base_ref` and `HEAD`.
- **Concurrency:** It spawns a worker `goroutine` for every commit in the target range using `golang.org/x/sync/errgroup`.
- **Thread Safety:** A `sync.Mutex` is utilized during initial object cache resolution to prevent map write panics inherent in `go-git`, while allowing subsequent `merkletrie` tree diffing to run entirely in parallel.

### Phase 1: Stratigraphy & Conflict Validation

Mathematical validation of the geological strata.

- **Topology:** Kahn's Algorithm validates the DAG topology of the proposed feature branches, instantly detecting cyclical dependencies.
- **Conflict Prediction:** Performs dry-run tree diffs to predict mathematical merge conflicts before any reconstruction occurs, providing early feedback to the user.

### Phase 2: Interactive Selection (Review)

The system transitions to an interactive Terminal Dashboard that orchestrates the entire pipeline.

- **Unified Interface:** Consolidates Stratigraphy, Selection, Validation, and Logs into a single navigable view using `Bubble Tea`.
- **Interactive Review:** Users can surgically toggle file migrations and monitor real-time validation progress.
- **High-Fidelity Trees:** Employs a 'Gravity' layout (modern work on top) with metadata-rich hierarchical trees.

### Phase 3: Surface Validation (Parallel Blocker)

Validates the current workspace (HEAD) against repository standards.

- **Parallel Execution:** Runs in parallel with Phases 0-2 to minimize developer idle time.
- **Execution Blocker:** Strictly blocks Phase 4 (Execution) until the surface is confirmed clean, ensuring no reconstruction begins on a broken foundation.
- **Fail-Fast:** If surface validation fails, the TUI exits immediately with descriptive technical output.

### Phase 4: Execution (Linear Reconstruction)

The core engine iterates through the history and constructs the new history entirely in RAM.

- **Merkletrie Deltas:** Calculates the delta required to force paths defined in `.gitseep.yaml` to their final, intended `HEAD` state.
- **Atomic Injection:** Builds nested `object.Tree` structures and writes new `Commit` objects directly into the database.

### Phase 5: Sedimentation (Autonomous Branch Sync)

Calculates the pointers for the configured feature branches.

- **Patch Projection:** To ensure absolute diff parity with the linear history (`main`) while correctly skipping intermediate strata, the engine utilizes a surgical "Patch Projection" strategy.
- **Change Set Calculation:** For each feature branch commit, the engine calculates the exact patch (diff) of the corresponding linear stratum in the reconstructed history.
- **Surgical Application:** This patch is applied directly to the feature branch's reconstructed parent. This ensures that `git show --name-only` for a feature branch commit will show exactly the same file list as its linear counterpart, even if they have different parents and hashes.
- **Global Exclusion Enforcement:** The engine explicitly removes any files present in `ExcludedPaths` (e.g., `CONTRIBUTING.md`) during the reconstruction of all branches, ensuring that unchecked files are consistently eliminated across the entire DAG.
- **Branch-Specific Preservation:** Truly branch-specific unmanaged files (those never present in the original repository history) are preserved and correctly inherited across the branch hierarchy.

### Phase 6: Strata Validation

Validates the generated history in an isolated environment.

- **Isolation:** Created using `git worktree add --detach`.
- **Deduplication:** Tests are cached by commit hash to optimize execution.
- **Shell Escape:** On failure, provides a "Shell Escape" (`s` key) to drop the user into the worktree with a pre-configured `pre-commit` run already executing for the failed files.

### Phase 7: Finalization & Workspace Sync

Updates the physical git references (branches) and synchronizes the user's physical workspace using system Git for maximum stability.

## 3. Key Design Decisions & Opinionated Views

### 3.1 The "Gravity" UI & Reverse Validation

**Decision:** The Dashboard and Validation pipeline MUST follow reverse geological gravity ("Tips to Bedrock").
**Rationale:** Developers care most about their recent work. By validating the modern surface first and then drilling down into history, GitSeep provides faster feedback on current changes while ensuring the bedrock foundations remain solid.

### 3.2 High-Fidelity Shell Escape

**Decision:** Validation failures MUST provide immediate, actionable troubleshooting paths.
**Implementation:** The 'Shell Escape' feature doesn't just drop the user into a directory; it detects exactly which files failed validation and automatically initiates a targeted `pre-commit run` in the sub-shell. This eliminates the "what broke?" discovery phase during debugging.

### 3.3 The In-Memory "Zero-Mutation" Guarantee

**Decision:** GitSeep MUST enforce a mathematical Zero-Mutation policy.
**Rationale:** Unnecessary commit hash changes trigger destructive consequences (breaking CI pipelines, spamming PR reviewers).
**Implementation:** Before writing a new commit, GitSeep compares state; if identical, the **original commit hash is reused**.

### 3.4 Safe Workspace Updates via System Git

**Decision:** Final workspace synchronization uses a system shell call: `git reset --hard` (or `git checkout -f`).
**Rationale:** `go-git` struggles with complex symlink structures and file modes. System Git is mathematically guaranteed to safely synchronize the worktree.

## 4. Data Models

- **`models.SeepageContext`:** The central state object. Holds resolved mappings, branch topologies, and configuration.
- **`models.DiscoveryResult`:** The output of Phase 0. Defines the work required for reconstruction.
- **`engine.TaskContextParams`:** Encapsulates dependencies and state for the orchestration pipeline, enabling robust dependency injection.
- **`engine.PreCommitParams`:** Configuration and resource handles for the isolated validation runner.
- **`models.PipelineEvent`:** Real-time telemetry carrying failure details, preserved worktree paths, and file lists for targeted debugging.

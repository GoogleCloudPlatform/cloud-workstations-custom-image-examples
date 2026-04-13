#!/usr/bin/env python3

# Copyright 2026 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     https://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

"""GitSeep: Architectural Bedrock Reconstruction Tool.

Technical Metaphor:
- Surface (HEAD): The current state of your working directory/index.
- Strata (Commits): The layers of history between a base commit and HEAD.
- Bedrock: The specific historical commit that "owns" a path (defined in rules).
- Seepage: Moving changes from the surface down through layers until they
           are permanently deposited in their respective bedrock commits.

This tool automates the process of "backfilling" a clean architectural state
into an unstructured evolution by re-applying the final state of specific paths
to their original introductory commits, ensuring a perfectly clean and
logically grouped historical record.
"""

# go/keep-sorted start
import argparse
import os
import re
import shutil
import subprocess
import sys
import termios
import tty
import uuid
# go/keep-sorted end
from dataclasses import dataclass, field
from typing import Dict, List, Optional, Set

import yaml

# go/keep-sorted start
__author__ = "Google LLC"
__license__ = "Apache 2.0"
__source__ = "https://github.com/sce-taid/cloud-workstations-custom-image-examples/blob/main/skills/persona-swe/scripts/gitseep.py"
__version__ = "1.1.0"
# go/keep-sorted end

# Default configuration names
RULES_FILENAME = ".gitseep.yaml"
EXAMPLE_FILENAME = ".gitseep.yaml.example"


# --- Exceptions ---


class Error(Exception):
    """Base class for exceptions in this module."""


class GitError(Error):
    """Raised when a git command fails."""


# --- Data Classes ---


@dataclass
class SeepageSummary:
    """Summary of the seepage process.

    Attributes:
        bedrock_files: Number of files deposited in their bedrock stratum.
        percolate_files: Number of files moved down to an earlier stratum.
        seep_files: Number of files moved up to a later stratum.
        compressed_files: Set of files that underwent history compression.
        parity_passed: Whether the final state matches the original HEAD.
    """

    bedrock_files: int = 0
    percolate_files: int = 0
    seep_files: int = 0
    compressed_files: Set[str] = field(default_factory=set)
    parity_passed: bool = False


@dataclass
class SeepageContext:
    """Context for the seepage process.

    Attributes:
        repo_root: Absolute path to the git repository root.
        original_head: Original HEAD hash before processing.
        current_branch: Branch name active at start.
        target_branch: Branch to update with refactored history.
        parent_of_strata: Parent of the first commit in the refactor range.
        strata: Chronological list of commit hashes in the range.
        resolved_rules: Mapping of bedrock hashes to list of owned paths.
        path_to_bedrock: Mapping of paths to their owning bedrock hash.
        all_files: If true, show all files in bedrock, not just changes.
        auto_approve: If true, skip non-critical user confirmations.
        compress: If true, permit multi-stratum compression automatically.
        stage_only: If true, do not update target_branch; leave result in temp.
    """

    repo_root: str
    original_head: str
    current_branch: str
    target_branch: str
    parent_of_strata: str
    strata: List[str]
    resolved_rules: Dict[str, List[str]]
    path_to_bedrock: Dict[str, str]
    all_files: bool
    auto_approve: bool
    compress: bool
    stage_only: bool


# --- Git Helpers ---


def run_git(
    args: List[str], check: bool = True, env: Optional[Dict[str, str]] = None
) -> subprocess.CompletedProcess:
    """Runs a git command and returns the result.

    Args:
        args: List of command-line arguments for git.
        check: If True, raise GitError on non-zero exit code.
        env: Optional environment variables for the command.

    Returns:
        The result of the command execution.

    Raises:
        GitError: If the command fails and check is True.
    """
    try:
        return subprocess.run(
            ["git"] + args,
            capture_output=True,
            text=True,
            check=check,
            env=env or os.environ,
        )
    except subprocess.CalledProcessError as e:
        error_msg = f"Error executing git {' '.join(args)}: {e.stderr}"
        if check:
            print(error_msg, file=sys.stderr)
            raise GitError(error_msg) from e
        return subprocess.CompletedProcess(
            e.args, e.returncode, e.stdout, e.stderr
        )


def get_git_root() -> str:
    """Returns the absolute path to the root of the git repository."""
    return run_git(["rev-parse", "--show-toplevel"]).stdout.strip()


def get_commit_sequence(start_hash: str) -> List[str]:
    """Returns chronological commit hashes from start_hash to HEAD.

    Args:
        start_hash: The hash of the first commit in the sequence.

    Returns:
        List of commit hashes (oldest to newest).
    """
    result = run_git(["log", "--format=%H", f"{start_hash}^..HEAD"])
    return result.stdout.strip().split("\n")[::-1]


def resolve_commit_by_date(date_str: str) -> str:
    """Finds a unique commit hash matching an author date string.

    Args:
        date_str: An author date string (e.g., from 'git log --format=%ai').

    Returns:
        A unique commit hash.

    Raises:
        GitError: If no commit matches or if multiple non-linear matches exist.
    """
    result = run_git(["log", "--all", "--format=%H %ai"])

    matches = []
    for line in result.stdout.strip().splitlines():
        if line.endswith(date_str):
            matches.append(line.split(" ")[0])

    if not matches:
        raise GitError(f"No commit found matching date: {date_str}")

    if len(matches) > 1:
        current_history = (
            run_git(["log", "--format=%H"]).stdout.strip().splitlines()
        )
        branch_matches = [m for m in matches if m in current_history]
        if len(branch_matches) == 1:
            return branch_matches[0]
        raise GitError(
            f"Multiple commits found matching date {date_str}: {matches}"
        )

    return matches[0]


def force_path_state(path: str, source_tree: str):
    """Forces the index and working tree of 'path' to match 'source_tree'.

    Args:
        path: File or directory path to reset.
        source_tree: A git tree-ish (e.g., hash, branch name) to copy from.
    """
    # 1. Completely remove from index and working tree
    run_git(["rm", "-rf", "--cached", "--ignore-unmatch", path], check=False)
    if os.path.exists(path):
        if os.path.isdir(path):
            shutil.rmtree(path)
        else:
            os.remove(path)

    # 2. Restore from source_tree if it exists there
    check = run_git(
        ["ls-tree", "-r", source_tree, "--name-only", path], check=False
    )
    if check.stdout.strip():
        run_git(["checkout", source_tree, "--", path])


# --- Terminal Helpers ---


def strip_ansi(text: str) -> str:
    """Removes ANSI escape sequences from a string."""
    return re.compile(r"\x1B(?:[@-Z\\-_]|\[[0-?]*[ -/]*[@-~])").sub("", text)


def get_single_key(prompt: str) -> str:
    """Reads a single key from stdin without requiring Enter.

    Args:
        prompt: Text to display before reading input.

    Returns:
        The character pressed (lowercase).
    """
    print(prompt, end="", flush=True)
    fd = sys.stdin.fileno()
    if not os.isatty(fd):
        line = sys.stdin.readline()
        return line[0].lower() if line else ""

    old_settings = termios.tcgetattr(fd)
    try:
        tty.setcbreak(fd)
        ch = sys.stdin.read(1)
    except KeyboardInterrupt:
        print("\nAborted.")
        sys.exit(1)
    finally:
        termios.tcsetattr(fd, termios.TCSADRAIN, old_settings)

    print(ch)  # Echo the key
    return ch.lower()


def user_confirm(prompt: str) -> bool:
    """Prompts the user for Y/n confirmation (single key)."""
    while True:
        choice = get_single_key(f"{prompt} [Y/n]: ")
        if not choice or choice in ("y", "\n", "\r"):
            return True
        if choice == "n":
            return False


# --- UI Helpers ---


def get_path_diff_stats(h1: str, h2: str, path: str) -> str:
    """Returns colored line-level diff stats for a path between two commits.

    Args:
        h1: Bedrock hash.
        h2: Source hash.
        path: File path to compare.

    Returns:
        A string like " (+5/-2)" with ANSI colors.
    """
    res = run_git(
        ["diff", "--shortstat", f"{h1}..{h2}", "--", path], check=False
    ).stdout.strip()

    if not res:
        return ""

    insertions = 0
    deletions = 0

    parts = res.split(",")
    for p in parts:
        if "insertion" in p:
            insertions = int(p.strip().split(" ")[0])
        if "deletion" in p:
            deletions = int(p.strip().split(" ")[0])

    stats = []
    if insertions > 0:
        stats.append(f"\033[32m+{insertions}\033[0m")
    if deletions > 0:
        stats.append(f"\033[31m-{deletions}\033[0m")

    if not stats:
        return " (ident)"
    return f" ({'/'.join(stats)})"


def _print_preflight_briefing(ctx: SeepageContext):
    """Prints the pre-flight briefing showing the refactor strata."""
    print("\n--- Pre-Flight Seepage Briefing ---")
    print(f"Target Branch: {ctx.target_branch}")
    print(f"Surface Hash:  {ctx.original_head[:8]}")
    print("-" * 70)
    hdr = (
        f"{'Stratum':<7} {'[Commit]':<10} {'(Author Date)':<27} - {'Message'}"
    )
    print(hdr)
    print("-" * 70)

    for i, h in enumerate(reversed(ctx.strata)):
        log_out = run_git(["log", "-1", "--format=%ai|%s", h]).stdout.strip()
        commit_info = log_out.split("|", 1)
        date = commit_info[0]
        msg = commit_info[1]

        layer_num = len(ctx.strata) - i
        layer_str = f"{layer_num}/{len(ctx.strata)}"
        print(f"{layer_str:<7} [{h[:8]}] ({date}) - {msg}")
        if h in ctx.resolved_rules:
            owned = ", ".join(ctx.resolved_rules[h])
            print(f"        ⮑  Bedrock for: {owned}")

    print("-" * 70)


# --- Seepage Core ---


def _initialize_seepage_context(
    rules_dict: Dict[str, List[str]],
    target_branch: str,
    all_files: bool,
    auto_approve: bool,
    compress: bool,
    stage_only: bool,
    base_commit: Optional[str],
) -> SeepageContext:
    """Validates repo state and prepares the seepage context."""
    repo_root = get_git_root()
    os.chdir(repo_root)

    status = run_git(["status", "--porcelain"])
    if status.stdout.strip():
        print("Error: Working directory is not clean.")
        sys.exit(1)

    original_head = run_git(["rev-parse", "HEAD"]).stdout.strip()
    current_branch = (
        run_git(["rev-parse", "--abbrev-ref", "HEAD"]).stdout.strip()
    )

    # Resolve date-based rules to hashes
    resolved_rules = {}
    try:
        for date_key, paths in rules_dict.items():
            commit_hash = resolve_commit_by_date(str(date_key))
            resolved_rules[commit_hash] = paths
    except GitError as e:
        print(f"Error: {e}")
        sys.exit(1)

    # Determine start of refactor range
    if base_commit:
        try:
            earliest_hash = run_git(["rev-parse", base_commit]).stdout.strip()
        except GitError as e:
            print(f"Error: Invalid base commit '{base_commit}': {e}")
            sys.exit(1)
    else:
        # Sort by timestamp to find the oldest
        sorted_hashes = sorted(
            resolved_rules.keys(),
            key=lambda x: int(
                run_git(["log", "-1", "--format=%at", x]).stdout.strip()
            ),
        )
        earliest_hash = sorted_hashes[0]

    parent_of_strata = (
        run_git(["rev-parse", f"{earliest_hash}^"]).stdout.strip()
    )
    strata = get_commit_sequence(earliest_hash)

    path_to_bedrock = {}
    for h, paths in resolved_rules.items():
        for p in paths:
            path_to_bedrock[p] = h

    return SeepageContext(
        repo_root=repo_root,
        original_head=original_head,
        current_branch=current_branch,
        target_branch=target_branch,
        parent_of_strata=parent_of_strata,
        strata=strata,
        resolved_rules=resolved_rules,
        path_to_bedrock=path_to_bedrock,
        all_files=all_files,
        auto_approve=auto_approve,
        compress=compress,
        stage_only=stage_only,
    )


def _calculate_percolation_schedule(ctx: SeepageContext):
    """Calculates which files belong to which bedrock commits."""
    # bedrock_hash -> source_hash -> [files]
    percolation_schedule = {}
    # file -> set(source_hashes)
    file_to_sources = {}

    for commit in ctx.strata:
        res = run_git(
            ["diff-tree", "--no-commit-id", "--name-only", "-r", commit]
        )
        changed = res.stdout.strip().splitlines()
        for f in changed:
            # Match file to most specific owned path
            matching_owners = [
                p for p in ctx.path_to_bedrock.keys() if f.startswith(p)
            ]
            if not matching_owners:
                continue

            owner_path = max(matching_owners, key=len)
            bedrock_hash = ctx.path_to_bedrock[owner_path]

            if bedrock_hash not in percolation_schedule:
                percolation_schedule[bedrock_hash] = {}
            if commit not in percolation_schedule[bedrock_hash]:
                percolation_schedule[bedrock_hash][commit] = []
            percolation_schedule[bedrock_hash][commit].append(f)

            if f not in file_to_sources:
                file_to_sources[f] = set()
            file_to_sources[f].add(commit)

    return percolation_schedule, file_to_sources


def _show_compression_diffs(ctx: SeepageContext, f: str, sources: Set[str]):
    """Displays diffs between historical sources and the final surface."""
    for s in sorted(sources, key=lambda x: ctx.strata.index(x), reverse=True):
        s_msg = run_git(["log", "-1", "--format=%s", s]).stdout.strip()
        s_stats = get_path_diff_stats(ctx.original_head, s, f)
        hdr = f"\n--- Diff: [{s[:8]}]{s_stats} ({s_msg}) -> Surface ---"
        print(hdr)

        res = run_git(
            ["diff", "--color", s, ctx.original_head, "--", f], check=False
        )
        if res.returncode != 0:
            print(f"  (Error: {res.stderr.strip()})")
        elif not res.stdout.strip():
            print("  (States are identical)")
        else:
            print(res.stdout)


def _handle_compression_interactive(
    ctx: SeepageContext, commit_hash: str, f: str, sources: Set[str]
) -> bool:
    """Handles user interaction for multi-stratum compression.

    Args:
        ctx: Seepage context.
        commit_hash: Current bedrock hash.
        f: File undergoing compression.
        sources: Set of source hashes that modified this file.

    Returns:
        True if compression is permitted.
    """
    print(f"  [COMPRESSION] File '{f}' modified in {len(sources)} strata.")
    for s in sorted(sources, key=lambda x: ctx.strata.index(x), reverse=True):
        s_msg = run_git(["log", "-1", "--format=%s", s]).stdout.strip()
        s_stats = get_path_diff_stats(commit_hash, s, f)
        print(f"    - [{s[:8]}]{s_stats} - {s_msg}")

    if ctx.compress:
        return True

    if ctx.auto_approve:
        print(f"Error: Compression detected for '{f}' but --compress not set.")
        sys.exit(3)

    _show_compression_diffs(ctx, f, sources)
    prompt = f"\nPermit compression for '{f}' from [{ctx.original_head[:8]}]?"
    return user_confirm(prompt)


def _handle_bedrock_path(
    ctx: SeepageContext,
    summary: SeepageSummary,
    commit_hash: str,
    owned_path: str,
    percolation_schedule: Dict,
    file_to_sources: Dict,
):
    """Handles an owned path at its bedrock stratum."""
    print(f"  [Bedrock] Capturing final surface sediment for '{owned_path}'")

    if commit_hash in percolation_schedule:
        # Collect unique files belonging to this specific bedrock ownership
        all_landing_files = set()
        for files in percolation_schedule[commit_hash].values():
            for f in files:
                owners = [
                    p for p in ctx.path_to_bedrock.keys() if f.startswith(p)
                ]
                if owners and max(owners, key=len) == owned_path:
                    all_landing_files.add(f)

        for f in sorted(all_landing_files):
            sources = file_to_sources[f]
            if len(sources) > 1:
                summary.compressed_files.add(f)
                if not _handle_compression_interactive(
                    ctx, commit_hash, f, sources
                ):
                    print(f"  Skipping compression: '{f}' remains in original.")
                    continue

            summary.bedrock_files += 1
            if commit_hash not in sources:
                src_list = ", ".join([s[:8] for s in sorted(sources)])
                print(f"  [{src_list}] {f} *")
            elif ctx.all_files:
                print(f"  [{commit_hash[:8]}] {f}")

    force_path_state(owned_path, ctx.original_head)


def _handle_seepage_path(
    ctx: SeepageContext,
    summary: SeepageSummary,
    commit_hash: str,
    owned_path: str,
    bedrock_commit: str,
    percolation_schedule: Dict,
):
    """Handles an owned path that is NOT at its bedrock stratum."""
    files_from_here = []
    if (
        bedrock_commit in percolation_schedule
        and commit_hash in percolation_schedule[bedrock_commit]
    ):
        files_from_here = [
            f
            for f in percolation_schedule[bedrock_commit][commit_hash]
            if max(
                [p for p in ctx.path_to_bedrock.keys() if f.startswith(p)],
                key=len,
            )
            == owned_path
        ]

    if files_from_here:
        current_idx = ctx.strata.index(commit_hash)
        bedrock_idx = ctx.strata.index(bedrock_commit)

        if current_idx < bedrock_idx:
            print(f"  [↑ Seep] Seeping {len(files_from_here)} up to Bedrock:")
            summary.seep_files += len(files_from_here)
        else:
            print(f"  [↓ Percolate] Percolating {len(files_from_here)} down:")
            summary.percolate_files += len(files_from_here)

        for f in files_from_here:
            print(f"  [{bedrock_commit[:8]}] {f}")

    force_path_state(owned_path, "HEAD")


def _seal_stratum(commit_hash: str):
    """Commits current index with original commit's metadata."""
    log_out = run_git(
        ["log", "-1", "--format=%an%n%ae%n%ad%n%cn%n%ce%n%cd", commit_hash]
    ).stdout.splitlines()

    if len(log_out) < 6:
        # Fallback to current user if log fails for some reason
        run_git(["commit", "--allow-empty", "-C", commit_hash])
        return

    custom_env = os.environ.copy()
    custom_env.update(
        {
            "GIT_AUTHOR_NAME": log_out[0],
            "GIT_AUTHOR_EMAIL": log_out[1],
            "GIT_AUTHOR_DATE": log_out[2],
            "GIT_COMMITTER_NAME": log_out[3],
            "GIT_COMMITTER_EMAIL": log_out[4],
            "GIT_COMMITTER_DATE": log_out[5],
        }
    )

    run_git(["commit", "--allow-empty", "-C", commit_hash], env=custom_env)


def _process_single_stratum(
    ctx: SeepageContext,
    summary: SeepageSummary,
    commit_hash: str,
    layer_num: int,
    schedule: Dict,
    sources: Dict,
):
    """Processes a single historical layer (commit)."""
    msg_full = run_git(["log", "-1", "--format=%B", commit_hash]).stdout.strip()
    msg_first = msg_full.splitlines()[0] if msg_full else "No message"
    msg = f"\n👉 {layer_num}/{len(ctx.strata)} [{commit_hash[:8]}] - {msg_first}"
    print(msg)

    # Reset index and working tree to this stratum's baseline
    run_git(["read-tree", "-u", "--reset", commit_hash])

    # Apply bedrock/seepage for each owned path
    for owned_path in sorted(ctx.path_to_bedrock.keys(), key=len):
        bedrock_commit = ctx.path_to_bedrock[owned_path]

        if commit_hash == bedrock_commit:
            _handle_bedrock_path(
                ctx, summary, commit_hash, owned_path, schedule, sources
            )
        else:
            _handle_seepage_path(
                ctx, summary, commit_hash, owned_path, bedrock_commit, schedule
            )

    _seal_stratum(commit_hash)


def _finalize_seepage(
    ctx: SeepageContext, summary: SeepageSummary, temp_branch_name: str
):
    """Finalizes the process, verifies parity, and updates branch."""
    print("\n--- Percolation Complete ---")
    print("Verifying sedimentation parity with original HEAD...")
    diff = run_git(["diff", ctx.original_head], check=False).stdout.strip()

    summary.parity_passed = not diff
    if not summary.parity_passed:
        print("WARNING: Final state differs from original surface (HEAD).")
        if not ctx.auto_approve and user_confirm("Show diff?"):
            print(diff)
    else:
        print("SUCCESS: Parity check passed. History matches original state.")

    # Report
    print("\n--- Seepage Summary Report ---")
    print(f"Strata Processed:  {len(ctx.strata)}")
    print(f"Bedrock Files:     {summary.bedrock_files}")
    print(f"Percolate Actions: {summary.percolate_files}")
    print(f"Seep Actions:      {summary.seep_files}")
    if summary.compressed_files:
        print(f"Compressed Files:  {len(summary.compressed_files)}")
        for f in sorted(summary.compressed_files):
            print(f"  - {f}")
    status_str = "PASSED" if summary.parity_passed else "FAILED"
    print(f"Parity Status:     {status_str}")
    print("-" * 30)

    # Finalization logic
    if ctx.stage_only:
        print(f"\n--- Stage-Only Mode ---")
        print(f"Refactored history prepared on: {temp_branch_name}")
        print(f"To finalize: git checkout {ctx.target_branch} && "
              f"git reset --hard {temp_branch_name}")
        return

    prompt = f"\nFinalize: Update branch '{ctx.target_branch}'?"
    if not ctx.auto_approve and not user_confirm(prompt):
        msg = (
            f"Finalization skipped. Cleaned history remains on "
            f"'{temp_branch_name}'."
        )
        print(msg)
        print(f"To manually finalize: git checkout {ctx.target_branch} && "
              f"git reset --hard {temp_branch_name}")
        run_git(["checkout", ctx.current_branch])
        return

    run_git(["checkout", "-B", ctx.target_branch, temp_branch_name])
    run_git(["branch", "-D", temp_branch_name])
    print(f"\nBranch '{ctx.target_branch}' has been updated.")


# --- Main ---


def seep_history(
    rules_dict: Dict[str, List[str]],
    target_branch: str,
    all_files: bool = False,
    auto_approve: bool = False,
    compress: bool = False,
    stage_only: bool = False,
    base_commit: Optional[str] = None,
):
    """Refactors git history by percolating changes to bedrock commits."""
    ctx = _initialize_seepage_context(
        rules_dict,
        target_branch,
        all_files,
        auto_approve,
        compress,
        stage_only,
        base_commit,
    )

    _print_preflight_briefing(ctx)

    temp_branch_name = f"gitseep-work-{uuid.uuid4().hex[:8]}"
    run_git(["checkout", "-b", temp_branch_name, ctx.parent_of_strata])

    schedule, sources = _calculate_percolation_schedule(ctx)
    summary = SeepageSummary()

    try:
        for i, commit_hash in enumerate(ctx.strata):
            _process_single_stratum(
                ctx, summary, commit_hash, i + 1, schedule, sources
            )

        _finalize_seepage(ctx, summary, temp_branch_name)

    except KeyboardInterrupt:
        run_git(["reset", "--hard", "HEAD"], check=False)
        run_git(["checkout", "-f", ctx.current_branch], check=False)
        sys.exit(1)
    except Exception as e:
        print(f"\nFATAL ERROR DURING SEEPAGE: {e}")
        run_git(["reset", "--hard", "HEAD"], check=False)
        run_git(["checkout", "-f", ctx.current_branch], check=False)
        sys.exit(1)


def main():
    """Main entry point."""
    parser = argparse.ArgumentParser(
        description="GitSeep: Geological History Percolation"
    )
    # go/keep-sorted start
    parser.add_argument(
        "--all-files",
        action="store_true",
        help="List all files in bedrock commits",
    )
    parser.add_argument(
        "--auto-approve",
        action="store_true",
        help="Skip per-stratum confirmations",
    )
    parser.add_argument(
        "--base",
        help="The earliest commit to consider for the refactor range",
    )
    parser.add_argument("--branch", help="Target branch name")
    parser.add_argument(
        "--compress",
        action="store_true",
        help="Allow multiple historical versions to be squashed",
    )
    parser.add_argument("--rules", help="Path to YAML seepage rules file")
    parser.add_argument(
        "--stage-only",
        action="store_true",
        help="Leave result on temporary branch for inspection",
    )
    # go/keep-sorted end
    args = parser.parse_args()

    try:
        current_branch = (
            run_git(["rev-parse", "--abbrev-ref", "HEAD"]).stdout.strip()
        )
    except GitError:
        print("Error: Could not determine current branch.")
        sys.exit(1)

    target_branch = args.branch or current_branch

    # Load rules
    rules_path = args.rules
    if not rules_path:
        local_rules = os.path.join(os.getcwd(), RULES_FILENAME)
        script_rules = os.path.join(os.path.dirname(__file__), RULES_FILENAME)
        if os.path.exists(local_rules):
            rules_path = local_rules
        elif os.path.exists(script_rules):
            rules_path = script_rules

    if not rules_path or not os.path.exists(rules_path):
        msg = f"Error: Seepage rules not found. Please create {RULES_FILENAME}."
        print(msg)
        sys.exit(1)

    with open(rules_path, "r", encoding="utf-8") as f:
        try:
            rules_dict = yaml.safe_load(f)
        except yaml.YAMLError as e:
            print(f"Error parsing YAML: {e}")
            sys.exit(1)

    if not isinstance(rules_dict, dict):
        print("Error: Rules must be a mapping of strata dates to path lists.")
        sys.exit(1)

    seep_history(
        rules_dict,
        target_branch=target_branch,
        all_files=args.all_files,
        auto_approve=args.auto_approve,
        compress=args.compress,
        stage_only=args.stage_only,
        base_commit=args.base,
    )


if __name__ == "__main__":
    main()

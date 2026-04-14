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

"""Interactive Source Code Versioning History Refactoring tool.

This script consolidates path-based changes into earlier base commits while
preserving the exact commit count and messages of the original branch.
It provides a step-by-step interactive review process for safety.
"""

import argparse
import os
import subprocess
import sys
from typing import Dict, List, Optional

import yaml

# Default configuration names
RULES_FILENAME = 'refactor_rules.yaml'
EXAMPLE_FILENAME = 'refactor_rules.yaml.example'


class GitError(Exception):
    """Raised when a git command fails."""


def run_git(
    args: List[str],
    check: bool = True,
    env: Optional[Dict[str, str]] = None
) -> subprocess.CompletedProcess:
    """Runs a git command and returns the result."""
    try:
        return subprocess.run(
            ['git'] + args,
            capture_output=True,
            text=True,
            check=check,
            env=env or os.environ
        )
    except subprocess.CalledProcessError as e:
        error_msg = f"Error executing git {' '.join(args)}: {e.stderr}"
        if check:
            print(error_msg, file=sys.stderr)
            raise GitError(error_msg) from e
        return subprocess.CompletedProcess(e.args, e.returncode, e.stdout, e.stderr)


def get_commit_sequence(start_hash: str) -> List[str]:
    """Returns a list of commit hashes from start_hash to HEAD in chronological order."""
    result = run_git(['log', '--format=%H', f'{start_hash}^..HEAD'])
    return result.stdout.strip().split('\n')[::-1]


def resolve_commit_by_date(date_str: str) -> str:
    """Finds a unique commit hash matching the given author date string."""
    result = run_git([
        'log', 
        '--format=%H', 
        '--author-date-order', 
        f'--after={date_str}', 
        f'--before={date_str}',
        '--all'
    ])
    hashes = [h.strip() for h in result.stdout.strip().split('\n') if h.strip()]

    if not hashes:
        raise GitError(f"No commit found matching date: {date_str}")
    if len(hashes) > 1:
        raise GitError(f"Multiple commits found matching date {date_str}: {hashes}")
    
    return hashes[0]


def user_confirm(prompt: str) -> bool:
    """Prompts the user for Y/n confirmation."""
    while True:
        try:
            choice = input(f"{prompt} [Y/n]: ").strip().lower()
            if not choice or choice == 'y':
                return True
            if choice == 'n':
                return False
        except EOFError:
            return False


def refactor_history(rules_dict: Dict[str, List[str]], target_branch: str):
    """Interactively refactors history based on path rules.

    Args:
        rules_dict: Mapping of dates to list of owned paths.
        target_branch: The name of the branch to update with the new history.
    """
    # 1. Validation
    status = run_git(['status', '--porcelain'])
    if status.stdout.strip():
        print("Error: Working directory is not clean. Please commit or stash changes.")
        sys.exit(1)

    original_head = run_git(['rev-parse', 'HEAD']).stdout.strip()
    current_branch = run_git(['rev-parse', '--abbrev-ref', 'HEAD']).stdout.strip()

    # Resolve date-based rules to hashes
    resolved_rules = {}
    print("Resolving base commits by date...")
    try:
        for date_key, paths in rules_dict.items():
            commit_hash = resolve_commit_by_date(str(date_key))
            resolved_rules[commit_hash] = paths
            print(f"  {date_key} -> {commit_hash[:8]}")
    except GitError as e:
        print(f"Error: {e}")
        sys.exit(1)

    # Determine the start of our refactor range
    sorted_hashes = sorted(
        resolved_rules.keys(),
        key=lambda x: int(run_git(['log', '-1', '--format=%at', x]).stdout.strip())
    )
    earliest_hash = sorted_hashes[0]
    parent_of_refactor = run_git(['rev-parse', f'{earliest_hash}^']).stdout.strip()

    commits = get_commit_sequence(earliest_hash)

    print("\n--- PRE-FLIGHT BRIEFING ---")
    print(f"Target branch: {target_branch}")
    print(f"Commit range: {earliest_hash[:8]} .. HEAD ({len(commits)} commits)")
    print("Path ownership (Final state will be projected into these bases):")
    for h, paths in resolved_rules.items():
        print(f"  {h[:8]}: {', '.join(paths)}")

    if not user_confirm("\nInitialize Synthetic History Reconstruction?"):
        print("Aborted.")
        return

    # 2. Reconstruction
    temp_branch_name = "temp-history-cleanup"
    run_git(['branch', '-D', temp_branch_name], check=False)
    run_git(['checkout', '-b', temp_branch_name, parent_of_refactor])

    try:
        path_to_owner = {}
        for h, paths in resolved_rules.items():
            for p in paths:
                path_to_owner[p] = h

        for i, commit_hash in enumerate(commits):
            print(f"\n[{i+1}/{len(commits)}] Processing: {commit_hash[:8]}")
            msg = run_git(['log', '-1', '--format=%B', commit_hash]).stdout.strip()
            print(f"Message: {msg.splitlines()[0]}")

            run_git(['cherry-pick', '-n', commit_hash])

            for owned_path, owner_base in path_to_owner.items():
                if commit_hash == owner_base:
                    print(f"  [OWNER] Projecting final state for '{owned_path}'")
                    run_git(['checkout', original_head, '--', owned_path])
                else:
                    print(f"  [CLEAN] Nullifying intermediate changes for '{owned_path}'")
                    run_git(['checkout', 'HEAD', '--', owned_path])

            if not user_confirm(f"Commit {commit_hash[:8]}?"):
                print("Aborted by user.")
                run_git(['checkout', current_branch])
                return

            author_info = run_git(
                ['log', '-1', '--format=%an%n%ae%n%ad%n%cn%n%ce%n%cd', commit_hash]
            ).stdout.splitlines()

            custom_env = os.environ.copy()
            custom_env.update({
                'GIT_AUTHOR_NAME': author_info[0],
                'GIT_AUTHOR_EMAIL': author_info[1],
                'GIT_AUTHOR_DATE': author_info[2],
                'GIT_COMMITTER_NAME': author_info[3],
                'GIT_COMMITTER_EMAIL': author_info[4],
                'GIT_COMMITTER_DATE': author_info[5],
            })

            run_git(['commit', '--allow-empty', '-C', commit_hash], env=custom_env)

        # 3. Final Parity Check
        print("\n--- RECONSTRUCTION COMPLETE ---")
        print("Verifying file tree parity with original HEAD...")
        diff = run_git(['diff', original_head], check=False)

        if diff.stdout.strip():
            print("WARNING: Final file tree differs from original HEAD.")
            if user_confirm("Show diff?"):
                print(diff.stdout)
        else:
            print("SUCCESS: Parity check passed. File trees are identical.")

        # 4. Finalization
        if not user_confirm(f"\nFinalize: Update branch '{target_branch}' with this new history?"):
            print(f"Finalization skipped. Refactored history remains on '{temp_branch_name}'.")
            run_git(['checkout', current_branch])
            return

        # If target_branch exists, reset it. Otherwise create it.
        run_git(['checkout', '-B', target_branch, temp_branch_name])
        run_git(['branch', '-D', temp_branch_name])
        
        print(f"\nBranch '{target_branch}' has been updated.")
        if target_branch == current_branch:
            print("You can now run: git push --force-with-lease")
        else:
            print(f"You can now switch to '{target_branch}' and push.")

    except Exception as e:
        print(f"\nFATAL ERROR: {e}")
        run_git(['checkout', current_branch])
        sys.exit(1)


def main():
    """Main entry point."""
    parser = argparse.ArgumentParser(description="Source Code Versioning History Refactoring")
    parser.add_argument("--branch", help="Target branch name (defaults to current branch)")
    parser.add_argument("--rules", help="Path to YAML rules file")
    args = parser.parse_args()

    # Determine current branch for default
    try:
        current_branch = run_git(['rev-parse', '--abbrev-ref', 'HEAD']).stdout.strip()
    except GitError:
        print("Error: Could not determine current branch.")
        sys.exit(1)
    
    target_branch = args.branch or current_branch

    # Search for rules: 1. Arg, 2. Local dir, 3. Script dir
    rules_path = args.rules
    if not rules_path:
        local_rules = os.path.join(os.getcwd(), RULES_FILENAME)
        script_rules = os.path.join(os.path.dirname(__file__), RULES_FILENAME)
        if os.path.exists(local_rules):
            rules_path = local_rules
        elif os.path.exists(script_rules):
            rules_path = script_rules

    if not rules_path or not os.path.exists(rules_path):
        print(f"Error: Rules file not found. Please create {RULES_FILENAME}.")
        print(f"An example is available at: {os.path.join(os.path.dirname(__file__), EXAMPLE_FILENAME)}")
        sys.exit(1)

    with open(rules_path, 'r', encoding='utf-8') as f:
        try:
            rules_dict = yaml.safe_load(f)
        except yaml.YAMLError as e:
            print(f"Error parsing YAML: {e}")
            sys.exit(1)

    refactor_history(rules_dict, target_branch=target_branch)


if __name__ == '__main__':
    main()

// Copyright 2026 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package git provides helpers for interacting with the Git repository using go-git.
package git

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

// StatusOptions defines configuration for checking repository status.
type StatusOptions struct {
	// UseGoGit forces the use of the experimental pure-Go status implementation.
	// By default, GitSeep uses the more robust system git binary.
	UseGoGit bool
}

// GetStatusSummary checks the worktree and returns whether there are tracked changes and/or untracked files.
func GetStatusSummary(repo *git.Repository, opts StatusOptions) (hasChanges bool, hasUntracked bool, err error) {
	if opts.UseGoGit {
		worktree, err := repo.Worktree()
		if err != nil {
			return false, false, err
		}
		status, err := worktree.Status()
		if err != nil {
			return false, false, err
		}

		for _, s := range status {
			if s.Worktree == git.Untracked || s.Staging == git.Untracked {
				hasUntracked = true
			} else if s.Worktree != git.Unmodified || s.Staging != git.Unmodified {
				hasChanges = true
			}
		}
		return hasChanges, hasUntracked, nil
	}

	// Default: use robust system git status
	wt, err := repo.Worktree()
	if err != nil {
		return false, false, err
	}
	root := wt.Filesystem.Root()

	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		return false, false, err
	}

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if len(line) < 3 {
			continue
		}
		status := line[:2]
		if status == "??" {
			hasUntracked = true
		} else {
			hasChanges = true
		}
	}

	return hasChanges, hasUntracked, nil
}

// IsDirty returns true if there are any uncommitted changes (tracked or untracked).
func IsDirty(repo *git.Repository, opts StatusOptions) (bool, error) {
	hasChanges, hasUntracked, err := GetStatusSummary(repo, opts)
	return hasChanges || hasUntracked, err
}

// Amend performs a "git commit --amend --no-edit" to fold already staged changes into the last commit.
// It does NOT perform an implicit "git add", ensuring that only explicitly staged modifications are included.
func Amend(repo *git.Repository) error {
	wt, err := repo.Worktree()
	if err != nil {
		return err
	}
	root := wt.Filesystem.Root()

	// Try system git first for performance (near-instant on large repos)
	if _, err := exec.LookPath("git"); err == nil {
		cmdCommit := exec.Command("git", "commit", "--amend", "--no-edit", "--no-verify")
		cmdCommit.Dir = root
		if err := cmdCommit.Run(); err == nil {
			return nil
		}
	}

	// Fallback to go-git if system git is missing or fails
	head, err := repo.Head()
	if err != nil {
		return err
	}

	lastCommit, err := repo.CommitObject(head.Hash())
	if err != nil {
		return err
	}

	_, err = wt.Commit(lastCommit.Message, &git.CommitOptions{
		Amend:  true,
		Author: &lastCommit.Author,
	})

	return err
}

// VirtualAmend simulates a "git commit --amend" by creating a dangling commit object
// representing the current index, without moving any branch references.
func VirtualAmend(repo *git.Repository) (plumbing.Hash, error) {
	wt, err := repo.Worktree()
	if err != nil {
		return plumbing.ZeroHash, err
	}
	root := wt.Filesystem.Root()

	// 1. Capture index as a tree
	cmdTree := exec.Command("git", "write-tree")
	cmdTree.Dir = root
	treeHashBytes, err := cmdTree.Output()
	if err != nil {
		return plumbing.ZeroHash, fmt.Errorf("failed to write-tree: %w", err)
	}
	treeHash := strings.TrimSpace(string(treeHashBytes))

	// 2. Identify parents and message of HEAD
	// format: %P (parents) \n %B (full message)
	cmdLog := exec.Command("git", "log", "-n", "1", "--format=%P%n%B", "HEAD")
	cmdLog.Dir = root
	logOutput, err := cmdLog.Output()
	if err != nil {
		return plumbing.ZeroHash, fmt.Errorf("failed to get HEAD metadata: %w", err)
	}

	lines := strings.SplitN(string(logOutput), "\n", 2)
	if len(lines) < 2 {
		return plumbing.ZeroHash, fmt.Errorf("unexpected log output format")
	}

	parentLine := strings.TrimSpace(lines[0])
	message := lines[1]

	// 3. Create dangling commit object
	// We use the parents from the HEAD commit to preserve the history structure
	args := []string{"commit-tree", treeHash}
	if parentLine != "" {
		parents := strings.Split(parentLine, " ")
		for _, p := range parents {
			args = append(args, "-p", p)
		}
	}
	args = append(args, "-m", message)

	cmdCommit := exec.Command("git", args...)
	cmdCommit.Dir = root
	commitHashBytes, err := cmdCommit.Output()
	if err != nil {
		return plumbing.ZeroHash, fmt.Errorf("failed to create virtual commit: %w", err)
	}

	return plumbing.NewHash(strings.TrimSpace(string(commitHashBytes))), nil
}

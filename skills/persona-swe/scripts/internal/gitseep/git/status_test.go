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

// Package git_test provides integration tests for the Git interaction layer.
package git_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	gogit "github.com/go-git/go-git/v5"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/git"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/testutil"
)

func TestSystemGitStatus(t *testing.T) {
	// Create a temporary disk-backed repository to test system git integration
	tmp, err := os.MkdirTemp("", "gitseep-system-git-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmp)

	// Initialize real git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = tmp
	if err := cmd.Run(); err != nil {
		t.Fatalf("git init failed: %v", err)
	}

	// Set local config to avoid failure if global config is missing
	_ = exec.Command("git", "-C", tmp, "config", "user.email", "test@example.com").Run()
	_ = exec.Command("git", "-C", tmp, "config", "user.name", "Test User").Run()

	repo, err := gogit.PlainOpen(tmp)
	if err != nil {
		t.Fatalf("Failed to open repo: %v", err)
	}

	t.Run("Clean State", func(t *testing.T) {
		hasChanges, hasUntracked, err := git.GetStatusSummary(repo, git.StatusOptions{UseGoGit: false})
		if err != nil {
			t.Fatalf("GetStatusSummary failed: %v", err)
		}
		if hasChanges || hasUntracked {
			t.Errorf("Expected clean status, got changes=%v, untracked=%v", hasChanges, hasUntracked)
		}
	})

	t.Run("Detected Changes", func(t *testing.T) {
		// Create a file and commit it
		path := filepath.Join(tmp, "tracked.txt")
		_ = os.WriteFile(path, []byte("initial"), 0644)
		_ = exec.Command("git", "-C", tmp, "add", "tracked.txt").Run()
		_ = exec.Command("git", "-C", tmp, "commit", "-m", "msg").Run()

		// Modify tracked file
		_ = os.WriteFile(path, []byte("modified"), 0644)

		// Create untracked file
		_ = os.WriteFile(filepath.Join(tmp, "untracked.txt"), []byte("new"), 0644)

		hasChanges, hasUntracked, err := git.GetStatusSummary(repo, git.StatusOptions{UseGoGit: false})
		if err != nil {
			t.Fatalf("GetStatusSummary failed: %v", err)
		}
		if !hasChanges {
			t.Errorf("System git failed to detect tracked changes")
		}
		if !hasUntracked {
			t.Errorf("System git failed to detect untracked files")
		}
	})

	t.Run("IsDirty", func(t *testing.T) {
		// Clean up from previous subtest
		_ = exec.Command("git", "-C", tmp, "reset", "--hard").Run()
		_ = exec.Command("git", "-C", tmp, "clean", "-fd").Run()

		// Clean repo
		dirty, err := git.IsDirty(repo, git.StatusOptions{UseGoGit: false})
		if err != nil {
			t.Fatalf("IsDirty failed: %v", err)
		}
		if dirty {
			t.Errorf("Expected repo to be clean")
		}

		// Make it dirty
		path := filepath.Join(tmp, "dirty.txt")
		_ = os.WriteFile(path, []byte("dirty"), 0644)
		dirty, _ = git.IsDirty(repo, git.StatusOptions{UseGoGit: false})
		if !dirty {
			t.Errorf("Expected repo to be dirty after adding file")
		}
	})

	t.Run("System Amend", func(t *testing.T) {
		// Ensure clean state
		_ = exec.Command("git", "-C", tmp, "reset", "--hard").Run()
		_ = exec.Command("git", "-C", tmp, "clean", "-fd").Run()

		// Create and commit a file
		path := filepath.Join(tmp, "system_amend.txt")
		_ = os.WriteFile(path, []byte("initial"), 0644)
		_ = exec.Command("git", "-C", tmp, "add", "system_amend.txt").Run()
		_ = exec.Command("git", "-C", tmp, "commit", "-m", "system msg").Run()
		h1, _ := repo.Head()

		// Modify file and STAGE it (Amend is now staged-only)
		_ = os.WriteFile(path, []byte("modified"), 0644)
		_ = exec.Command("git", "-C", tmp, "add", "system_amend.txt").Run()

		// Amend using the optimized system git path
		if err := git.Amend(repo); err != nil {
			t.Fatalf("System Amend failed: %v", err)
		}

		h2, _ := repo.Head()
		if h1.Hash() == h2.Hash() {
			t.Errorf("Hash did not change after system amend")
		}

		commit, _ := repo.CommitObject(h2.Hash())
		if commit.Message != "system msg\n" && commit.Message != "system msg" {
			t.Errorf("Expected message to be preserved, got %q", commit.Message)
		}

		tree, _ := commit.Tree()
		file, _ := tree.File("system_amend.txt")
		content, _ := file.Contents()
		if content != "modified" {
			t.Errorf("Expected content 'modified', got %q", content)
		}
	})
}

func TestGitStatusSuite(t *testing.T) {
	repo, wt := testutil.SetupMemRepo(t)

	t.Run("Get Status Summary", func(t *testing.T) {
		d := time.Now()
		_, _ = testutil.CommitFile(t, testutil.CommitParams{
			Repo: repo, Worktree: wt, Path: "base.txt", Content: "base", Msg: "msg", AuthorDate: d,
		})

		// 1. Tracked change
		f, _ := wt.Filesystem.Create("base.txt")
		_, _ = f.Write([]byte("modified"))
		f.Close()

		// 2. Untracked file
		f2, _ := wt.Filesystem.Create("untracked.txt")
		_, _ = f2.Write([]byte("new"))
		f2.Close()

		// Use experimental=true in tests because system git won't find the memory-backed repository.
		hasChanges, hasUntracked, err := git.GetStatusSummary(repo, git.StatusOptions{UseGoGit: true})
		if err != nil {
			t.Fatalf("GetStatusSummary failed: %v", err)
		}

		if !hasChanges {
			t.Errorf("Expected tracked changes to be detected")
		}
		if !hasUntracked {
			t.Errorf("Expected untracked files to be detected")
		}
	})

	t.Run("Amend", func(t *testing.T) {
		repo, wt := testutil.SetupMemRepo(t)
		d := time.Now()
		h1, _ := testutil.CommitFile(t, testutil.CommitParams{
			Repo: repo, Worktree: wt, Path: "test.txt", Content: "initial", Msg: "msg1", AuthorDate: d,
		})

		// Modify file and STAGE it
		f, _ := wt.Filesystem.Create("test.txt")
		_, _ = f.Write([]byte("modified"))
		f.Close()
		_, _ = wt.Add("test.txt")

		err := git.Amend(repo)
		if err != nil {
			t.Fatalf("Amend failed: %v", err)
		}

		head, _ := repo.Head()
		if head.Hash().String() == h1 {
			t.Errorf("Expected hash to change after amend")
		}

		commit, _ := repo.CommitObject(head.Hash())
		if commit.Message != "msg1" {
			t.Errorf("Expected message to be preserved, got %s", commit.Message)
		}

		tree, _ := commit.Tree()
		file, _ := tree.File("test.txt")
		content, _ := file.Contents()
		if content != "modified" {
			t.Errorf("Expected content to be 'modified', got %s", content)
		}
	})
}

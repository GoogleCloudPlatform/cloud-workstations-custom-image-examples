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

// Package testutil provides common testing utilities and mocks for GitSeep.
package testutil

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"

	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/i18n"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/locales"
)

// InitI18n initializes the localization system for tests.
func InitI18n() {
	if err := i18n.Init("", "en", locales.Content); err != nil {
		panic(err)
	}
}

// SetupMemRepo creates a fast, in-memory Git repository for testing.
func SetupMemRepo(t *testing.T) (*git.Repository, *git.Worktree) {
	t.Helper()
	storer := memory.NewStorage()
	fs := memfs.New()
	repo, _ := git.Init(storer, fs)
	wt, _ := repo.Worktree()
	return repo, wt
}

// CommitParams defines the parameters for creating a test commit.
type CommitParams struct {
	Repo       *git.Repository
	Worktree   *git.Worktree
	Path       string
	Content    string
	Msg        string
	AuthorDate time.Time
}

// CommitFile creates or updates a file and commits it to the repo.
func CommitFile(t *testing.T, p CommitParams) (string, error) {
	t.Helper()
	if p.Worktree == nil {
		wt, _ := p.Repo.Worktree()
		p.Worktree = wt
	}

	f, err := p.Worktree.Filesystem.Create(p.Path)
	if err != nil {
		return "", err
	}
	if _, err := f.Write([]byte(p.Content)); err != nil {
		f.Close()
		return "", err
	}
	if err := f.Close(); err != nil {
		return "", err
	}

	_, err = p.Worktree.Add(p.Path)
	if err != nil {
		return "", err
	}

	hash, err := p.Worktree.Commit(p.Msg, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test User",
			Email: "test@example.com",
			When:  p.AuthorDate,
		},
	})
	if err != nil {
		if err.Error() == "cannot create empty commit: clean working tree" {
			// If nothing to commit (e.g. file already gone), just return empty
			return "", nil
		}
		return "", err
	}

	return hash.String(), nil
}

// RemoveFile deletes a file and commits the removal.
func RemoveFile(t *testing.T, p CommitParams) (string, error) {
	t.Helper()
	if p.Worktree == nil {
		wt, _ := p.Repo.Worktree()
		p.Worktree = wt
	}

	// Ensure the index is up to date before removing
	if _, err := p.Worktree.Add(p.Path); err != nil {
		return "", err
	}

	// In memfs/go-git, we need to ensure the file is removed from the worktree index
	if _, err := p.Worktree.Remove(p.Path); err != nil {
		return "", err
	}

	hash, err := p.Worktree.Commit(p.Msg, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test User",
			Email: "test@example.com",
			When:  p.AuthorDate,
		},
	})
	if err != nil {
		if err.Error() == "cannot create empty commit: clean working tree" {
			// If nothing to commit (e.g. file already gone), just return empty
			return "", nil
		}
		return "", err
	}

	return hash.String(), nil
}

// SetupTestRepoSystemGit creates a real Git repository on disk using the system git binary.
func SetupTestRepoSystemGit(t *testing.T) (string, string, string, func()) {
	t.Helper()
	dir := t.TempDir()
	run := func(args ...string) string {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v failed: %v\nOutput: %s", args, err, string(out))
		}
		return strings.TrimSpace(string(out))
	}

	run("init")
	run("config", "user.email", "t@t.com")
	run("config", "user.name", "T")
	_ = os.WriteFile(filepath.Join(dir, "a.txt"), []byte("v1"), 0644)
	_ = os.WriteFile(filepath.Join(dir, ".pre-commit-config.yaml"), []byte("repos: []"), 0644)
	run("add", ".")
	run("commit", "-m", "initial")
	h1 := run("rev-parse", "HEAD")

	_ = os.WriteFile(filepath.Join(dir, "a.txt"), []byte("v2"), 0644)
	run("add", "a.txt")
	run("commit", "-m", "second")
	h2 := run("rev-parse", "HEAD")

	cleanup := func() { os.RemoveAll(dir) }
	return dir, h1, h2, cleanup
}

// SetupMockPreCommit sets up a mock pre-commit binary in the PATH.
func SetupMockPreCommit(t *testing.T, exitCode int, countFile string) {
	t.Helper()
	tmpDir := t.TempDir()
	mockPreCommitPath := filepath.Join(tmpDir, "pre-commit")

	src := fmt.Sprintf(`package main
import (
	"os"
)
func main() {
	cf := "%s"
	if cf != "" {
		f, _ := os.OpenFile(cf, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if f != nil {
			f.WriteString("ran\n")
			f.Close()
		}
	}
	os.Exit(%d)
}
`, strings.ReplaceAll(countFile, "\\", "\\\\"), exitCode)

	mockSrc := filepath.Join(tmpDir, "mock_precommit.go")
	if err := os.WriteFile(mockSrc, []byte(src), 0644); err != nil {
		t.Fatalf("failed to write mock src: %v", err)
	}
	cmd := exec.Command("go", "build", "-o", mockPreCommitPath, mockSrc)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build mock: %v\n%s", err, string(out))
	}

	oldPath := os.Getenv("PATH")
	t.Setenv("PATH", tmpDir+":"+oldPath)
}

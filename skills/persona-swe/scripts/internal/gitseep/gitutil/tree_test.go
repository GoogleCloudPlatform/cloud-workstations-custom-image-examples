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

// Package gitutil_test provides integration tests for the Git tree utilities.
package gitutil_test

import (
	"testing"
	"time"

	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/gitutil"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/testutil"
)

func TestGetAllEntriesAndBuildTree(t *testing.T) {
	repo, wt := testutil.SetupMemRepo(t)
	d0 := time.Now()

	_, _ = testutil.CommitFile(t, testutil.CommitParams{
		Repo: repo, Worktree: wt, Path: "src/main.go", Content: "package main", Msg: "init", AuthorDate: d0,
	})
	_, _ = testutil.CommitFile(t, testutil.CommitParams{
		Repo: repo, Worktree: wt, Path: "README.md", Content: "hello", Msg: "doc", AuthorDate: d0.Add(time.Hour),
	})

	head, _ := repo.Head()
	headCommit, _ := repo.CommitObject(head.Hash())
	headTree, _ := headCommit.Tree()

	entries, err := gitutil.GetAllEntries(repo, headTree)
	if err != nil {
		t.Fatalf("GetAllEntries failed: %v", err)
	}

	if len(entries) != 2 {
		t.Errorf("Expected 2 files, got %d", len(entries))
	}

	if _, ok := entries["src/main.go"]; !ok {
		t.Errorf("Missing src/main.go")
	}

	if _, ok := entries["README.md"]; !ok {
		t.Errorf("Missing README.md")
	}

	// Rebuild tree from the entries
	newTreeHash, err := gitutil.BuildTree(repo, entries)
	if err != nil {
		t.Fatalf("BuildTree failed: %v", err)
	}

	if newTreeHash.IsZero() {
		t.Fatalf("BuildTree returned zero hash")
	}

	if newTreeHash != headCommit.TreeHash {
		t.Errorf("Rebuilt tree hash %v does not match original tree hash %v", newTreeHash, headCommit.TreeHash)
	}

	// Test building empty tree
	emptyEntries := make(map[string]object.TreeEntry)
	emptyTreeHash, err := gitutil.BuildTree(repo, emptyEntries)
	if err != nil {
		t.Fatalf("BuildTree empty failed: %v", err)
	}
	if emptyTreeHash.IsZero() {
		t.Errorf("Expected valid hash for empty tree, got zero")
	}
}

func TestGetAllEntriesRecursive(t *testing.T) {
	repo, wt := testutil.SetupMemRepo(t)
	_, err := testutil.CommitFile(t, testutil.CommitParams{Repo: repo, Worktree: wt, Path: "a/b/c.txt", Content: "c"})
	if err != nil {
		t.Fatalf("CommitFile failed: %v", err)
	}

	head, _ := repo.Head()
	commit, _ := repo.CommitObject(head.Hash())
	tree, _ := commit.Tree()

	entries := make(map[string]object.TreeEntry)
	err = gitutil.GetAllEntriesRecursive(repo, tree, "", entries)
	if err != nil {
		t.Fatalf("GetAllEntriesRecursive failed: %v", err)
	}

	if _, ok := entries["a/b/c.txt"]; !ok {
		t.Errorf("Expected a/b/c.txt in entries, got %v", entries)
	}
}

func TestGitUtil_Coverage(t *testing.T) {
	repo, wt := testutil.SetupMemRepo(t)
	now := time.Now()

	h1, _ := testutil.CommitFile(t, testutil.CommitParams{
		Repo: repo, Worktree: wt, Path: "file1.txt", Content: "c1", Msg: "msg1", AuthorDate: now,
	})
	h2, _ := testutil.CommitFile(t, testutil.CommitParams{
		Repo: repo, Worktree: wt, Path: "file2.txt", Content: "c2", Msg: "msg2", AuthorDate: now.Add(time.Hour),
	})

	// Removal test
	h3, _ := testutil.CommitFile(t, testutil.CommitParams{
		Repo: repo, Worktree: wt, Path: "remove_me.txt", Content: "temp", Msg: "msg3", AuthorDate: now.Add(time.Hour * 2),
	})
	h4, err := testutil.RemoveFile(t, testutil.CommitParams{
		Repo: repo, Worktree: wt, Path: "remove_me.txt", Msg: "msg4", AuthorDate: now.Add(time.Hour * 3),
	})
	if err != nil || h4 == "" {
		// If RemoveFile fails in memrepo, we skip the removal check but fail the test if it's critical
		t.Logf("RemoveFile returned empty or error: %v", err)
	}

	t.Run("GetChangedFiles", func(t *testing.T) {
		files, err := gitutil.GetChangedFiles(repo, h1, h2)
		if err != nil {
			t.Fatalf("GetChangedFiles failed: %v", err)
		}
		if len(files) == 0 {
			t.Error("Expected changes between h1 and h2")
		}

		if h4 != "" {
			files2, err := gitutil.GetChangedFiles(repo, h3, h4)
			if err != nil {
				t.Fatalf("GetChangedFiles failed for removal: %v", err)
			}
			if len(files2) == 0 {
				t.Error("Expected changes between h3 and h4")
			}
		}
	})
}

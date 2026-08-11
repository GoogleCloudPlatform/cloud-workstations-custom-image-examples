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

package git

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/models"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/testutil"
)

func TestHybridGitService_RealRepo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping real repo test in short mode")
	}

	root, h1, h2, cleanup := testutil.SetupTestRepoSystemGit(t)
	defer cleanup()

	svc, err := NewGitService(root, false)
	if err != nil {
		t.Fatalf("Failed to create GitService: %v", err)
	}

	t.Run("IsClean and Amend", func(t *testing.T) {
		isClean, err := svc.IsClean()
		if err != nil {
			t.Fatalf("IsClean failed: %v", err)
		}
		if !isClean {
			t.Error("Repo should be clean initially")
		}

		// Create a staged change for Amend
		_ = os.WriteFile(filepath.Join(root, "staged.txt"), []byte("staged"), 0644)

		// Use real git to add
		repo := svc.GetRepo()
		wt, _ := repo.Worktree()
		_, _ = wt.Add("staged.txt")

		err = svc.Amend()
		if err != nil {
			t.Errorf("Amend failed: %v", err)
		}
	})

	t.Run("ResetMixed", func(t *testing.T) {
		err := svc.ResetMixed(h1)
		if err != nil {
			t.Errorf("ResetMixed failed: %v", err)
		}
	})

	t.Run("UpdateBranch", func(t *testing.T) {
		err := svc.UpdateBranch("test-branch", plumbing.NewHash(h2))
		if err != nil {
			t.Errorf("UpdateBranch failed: %v", err)
		}
	})

	t.Run("GetChangedFiles", func(t *testing.T) {
		files, err := svc.GetChangedFiles(h1, h2)
		if err != nil {
			t.Fatalf("GetChangedFiles failed: %v", err)
		}
		if len(files) == 0 {
			t.Error("Expected changed files between h1 and h2")
		}
	})
}

func TestHybridGitService_SafeFinalize(t *testing.T) {
	repo, wt := testutil.SetupMemRepo(t)
	h1, _ := testutil.CommitFile(t, testutil.CommitParams{Repo: repo, Worktree: wt, Path: "f.txt", Content: "c"})

	svc := NewTestGitService(repo)

	t.Run("StageOnly", func(t *testing.T) {
		seepage := &models.SeepageContext{
			TargetBranch: "main",
			Options:      models.Options{StageOnly: true},
		}
		targetBranch, err := svc.SafeFinalize(seepage, h1, nil)
		if err != nil {
			t.Fatalf("SafeFinalize failed: %v", err)
		}

		// Verify dynamic staged reference
		ref, err := repo.Reference(plumbing.NewBranchReferenceName(targetBranch), true)
		if err != nil {
			t.Fatalf("%s ref not found: %v", targetBranch, err)
		}
		if ref.Hash().String() != h1 {
			t.Errorf("Expected hash %s, got %s", h1, ref.Hash())
		}
	})

	t.Run("Full Finalize with Worktree Sync", func(t *testing.T) {
		seepage := &models.SeepageContext{
			TargetBranch: "main",
			Options:      models.Options{StageOnly: false, DryRun: false},
		}
		// Ensure we are on 'main' branch
		_ = wt.Checkout(&git.CheckoutOptions{Branch: plumbing.NewBranchReferenceName("main")})

		_, err := svc.SafeFinalize(seepage, h1, map[string]string{"feat1": h1})
		if err != nil {
			t.Fatalf("SafeFinalize failed: %v", err)
		}

		// Verify main branch updated
		ref, _ := repo.Reference("refs/heads/main", true)
		if ref.Hash().String() != h1 {
			t.Errorf("Expected main hash %s, got %s", h1, ref.Hash())
		}
	})

	t.Run("Detached HEAD", func(t *testing.T) {
		h2, _ := testutil.CommitFile(t, testutil.CommitParams{Repo: repo, Worktree: wt, Path: "f2.txt", Content: "c2"})
		seepage := &models.SeepageContext{
			TargetBranch: h1, // Detached on h1
			Options:      models.Options{StageOnly: false, DryRun: false},
		}

		// Simulate detached HEAD in the service's repo view
		_ = wt.Checkout(&git.CheckoutOptions{Hash: plumbing.NewHash(h1)})

		_, err := svc.SafeFinalize(seepage, h2, nil)
		if err != nil {
			t.Fatalf("SafeFinalize failed: %v", err)
		}

		// Verify no branch named after h1 was created
		_, err = repo.Reference(plumbing.NewBranchReferenceName(h1), true)
		if err == nil {
			t.Errorf("Should NOT have created a branch named after the hash %s", h1)
		}
	})
}

func TestGitService_Coverage(t *testing.T) {
	repo, wt := testutil.SetupMemRepo(t)
	now := time.Now()
	h1, _ := testutil.CommitFile(t, testutil.CommitParams{
		Repo: repo, Worktree: wt, Path: "f1", Content: "c1", Msg: "m1", AuthorDate: now,
	})

	svc := NewTestGitService(repo)

	t.Run("ResolveRevision", func(t *testing.T) {
		h, err := svc.ResolveRevision("HEAD")
		if err != nil {
			t.Errorf("ResolveRevision failed: %v", err)
		}
		if h.String() != h1 {
			t.Errorf("Expected %s, got %s", h1, h.String())
		}
	})

	t.Run("GetCurrentBranch", func(t *testing.T) {
		b, err := svc.GetCurrentBranch()
		if err != nil {
			t.Errorf("GetCurrentBranch failed: %v", err)
		}
		if b == "" || b == "HEAD" {
			t.Errorf("Expected a branch name, got %s", b)
		}
	})

	t.Run("RepoRoot", func(t *testing.T) {
		if svc.RepoRoot() == "" {
			t.Error("RepoRoot should not be empty")
		}
	})
}

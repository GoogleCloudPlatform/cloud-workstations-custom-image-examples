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

package engine

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-git/go-git/v5"
	gitseepGit "github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/git"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/models"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/testutil"
)

func TestValidateHistory_Skip(t *testing.T) {
	testutil.InitI18n()
	err := ValidateHistory(ValidationParams{Skip: true, RepoRoot: "some-repo", OriginalHead: "HEAD", LinearHashes: []string{"hash1"}, BranchHeads: map[string]string{"branch1": "hash2"}})
	if err != nil {
		t.Errorf("Expected nil error when skip is true, got %v", err)
	}
}

func TestValidateHistory_NoPreCommitConfig(t *testing.T) {
	testutil.InitI18n()
	tmpDir := t.TempDir()

	// Since there is no .pre-commit-config.yaml in HEAD, it should log a warning and return nil
	err := ValidateHistory(ValidationParams{Seepage: &models.SeepageContext{}, RepoRoot: tmpDir, OriginalHead: "HEAD", LinearHashes: []string{"hash1"}, BranchHeads: map[string]string{"branch1": "hash2"}})
	if err != nil {
		t.Errorf("Expected nil error when pre-commit-config.yaml is missing in HEAD, got %v", err)
	}
}

func TestValidateHistory_SuccessAndCleanup(t *testing.T) {
	testutil.InitI18n()
	repoDir, h1, h2, cleanupRepo := testutil.SetupTestRepoSystemGit(t)
	defer cleanupRepo()

	testutil.SetupMockPreCommit(t, 0, "")

	repo, _ := git.PlainOpen(repoDir)
	err := ValidateHistory(ValidationParams{Seepage: &models.SeepageContext{}, RepoRoot: repoDir, Repo: repo, OriginalHead: h2, LinearHashes: []string{h1}})
	if err != nil {
		t.Fatalf("ValidateHistory failed unexpectedly: %v", err)
	}

	out, err := exec.Command("git", "-C", repoDir, "worktree", "list").Output()
	if err != nil {
		t.Fatalf("Failed to list worktrees: %v", err)
	}
	if strings.Contains(string(out), "gitseep-validation") {
		t.Errorf("Expected validation worktree to be removed on success, but it was found: %s", string(out))
	}
}

func TestValidateHistory_FailurePreservesWorktree(t *testing.T) {
	testutil.InitI18n()
	repoDir, h1, h2, cleanupRepo := testutil.SetupTestRepoSystemGit(t)
	defer cleanupRepo()

	testutil.SetupMockPreCommit(t, 1, "") // Fails!

	repo, _ := git.PlainOpen(repoDir)
	err := ValidateHistory(ValidationParams{Seepage: &models.SeepageContext{}, RepoRoot: repoDir, Repo: repo, OriginalHead: h2, LinearHashes: []string{h1}})
	if err == nil {
		t.Fatalf("ValidateHistory expected to fail, but it succeeded")
	}

	out, err := exec.Command("git", "-C", repoDir, "worktree", "list").Output()
	if err != nil {
		t.Fatalf("Failed to list worktrees: %v", err)
	}
	if !strings.Contains(string(out), "gitseep-validation") {
		t.Errorf("Expected validation worktree to be PRESERVED on failure, but it was not found: %s", string(out))
	}
}

func TestValidateHistory_Optimization(t *testing.T) {
	testutil.InitI18n()
	repoDir, h1, h2, cleanupRepo := testutil.SetupTestRepoSystemGit(t)
	defer cleanupRepo()

	countFile := filepath.Join(repoDir, "count_opt.txt")
	testutil.SetupMockPreCommit(t, 0, countFile)

	repo, _ := git.PlainOpen(repoDir)
	// Mock OriginalToReconstructed where h1 matches itself (unchanged)
	seepageCtx := &models.SeepageContext{
		OriginalToReconstructed: map[string]string{h1: h1},
		HashToSubject:           map[string]string{h1: "initial", h2: "second"},
	}

	// Validate history: Tips (h2) to Bedrock (h1)
	// Since h1 is marked as unchanged in OriginalToReconstructed, it should be skipped.
	err := ValidateHistory(ValidationParams{Seepage: seepageCtx, RepoRoot: repoDir, Repo: repo, OriginalHead: h2, LinearHashes: []string{h1}})
	if err != nil {
		t.Fatalf("ValidateHistory failed unexpectedly: %v", err)
	}

	data, _ := os.ReadFile(countFile)
	count := strings.Count(string(data), "ran")
	// Expected 0 runs because h1 is unchanged and it's the only one in the list.
	// (Actually, h2 is our Tip, but it's not in the linearHashes list in this call)
	if count != 0 {
		t.Errorf("Expected 0 pre-commit runs due to optimization, but ran %d times", count)
	}
}

func TestValidateHistory_SuccessCoverage(t *testing.T) {
	testutil.InitI18n()
	repo, _ := testutil.SetupMemRepo(t)
	gitSvc := gitseepGit.NewTestGitService(repo)

	tmpRoot := t.TempDir()
	// Create mock pre-commit config
	_ = os.WriteFile(filepath.Join(tmpRoot, ".pre-commit-config.yaml"), []byte("repos: []"), 0644)

	events := make(chan models.PipelineEvent, 100)
	seepage := &models.SeepageContext{
		DateToBranch:  map[string]string{"2026-01-01": "feat1"},
		HashToDate:    map[string]string{"h1": "2026-01-01"},
		HashToSubject: map[string]string{"h1": "feat: test"},
	}

	p := ValidationParams{
		RepoRoot:     tmpRoot,
		Repo:         repo,
		GitSvc:       gitSvc,
		Seepage:      seepage,
		LinearHashes: []string{"h1"},
		BranchHeads:  map[string]string{"feat1": "h1"},
		Events:       events,
		Shell:        &mockShellService{},
	}

	err := ValidateHistory(p)
	if err != nil {
		t.Errorf("ValidateHistory failed: %v", err)
	}
}

func TestValidateHistory_MissingPreCommitBinary(t *testing.T) {
	testutil.InitI18n()

	// Create a temp directory and set it as PATH to simulate missing binary
	tmpPath := t.TempDir()
	oldPath := os.Getenv("PATH")
	t.Setenv("PATH", tmpPath)
	defer os.Setenv("PATH", oldPath)

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".pre-commit-config.yaml")
	_ = os.WriteFile(configPath, []byte("repos: []"), 0644)

	err := ValidateHistory(ValidationParams{
		RepoRoot:     tmpDir,
		OriginalHead: "HEAD",
	})

	if err != nil {
		t.Errorf("Expected nil error even if pre-commit is missing (it should skip with a message), got %v", err)
	}
}

func TestValidateSurface_NoChanges(t *testing.T) {
	testutil.InitI18n()
	repo, _ := testutil.SetupMemRepo(t)
	h0, _ := testutil.CommitFile(t, testutil.CommitParams{Repo: repo, Path: "f.txt", Content: "c"})

	// If no changes are detected on surface, it should return success
	_, err := ValidateSurface(ValidationParams{
		Repo:         repo,
		OriginalHead: h0,
	})
	if err != nil {
		t.Errorf("Expected success for no changes, got %v", err)
	}
}

func TestValidateHistory_ComplexScenarios(t *testing.T) {
	testutil.InitI18n()
	repoDir, h1, h2, cleanupRepo := testutil.SetupTestRepoSystemGit(t)
	defer cleanupRepo()
	testutil.SetupMockPreCommit(t, 0, "")
	repo, _ := git.PlainOpen(repoDir)

	events := make(chan models.PipelineEvent, 100)

	t.Run("Feature Branches and Already Validated", func(t *testing.T) {
		p := ValidationParams{
			Repo:         repo,
			RepoRoot:     repoDir,
			OriginalHead: h2,
			BranchHeads:  map[string]string{"feat/1": h1, "feat/2": h1},
			Seepage: &models.SeepageContext{
				DateToBranch: map[string]string{"2026-01-01": "feat/1", "2026-01-02": "feat/2"},
			},
			Events: events,
		}
		err := ValidateHistory(p)
		if err != nil {
			t.Fatalf("ValidateHistory failed: %v", err)
		}

		// Consume events and verify "already_validated" for feat/1 if it was already processed
		// (The order depends on dates, 2026-01-02 is newest so feat/2 comes first)
	})

	t.Run("SkipHash Optimization", func(t *testing.T) {
		p := ValidationParams{
			Repo:         repo,
			RepoRoot:     repoDir,
			OriginalHead: h2,
			LinearHashes: []string{h1},
			SkipHash:     h1,
			Seepage:      &models.SeepageContext{},
			Events:       events,
		}
		err := ValidateHistory(p)
		if err != nil {
			t.Fatalf("ValidateHistory failed: %v", err)
		}
	})
}

func TestValidateHistory_DAMP(t *testing.T) {
	testutil.InitI18n()

	// Create a real temp dir for the "repo root" so pre-commit config check passes
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".pre-commit-config.yaml")
	_ = os.WriteFile(configPath, []byte("repos: []"), 0644)

	repo, _ := testutil.SetupMemRepo(t)
	events := make(chan models.PipelineEvent, 100)
	mock := &mockShellService{}

	t.Run("History Validation Success", func(t *testing.T) {
		p := ValidationParams{
			Repo:         repo,
			RepoRoot:     tmpDir,
			OriginalHead: "head",
			Seepage: &models.SeepageContext{
				OriginalToReconstructed: map[string]string{
					"old1": "new1",
				},
				Strata: []string{"old1"},
			},
			Events: events,
			Shell:  mock,
		}

		err := ValidateHistory(p)
		if err != nil {
			t.Fatalf("ValidateHistory failed: %v", err)
		}

		// Verify expected events
		hasStart := false
		hasSuccess := false
		for len(events) > 0 {
			ev := <-events
			if ev.Phase == models.PhaseValidation && ev.Type == models.EventStart {
				hasStart = true
			}
			if ev.Phase == models.PhaseValidation && ev.Type == models.EventSuccess {
				hasSuccess = true
			}
		}

		if !hasStart {
			t.Error("Missing Validation Start event")
		}
		if !hasSuccess {
			t.Error("Missing Validation Success event")
		}
	})

	t.Run("History Validation Skipped", func(t *testing.T) {
		p := ValidationParams{
			Skip:   true,
			Events: events,
		}

		err := ValidateHistory(p)
		if err != nil {
			t.Fatalf("ValidateHistory failed when skipped: %v", err)
		}

		// First event should be Start
		ev := <-events
		if ev.Type != models.EventStart {
			t.Errorf("Expected start event, got %v", ev)
		}

		ev = <-events
		if ev.Phase != models.PhaseValidation || ev.Type != models.EventSkipped {
			t.Errorf("Expected skipped event, got %v", ev)
		}
	})
}

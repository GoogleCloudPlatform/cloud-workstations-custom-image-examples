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
	"testing"

	"github.com/go-git/go-git/v5"
	gitseepGit "github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/git"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/models"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/testutil"
)

func TestSynchronizeWorktreeIfActive_NonDestructive(t *testing.T) {
	testutil.InitI18n()
	// Use System Git for realistic reset --mixed test
	repoDir, _, _, cleanup := testutil.SetupTestRepoSystemGit(t)
	defer cleanup()

	repo, _ := git.PlainOpen(repoDir)
	wt, _ := repo.Worktree()

	// 1. Create a "reconstructed" commit hash using system git
	filePath := filepath.Join(repoDir, "tracked.txt")
	if err := os.WriteFile(filePath, []byte("original"), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}
	_ = exec.Command("git", "-C", repoDir, "add", "tracked.txt").Run()
	_ = exec.Command("git", "-C", repoDir, "commit", "-m", "msg1").Run()
	h1Ref, _ := repo.Head()
	h1 := h1Ref.Hash()

	if err := os.WriteFile(filePath, []byte("new state in history"), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}
	_ = exec.Command("git", "-C", repoDir, "add", "tracked.txt").Run()
	_ = exec.Command("git", "-C", repoDir, "commit", "-m", "msg2").Run()
	h2Ref, _ := repo.Head()
	h2Str := h2Ref.Hash().String()
	t.Logf("h2: %s", h2Str)

	// Verify system git sees h2
	checkCmd := exec.Command("git", "-C", repoDir, "rev-parse", h2Str)
	if out, err := checkCmd.CombinedOutput(); err != nil {
		t.Fatalf("System git does not see h2 (%s): %v\nOutput: %s", h2Str, err, string(out))
	}

	// Reset back to h1 for the "before" state
	_ = exec.Command("git", "-C", repoDir, "reset", "--hard", h1.String()).Run()

	// 2. Introduce UNSTAGED changes in worktree (This work must NOT be lost)
	unstagedPath := filepath.Join(repoDir, "tracked.txt")
	if err := os.WriteFile(unstagedPath, []byte("concurrent work"), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	// 3. Introduce UNTRACKED file
	untrackedPath := filepath.Join(repoDir, "untracked.txt")
	if err := os.WriteFile(untrackedPath, []byte("new file"), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	// 4. Run non-destructive sync (Safe Pointer Move)
	headRef, _ := repo.Head()
	currentBranch := headRef.Name().Short()
	t.Logf("Current branch: %s", currentBranch)

	seepage := &models.SeepageContext{
		TargetBranch: currentBranch,
		RepoRoot:     repoDir,
		Options: models.Options{
			ExperimentalGoGit: false, // Use system git reset --mixed
		},
	}

	gitSvc := gitseepGit.NewTestGitService(repo)
	_, err := gitSvc.SafeFinalize(seepage, h2Str, nil)
	if err != nil {
		t.Fatalf("SafeFinalize failed: %v", err)
	}

	// 5. Verify Branch Pointer moved to h2
	head, _ := repo.Head()
	if head.Hash().String() != h2Str {
		t.Errorf("Expected branch to be at %s, got %s", h2Str, head.Hash().String())
	}

	// 6. Verify Worktree is PRESERVED (Non-Destructive)
	data, _ := os.ReadFile(unstagedPath)
	if string(data) != "concurrent work" {
		t.Errorf("Expected unstaged changes to be preserved, but got %q", string(data))
	}

	// 7. Verify Untracked file is PRESERVED
	if _, err := os.Stat(untrackedPath); os.IsNotExist(err) {
		t.Errorf("Untracked file was lost")
	}

	// 8. Verify Index (Staging Area) matches h2 (Mixed Reset effect)
	status, _ := wt.Status()
	s := status.File("tracked.txt")
	// If reset --mixed worked, the file should be modified in worktree
	// but the index version should match the commit (so it shows as modified).
	if s.Worktree != git.Modified {
		t.Errorf("Expected tracked.txt to be Modified in worktree, got %v", s.Worktree)
	}
}

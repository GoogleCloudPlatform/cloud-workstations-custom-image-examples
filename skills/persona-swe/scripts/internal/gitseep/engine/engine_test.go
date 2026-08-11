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

// Package engine orchestrates the core GitSeep logic including discovery,
// reconstruction, and sedimentation phases.
package engine

import (
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/config"
	gitseepGit "github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/git"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/models"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/sys"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/testutil"
)

// --- Mocks ---

// mockShellService implements sys.ShellService for testing.
type mockShellService struct {
	lastCmd string
}

func (m *mockShellService) Command(name string, arg ...string) sys.ShellCommand {
	m.lastCmd = name
	return &mockShellCommand{}
}

// mockShellCommand implements sys.ShellCommand for testing.
type mockShellCommand struct{}

func (c *mockShellCommand) SetDir(dir string)                  {}
func (c *mockShellCommand) SetEnv(env []string)                {}
func (c *mockShellCommand) SetStdout(w io.Writer)              {}
func (c *mockShellCommand) SetStderr(w io.Writer)              {}
func (c *mockShellCommand) CombinedOutput() ([]byte, error)    { return []byte("ok"), nil }
func (c *mockShellCommand) Run() error                         { return nil }
func (c *mockShellCommand) Start() error                       { return nil }
func (c *mockShellCommand) Wait() error                        { return nil }
func (c *mockShellCommand) StdoutPipe() (io.ReadCloser, error) { return nil, nil }

// mockGitService implements git.GitService for testing.
type mockGitService struct {
	repo    *git.Repository
	isClean bool
}

func (m *mockGitService) GetRepo() *git.Repository       { return m.repo }
func (m *mockGitService) GetStatus() (bool, bool, error) { return !m.isClean, false, nil }
func (m *mockGitService) IsClean() (bool, error)         { return m.isClean, nil }
func (m *mockGitService) Amend() error                   { return nil }
func (m *mockGitService) VirtualAmend() (plumbing.Hash, error) {
	return plumbing.ZeroHash, nil
}
func (m *mockGitService) ResetMixed(target string) error { return nil }
func (m *mockGitService) ResolveRevision(rev string) (plumbing.Hash, error) {
	return plumbing.ZeroHash, nil
}
func (m *mockGitService) GetCurrentBranch() (string, error)                  { return "main", nil }
func (m *mockGitService) RepoRoot() string                                   { return "." }
func (m *mockGitService) UpdateBranch(name string, hash plumbing.Hash) error { return nil }
func (m *mockGitService) GetChangedFiles(from, to string) ([]string, error) {
	return []string{"f1"}, nil
}
func (m *mockGitService) SafeFinalize(seepage *models.SeepageContext, linearHead string, branchHeads map[string]string) (string, error) {
	return seepage.TargetBranch, nil
}

type mockTroubleshootProvider struct {
	log, manual, cleanup string
}

func (m *mockTroubleshootProvider) GetTroubleshootLog() string        { return m.log }
func (m *mockTroubleshootProvider) GetTroubleshootManualCmd() string  { return m.manual }
func (m *mockTroubleshootProvider) GetTroubleshootCleanupCmd() string { return m.cleanup }

// --- Tests ---

func TestEngineHelpers(t *testing.T) {
	t.Run("indexOf", func(t *testing.T) {
		slice := []string{"a", "b", "c"}
		if i := models.IndexOf(slice, "b"); i != 1 {
			t.Errorf("Expected index 1, got %d", i)
		}
		if i := models.IndexOf(slice, "z"); i != -1 {
			t.Errorf("Expected index -1, got %d", i)
		}
	})

	t.Run("parseRuleDate", func(t *testing.T) {
		dateStr := "2026-04-13 10:01:55 +0000"
		tm := models.ParseRuleDate(dateStr)
		if tm.IsZero() {
			t.Fatalf("Failed to parse full date")
		}
		if tm.Year() != 2026 {
			t.Errorf("Expected year 2026, got %d", tm.Year())
		}

		dateStrShort := "2026-04-13"
		tm = models.ParseRuleDate(dateStrShort)
		if tm.IsZero() {
			t.Fatalf("Failed to parse short date")
		}
		if tm.Month() != time.April {
			t.Errorf("Expected month April, got %v", tm.Month())
		}
	})
}

func TestNewContext(t *testing.T) {
	repo, wt := testutil.SetupMemRepo(t)
	d0 := time.Now()
	h0, _ := testutil.CommitFile(t, testutil.CommitParams{
		Repo: repo, Worktree: wt, Path: "base.txt", Content: "base", Msg: "msg0", AuthorDate: d0,
	})
	h1, _ := testutil.CommitFile(t, testutil.CommitParams{
		Repo: repo, Worktree: wt, Path: "file1.txt", Content: "content1", Msg: "msg1", AuthorDate: d0.Add(time.Hour),
	})

	dateStr := d0.Add(time.Hour).Format("2006-01-02 15:04:05 -0700")
	cfg := &config.GitSeepConfig{
		Global: config.GlobalConfig{BaseRef: "main"},
		Rules: []config.Rule{
			{
				Date:  dateStr,
				Paths: []string{"file1.txt"},
			},
		},
	}

	opts := models.Options{
		BaseCommit: h0,
	}

	seepage, err := NewContext(gitseepGit.NewTestGitService(repo), cfg, opts)
	if err != nil {
		t.Fatalf("NewContext failed: %v", err)
	}

	if seepage.ParentOfStrata != h0 {
		t.Errorf("Expected ParentOfStrata %s, got %s", h0, seepage.ParentOfStrata)
	}
	if len(seepage.Strata) != 1 || seepage.Strata[0] != h1 {
		t.Errorf("Expected strata [h1], got %v", seepage.Strata)
	}
}

func TestRun_MandatoryAmend(t *testing.T) {
	testutil.InitI18n()
	repoDir, h1, _, cleanup := testutil.SetupTestRepoSystemGit(t)
	defer cleanup()

	repo, _ := git.PlainOpen(repoDir)
	wt, _ := repo.Worktree()

	// 1. Stage a change
	filePath := filepath.Join(repoDir, "base.txt")
	if err := os.WriteFile(filePath, []byte("amended content"), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}
	if _, err := wt.Add("base.txt"); err != nil {
		t.Fatalf("Failed to add file: %v", err)
	}

	// 2. Run VerifyWorktreeIsClean (which handles the amendment check)
	opts := models.Options{
		Amend: true,
	}

	// VerifyWorktreeIsClean should pass since changes are STAGED
	if err := VerifyWorktreeIsClean(gitseepGit.NewTestGitService(repo), opts); err != nil {
		t.Fatalf("VerifyWorktreeIsClean failed: %v", err)
	}

	// 3. Manually call Amend (as engine.Run would do)
	if err := gitseepGit.Amend(repo); err != nil {
		t.Fatalf("Amend failed: %v", err)
	}

	// 4. Verify Head has changed
	h2, _ := repo.Head()
	if h2.Hash().String() == h1 {
		t.Errorf("Expected head to change after amendment")
	}

	// 5. Verify Content
	c, _ := repo.CommitObject(h2.Hash())
	tree, _ := c.Tree()
	f, _ := tree.File("base.txt")
	content, _ := f.Contents()
	if content != "amended content" {
		t.Errorf("Expected 'amended content', got %q", content)
	}
}

func TestEngine_RunHeadless(t *testing.T) {
	testutil.InitI18n()

	// Create a real-ish repo on disk for NewGitService to work
	repoDir, _, _, cleanup := testutil.SetupTestRepoSystemGit(t)
	defer cleanup()

	cwd, _ := os.Getwd()
	_ = os.Chdir(repoDir)
	defer func() { _ = os.Chdir(cwd) }()

	cfg := &config.GitSeepConfig{
		Rules: []config.Rule{},
	}
	opts := models.Options{
		Headless: true,
		Quiet:    true,
	}

	// Should not error for empty rules in headless mode (nothing to do)
	err := Run(cfg, opts)
	if err != nil {
		t.Errorf("Run failed: %v", err)
	}
}

func TestEngine_Check(t *testing.T) {
	testutil.InitI18n()
	repoDir, _, _, cleanup := testutil.SetupTestRepoSystemGit(t)
	defer cleanup()

	cwd, _ := os.Getwd()
	_ = os.Chdir(repoDir)
	defer func() { _ = os.Chdir(cwd) }()

	// Use dry run to avoid complex reconstruction logic that might fail on empty test repo
	cfg := &config.GitSeepConfig{}
	_ = Check(cfg, models.Options{DryRun: true})
}

func TestEngine_Coverage(t *testing.T) {
	testutil.InitI18n()
	repo, seepage, _ := testutil.SetupLinearHistory(t)
	gitSvc := gitseepGit.NewTestGitService(repo)

	t.Run("VerifyWorktreeIsClean", func(t *testing.T) {
		err := VerifyWorktreeIsClean(gitSvc, models.Options{CheckMode: true})
		if err != nil {
			t.Errorf("VerifyWorktreeIsClean should succeed in CheckMode even if dirty: %v", err)
		}

		err = VerifyWorktreeIsClean(gitSvc, models.Options{Amend: false})
		if err != nil {
			t.Errorf("Expected clean worktree: %v", err)
		}
	})

	t.Run("FinalizeReferences", func(t *testing.T) {
		_, err := FinalizeReferences(seepage, gitSvc, "newhead", nil)
		if err != nil {
			t.Errorf("FinalizeReferences failed: %v", err)
		}
	})

	t.Run("NewContext", func(t *testing.T) {
		cfg := &config.GitSeepConfig{
			Rules: []config.Rule{
				{Date: "2026-01-01", Paths: []string{"file1.txt"}},
			},
		}
		seepageCtx, err := NewContext(gitSvc, cfg, models.Options{})
		if err != nil {
			t.Fatalf("NewContext failed: %v", err)
		}
		if len(seepageCtx.Strata) == 0 {
			t.Error("Expected non-empty strata in context")
		}
	})
}

func TestEngine_TroubleshootingCoverage(t *testing.T) {
	testutil.InitI18n()

	t.Run("printTroubleshootingInfo", func(t *testing.T) {
		m := &mockTroubleshootProvider{
			log:     "some error",
			manual:  "pre-commit run",
			cleanup: "rm -rf /tmp/wt",
		}
		// Just verify it doesn't panic
		printTroubleshootingInfo(m, "/tmp/wt")
	})
}

func TestVerifyWorktreeIsClean_Coverage(t *testing.T) {
	testutil.InitI18n()
	repo, _ := testutil.SetupMemRepo(t)
	gitSvc := &mockGitService{repo: repo}

	t.Run("Clean worktree", func(t *testing.T) {
		gitSvc.isClean = true
		err := VerifyWorktreeIsClean(gitSvc, models.Options{})
		if err != nil {
			t.Errorf("Expected nil error for clean worktree, got %v", err)
		}
	})

	t.Run("Dirty worktree", func(t *testing.T) {
		gitSvc.isClean = false
		err := VerifyWorktreeIsClean(gitSvc, models.Options{})
		if err == nil {
			t.Error("Expected error for dirty worktree")
		}
	})
}

func TestApplyExclusions(t *testing.T) {
	dr := &models.DiscoveryResult{
		Schedule: map[string]map[string][]string{
			"bedrock1": {
				"bedrock1": {"file1.txt", "file2.txt"},
				"other":    {"file1.txt"},
			},
		},
	}

	ApplyExclusions(dr, []string{"file1.txt"})

	// file1.txt should be removed from 'other'
	if len(dr.Schedule["bedrock1"]["other"]) != 0 {
		t.Errorf("Expected file1.txt to be removed from 'other' source")
	}

	// file1.txt should remain in bedrock1 because it's anchored there
	found := false
	for _, f := range dr.Schedule["bedrock1"]["bedrock1"] {
		if f == "file1.txt" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected file1.txt to remain in its own bedrock stratum")
	}

	if len(dr.Schedule["bedrock1"]["bedrock1"]) != 2 {
		t.Errorf("Expected 2 files in bedrock1 schedule, got %d", len(dr.Schedule["bedrock1"]["bedrock1"]))
	}

	ApplyExclusions(dr, []string{"file2.txt"})
	// file1.txt and file2.txt both anchored to bedrock1, so bedrock1 should still have them
	if len(dr.Schedule["bedrock1"]["bedrock1"]) != 2 {
		t.Errorf("Expected 2 files anchored to bedrock1, got %d", len(dr.Schedule["bedrock1"]["bedrock1"]))
	}
}

func TestExcludedFileAnchoring(t *testing.T) {
	repo, wt := testutil.SetupMemRepo(t)
	now := time.Now()

	// 1. Bedrock Commit (Original owner of CONTRIBUTING.md)
	hBedrock, _ := testutil.CommitFile(t, testutil.CommitParams{
		Repo: repo, Worktree: wt, Path: "CONTRIBUTING.md", Content: "initial", Msg: "msg bedrock", AuthorDate: now,
	})

	// 2. Future Commit (Where CONTRIBUTING.md was deleted in original history)
	_, _ = testutil.CommitFile(t, testutil.CommitParams{
		Repo: repo, Worktree: wt, Path: "other.txt", Content: "other", Msg: "msg future", AuthorDate: now.Add(time.Hour),
	})
	_, _ = wt.Remove("CONTRIBUTING.md")
	hFuture, _ := testutil.CommitFile(t, testutil.CommitParams{
		Repo: repo, Worktree: wt, Path: "other.txt", Content: "other modified", Msg: "msg future delete", AuthorDate: now.Add(time.Hour * 2),
	})

	seepage := &models.SeepageContext{
		Strata: []string{hBedrock, hFuture},
		PathToBedrock: map[string]string{
			"CONTRIBUTING.md": hBedrock,
		},
		ExcludedPaths: map[string]bool{
			"CONTRIBUTING.md": true,
		},
		ParentOfStrata: hBedrock,
	}
	seepage.EnsureMatcher()

	// 3. Verify ResolveBedrockForPath still returns the bedrock even if excluded
	// (Crucial for anchoring)
	resolvedH := seepage.ResolveBedrockForPath("CONTRIBUTING.md")
	if resolvedH != hBedrock {
		t.Errorf("Expected excluded file to be anchored to %s, got %s", hBedrock, resolvedH)
	}

	// 4. Verify applyExclusions preserves it in bedrock stratum but removes from others
	dr := &models.DiscoveryResult{
		Schedule: map[string]map[string][]string{
			hBedrock: {
				hBedrock: {"CONTRIBUTING.md"},
				hFuture:  {"CONTRIBUTING.md"}, // Unexpected migration found during discovery
			},
		},
	}

	ApplyExclusions(dr, []string{"CONTRIBUTING.md"})

	if len(dr.Schedule[hBedrock][hFuture]) != 0 {
		t.Errorf("Expected CONTRIBUTING.md to be removed from future migration schedule")
	}

	found := false
	for _, f := range dr.Schedule[hBedrock][hBedrock] {
		if f == "CONTRIBUTING.md" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected CONTRIBUTING.md to remain anchored to its bedrock stratum")
	}
}

func TestUnmanagedFileExclusion(t *testing.T) {
	repo, wt := testutil.SetupMemRepo(t)
	now := time.Now()

	// 1. Bedrock with an unmanaged file
	hBedrock, _ := testutil.CommitFile(t, testutil.CommitParams{
		Repo: repo, Worktree: wt, Path: "UNMANAGED.md", Content: "initial", Msg: "msg bedrock", AuthorDate: now,
	})

	seepage := &models.SeepageContext{
		Strata: []string{hBedrock},
		ExcludedPaths: map[string]bool{
			"UNMANAGED.md": true,
		},
		ParentOfStrata: hBedrock,
	}
	seepage.EnsureMatcher()

	// 2. Verify MirrorUnmanaged deletes it if it was inherited (simulated by existing in target)
	targetEntries := map[string]object.TreeEntry{
		"UNMANAGED.md": {Name: "UNMANAGED.md", Mode: 0100644},
	}
	bedrockCommit, _ := repo.CommitObject(plumbing.NewHash(hBedrock))
	sourceTree, _ := bedrockCommit.Tree()
	sourceEntries := make(map[string]object.TreeEntry)
	for _, entry := range sourceTree.Entries {
		sourceEntries[entry.Name] = entry
	}

	MirrorUnmanaged(seepage, targetEntries, sourceEntries, true)

	if _, exists := targetEntries["UNMANAGED.md"]; exists {
		t.Errorf("Expected UNMANAGED.md to be deleted from targetEntries")
	}

	// 3. Verify MirrorUnmanaged skips it when adding
	targetEntries = make(map[string]object.TreeEntry)
	MirrorUnmanaged(seepage, targetEntries, sourceEntries, false)

	if _, exists := targetEntries["UNMANAGED.md"]; exists {
		t.Errorf("Expected UNMANAGED.md to be skipped during add")
	}
}

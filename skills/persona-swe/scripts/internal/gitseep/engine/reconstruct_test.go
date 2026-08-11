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
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/models"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/testutil"
)

// mockStrategy simply returns the parent hash as the "new" hash, effectively squashing history for testing purposes.
type mockStrategy struct{}

func (s *mockStrategy) Project(p ReconstructionParams, commitHashStr string, parentHash plumbing.Hash) (plumbing.Hash, error) {
	return parentHash, nil
}

func TestReconstructHistorySuite(t *testing.T) {
	repo, wt := testutil.SetupMemRepo(t)
	d0 := time.Now()
	h0, _ := testutil.CommitFile(t, testutil.CommitParams{
		Repo: repo, Worktree: wt, Path: "base.txt", Content: "base", Msg: "base msg", AuthorDate: d0,
	})

	t.Run("Basic In-Memory Reconstruction", func(t *testing.T) {
		h1, _ := testutil.CommitFile(t, testutil.CommitParams{
			Repo: repo, Worktree: wt, Path: "file1.txt", Content: "content1", Msg: "msg1", AuthorDate: d0.Add(time.Hour),
		})
		h2, _ := testutil.CommitFile(t, testutil.CommitParams{
			Repo: repo, Worktree: wt, Path: "file2.txt", Content: "content2", Msg: "msg2", AuthorDate: d0.Add(time.Hour * 2),
		})
		h3, _ := testutil.CommitFile(t, testutil.CommitParams{
			Repo: repo, Worktree: wt, Path: "file1.txt", Content: "content1_updated", Msg: "msg3", AuthorDate: d0.Add(time.Hour * 3),
		})

		seepage := &models.SeepageContext{
			RepoRoot:       "/mem",
			OriginalHead:   h3,
			CurrentBranch:  "main",
			TargetBranch:   "main",
			ParentOfStrata: h0,
			Strata:         []string{h1, h2, h3},
			PathToBedrock: map[string]string{
				"file1.txt": h1,
			},
			Options: models.Options{DryRun: false, StageOnly: true},
		}

		dr := &models.DiscoveryResult{
			Schedule: map[string]map[string][]string{
				h1: {
					h1: {"file1.txt"},
					h3: {"file1.txt"},
				},
			},
		}

		linearCommits, _, err := ReconstructHistory(ReconstructionParams{Seepage: seepage, Result: dr, Repo: repo})
		if err != nil {
			t.Fatalf("ReconstructHistory failed: %v", err)
		}

		if len(linearCommits) != 3 {
			t.Errorf("Expected 3 linear commits mapped, got %d", len(linearCommits))
		}

		newH1Str := linearCommits[h1]
		newH1Obj, _ := repo.CommitObject(plumbing.NewHash(newH1Str))
		tree1, _ := newH1Obj.Tree()
		file1Obj, err := tree1.File("file1.txt")
		if err != nil {
			t.Fatalf("file1.txt missing in reconstructed Bedrock: %v", err)
		}
		content, _ := file1Obj.Contents()
		if content != "content1_updated" {
			t.Errorf("Bedrock did not capture final state. Expected 'content1_updated', got '%s'", content)
		}
	})

	t.Run("Faithful Mirroring (Unmanaged History)", func(t *testing.T) {
		repo, wt := testutil.SetupMemRepo(t)
		d := time.Now()
		h0, _ := testutil.CommitFile(t, testutil.CommitParams{
			Repo: repo, Worktree: wt, Path: "base.txt", Content: "base", Msg: "msg0", AuthorDate: d,
		})

		h1, _ := testutil.CommitFile(t, testutil.CommitParams{
			Repo: repo, Worktree: wt, Path: "unmanaged.txt", Content: "content", Msg: "msg1", AuthorDate: d.Add(time.Hour),
		})
		h2, _ := testutil.CommitFile(t, testutil.CommitParams{
			Repo: repo, Worktree: wt, Path: "managed.txt", Content: "m1", Msg: "msg2", AuthorDate: d.Add(time.Hour * 2),
		})

		seepage := &models.SeepageContext{
			OriginalHead:   h2,
			ParentOfStrata: h0,
			Strata:         []string{h1, h2},
			PathToBedrock: map[string]string{
				"managed.txt": h1,
			},
			Options: models.Options{DryRun: false, StageOnly: true},
		}

		dr := &models.DiscoveryResult{
			Schedule: map[string]map[string][]string{
				h1: {h1: {"managed.txt"}, h2: {"managed.txt"}},
			},
		}

		linearCommits, _, err := ReconstructHistory(ReconstructionParams{Seepage: seepage, Result: dr, Repo: repo})
		if err != nil {
			t.Fatalf("ReconstructHistory failed: %v", err)
		}

		h1Obj, _ := repo.CommitObject(plumbing.NewHash(linearCommits[h1]))
		tree, _ := h1Obj.Tree()
		if _, err := tree.File("unmanaged.txt"); err != nil {
			t.Errorf("unmanaged.txt lost during mirroring")
		}
	})

	t.Run("Surface Injection (.gitseep.yaml preservation)", func(t *testing.T) {
		repo, wt := testutil.SetupMemRepo(t)
		d := time.Now()
		h0, _ := testutil.CommitFile(t, testutil.CommitParams{
			Repo: repo, Worktree: wt, Path: "base.txt", Content: "base", Msg: "msg0", AuthorDate: d,
		})

		h1, _ := testutil.CommitFile(t, testutil.CommitParams{
			Repo: repo, Worktree: wt, Path: "src/main.go", Content: "v1", Msg: "msg1", AuthorDate: d.Add(time.Hour),
		})

		f, _ := wt.Filesystem.Create(".gitseep.yaml")
		_, _ = f.Write([]byte("config: ..."))
		_ = f.Close()
		_, _ = wt.Add(".gitseep.yaml")
		h2, _ := wt.Commit("new config", &git.CommitOptions{
			Author: &object.Signature{Name: "T", Email: "t", When: d.Add(time.Hour * 2)},
		})
		h2Str := h2.String()

		seepage := &models.SeepageContext{
			OriginalHead:   h2Str,
			ParentOfStrata: h0,
			Strata:         []string{h1, h2Str},
			PathToBedrock:  map[string]string{},
			Options:        models.Options{DryRun: false, StageOnly: true},
		}

		linearCommits, _, err := ReconstructHistory(ReconstructionParams{Seepage: seepage, Result: &models.DiscoveryResult{}, Repo: repo})
		if err != nil {
			t.Fatalf("ReconstructHistory failed: %v", err)
		}

		finalObj, _ := repo.CommitObject(plumbing.NewHash(linearCommits[h2Str]))
		tree, _ := finalObj.Tree()
		if _, err := tree.File(".gitseep.yaml"); err != nil {
			t.Errorf(".gitseep.yaml lost during surface injection")
		}
	})
}

func TestReconstructHistory_StrategyPattern(t *testing.T) {
	repo, wt := testutil.SetupMemRepo(t)
	d0 := time.Now()
	h0, _ := testutil.CommitFile(t, testutil.CommitParams{
		Repo: repo, Worktree: wt, Path: "base.txt", Content: "base", Msg: "base msg", AuthorDate: d0,
	})
	h1, _ := testutil.CommitFile(t, testutil.CommitParams{
		Repo: repo, Worktree: wt, Path: "file1.txt", Content: "content1", Msg: "msg1", AuthorDate: d0.Add(time.Hour),
	})

	seepage := &models.SeepageContext{
		ParentOfStrata: h0,
		Strata:         []string{h1},
		Options:        models.Options{DryRun: false},
	}

	t.Run("Default Strategy", func(t *testing.T) {
		p := ReconstructionParams{
			Seepage: seepage,
			Result:  &models.DiscoveryResult{Schedule: make(map[string]map[string][]string)},
			Repo:    repo,
		}
		linearCommits, _, err := ReconstructHistory(p)
		if err != nil {
			t.Fatalf("ReconstructHistory failed: %v", err)
		}
		if len(linearCommits) != 1 {
			t.Errorf("Expected 1 mapped commit")
		}
		if linearCommits[h1] == h0 {
			t.Errorf("Default strategy should have created a NEW commit hash")
		}
	})

	t.Run("Custom Mock Strategy", func(t *testing.T) {
		p := ReconstructionParams{
			Seepage:  seepage,
			Result:   &models.DiscoveryResult{Schedule: make(map[string]map[string][]string)},
			Repo:     repo,
			Strategy: &mockStrategy{},
		}
		linearCommits, tip, err := ReconstructHistory(p)
		if err != nil {
			t.Fatalf("ReconstructHistory failed: %v", err)
		}
		if linearCommits[h1] != h0 {
			t.Errorf("Mock strategy should have returned parent hash h0, got %s", linearCommits[h1])
		}
		if tip != h0 {
			t.Errorf("Expected tip to be h0")
		}
	})
}

func TestStateProjectionStrategy_ErrorPaths(t *testing.T) {
	repo, _ := testutil.SetupMemRepo(t)
	strategy := &StateProjectionStrategy{}

	t.Run("Missing Original Commit", func(t *testing.T) {
		p := ReconstructionParams{
			Repo:    repo,
			Seepage: &models.SeepageContext{},
		}
		_, err := strategy.Project(p, "nonexistent", plumbing.ZeroHash)
		if err == nil {
			t.Error("Expected error for missing original commit")
		}
	})

	t.Run("Missing Parent Commit", func(t *testing.T) {
		h0, _ := testutil.CommitFile(t, testutil.CommitParams{Repo: repo, Path: "f.txt", Content: "c"})
		p := ReconstructionParams{
			Repo:    repo,
			Seepage: &models.SeepageContext{},
		}
		// parentHash is ZeroHash but we didn't mock a root commit correctly for this strategy
		_, err := strategy.Project(p, h0, plumbing.NewHash("0000000000000000000000000000000000000001"))
		if err == nil {
			t.Error("Expected error for missing parent commit")
		}
	})
}

func TestReconstructHistory_EmptySchedule(t *testing.T) {
	repo, _ := testutil.SetupMemRepo(t)
	h0, _ := testutil.CommitFile(t, testutil.CommitParams{Repo: repo, Path: "f.txt", Content: "c"})

	seepage := &models.SeepageContext{
		ParentOfStrata: h0,
		OriginalHead:   h0,
		Strata:         []string{h0},
	}
	dr := &models.DiscoveryResult{
		Schedule: make(map[string]map[string][]string),
	}

	linear, head, err := ReconstructHistory(ReconstructionParams{
		Repo:    repo,
		Seepage: seepage,
		Result:  dr,
	})

	if err != nil {
		t.Fatalf("ReconstructHistory failed: %v", err)
	}
	if head == "" {
		t.Error("Expected non-empty head")
	}
	if len(linear) != 1 {
		t.Errorf("Expected 1 mapped commit, got %d", len(linear))
	}
}

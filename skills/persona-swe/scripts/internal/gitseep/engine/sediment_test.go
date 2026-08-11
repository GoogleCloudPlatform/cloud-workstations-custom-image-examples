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
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/gitutil"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/models"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/testutil"
)

func TestPerformSedimentationInMemory(t *testing.T) {
	repo, wt := testutil.SetupMemRepo(t)

	d0 := time.Now()
	h0, _ := testutil.CommitFile(t, testutil.CommitParams{
		Repo: repo, Worktree: wt, Path: "base.txt", Content: "base", Msg: "base msg", AuthorDate: d0,
	})

	// Create a new branch "parent_branch" off h0
	_, _ = testutil.CommitFile(t, testutil.CommitParams{
		Repo: repo, Worktree: wt, Path: "parent.txt", Content: "parent_content", Msg: "parent msg", AuthorDate: d0.Add(time.Hour),
	})
	parentBranchHash, _ := repo.Head()
	_ = repo.Storer.SetReference(plumbing.NewHashReference(plumbing.NewBranchReferenceName("parent_branch"), parentBranchHash.Hash()))

	// Go back to h0 to simulate main line
	_ = wt.Checkout(&git.CheckoutOptions{Hash: plumbing.NewHash(h0)})

	// The linear commit we want to cherry-pick
	linearH, _ := testutil.CommitFile(t, testutil.CommitParams{
		Repo: repo, Worktree: wt, Path: "feat.txt", Content: "feat_content", Msg: "feat msg", AuthorDate: d0.Add(time.Hour * 2),
	})

	seepage := &models.SeepageContext{
		DateToBranch: map[string]string{
			"d1": "feat_branch",
		},
		BranchToParent: map[string]string{
			"feat_branch": "parent_branch",
		},
		DateToHash: map[string]string{
			"d1": h0, // Use base as bedrock for test
		},
		Strata:         []string{h0},
		ParentOfStrata: h0,
		ResolvedRules: map[string][]string{
			h0: {"feat.txt"},
		},
		PathToBedrock: map[string]string{
			"feat.txt": h0,
		},
		Options: models.Options{DryRun: false},
	}
	seepage.EnsureMatcher()

	orderedBranches := []string{"feat_branch"}
	linearCommits := map[string]string{
		h0: linearH,
	}

	// Managed files MUST be in the schedule to be projected.
	dr := &models.DiscoveryResult{
		Schedule: map[string]map[string][]string{
			h0: {
				linearH: {"feat.txt"},
			},
		},
	}

	branchHeads, err := PerformSedimentation(SedimentationParams{
		Seepage:         seepage,
		Result:          dr,
		Repo:            repo,
		OrderedBranches: orderedBranches,
		LinearCommits:   linearCommits,
	})
	if err != nil {
		t.Fatalf("Sedimentation failed: %v", err)
	}

	// Verify feat_branch was created
	hashStr, ok := branchHeads["feat_branch"]
	if !ok {
		t.Fatalf("feat_branch not created")
	}

	// Verify the commit contains both parent.txt and feat.txt
	featCommit, _ := repo.CommitObject(plumbing.NewHash(hashStr))
	tree, _ := featCommit.Tree()

	_, err = tree.File("parent.txt")
	if err != nil {
		t.Errorf("parent.txt missing from cherry-picked feature branch")
	}

	_, err = tree.File("feat.txt")
	if err != nil {
		t.Errorf("feat.txt missing from cherry-picked feature branch")
	}
}

func TestSedimentationZeroMutation(t *testing.T) {
	repo, wt := testutil.SetupMemRepo(t)
	d0 := time.Now()
	h0, _ := testutil.CommitFile(t, testutil.CommitParams{
		Repo: repo, Worktree: wt, Path: "base.txt", Content: "base", Msg: "base msg", AuthorDate: d0,
	})

	// The parent branch
	hP, _ := testutil.CommitFile(t, testutil.CommitParams{
		Repo: repo, Worktree: wt, Path: "parent.txt", Content: "p", Msg: "p msg", AuthorDate: d0.Add(time.Hour),
	})
	_ = repo.Storer.SetReference(plumbing.NewHashReference(plumbing.NewBranchReferenceName("parent_branch"), plumbing.NewHash(hP)))

	// Create the feat_branch with a specific tree and message
	_ = wt.Checkout(&git.CheckoutOptions{Hash: plumbing.NewHash(hP)})
	hF, _ := testutil.CommitFile(t, testutil.CommitParams{
		Repo: repo, Worktree: wt, Path: "feat.txt", Content: "f", Msg: "f msg", AuthorDate: d0.Add(time.Hour * 2),
	})
	_ = repo.Storer.SetReference(plumbing.NewHashReference(plumbing.NewBranchReferenceName("feat_branch"), plumbing.NewHash(hF)))

	seepage := &models.SeepageContext{
		DateToBranch:        map[string]string{"d1": "feat_branch"},
		BranchToParent:      map[string]string{"feat_branch": "parent_branch"},
		DateToHash:          map[string]string{"d1": hF},
		Strata:              []string{hF},
		ParentOfStrata:      h0,
		OriginalBranchHeads: map[string]string{"feat_branch": hF},
		Options:             models.Options{DryRun: false},
	}

	orderedBranches := []string{"feat_branch"}
	linearCommits := map[string]string{
		hF: hF,
	}

	branchHeads, err := PerformSedimentation(SedimentationParams{
		Seepage:         seepage,
		Result:          &models.DiscoveryResult{Schedule: make(map[string]map[string][]string)},
		Repo:            repo,
		OrderedBranches: orderedBranches,
		LinearCommits:   linearCommits,
	})
	if err != nil {
		t.Fatalf("Sedimentation failed: %v", err)
	}

	// Check that the branch pointer did NOT move (it still points to hF)
	if branchHeads["feat_branch"] != hF {
		t.Errorf("Expected branch to not be mutated, but it moved from %s to %s", hF, branchHeads["feat_branch"])
	}
}

func TestDateBasedImplicitParent(t *testing.T) {
	repo, wt := testutil.SetupMemRepo(t)
	d := time.Now()
	hBase, _ := testutil.CommitFile(t, testutil.CommitParams{
		Repo: repo, Worktree: wt, Path: "base.txt", Content: "base", Msg: "msg0", AuthorDate: d,
	})

	// Create a feature branch for an OLD bedrock
	hA, _ := testutil.CommitFile(t, testutil.CommitParams{
		Repo: repo, Worktree: wt, Path: "a.txt", Content: "a", Msg: "msgA", AuthorDate: d.Add(time.Hour),
	})
	_ = repo.Storer.SetReference(plumbing.NewHashReference(plumbing.NewBranchReferenceName("feat/A"), plumbing.NewHash(hA)))

	// Feature branch for a NEWER bedrock
	hB, _ := testutil.CommitFile(t, testutil.CommitParams{
		Repo: repo, Worktree: wt, Path: "b.txt", Content: "b", Msg: "msgB", AuthorDate: d.Add(time.Hour * 2),
	})

	seepage := &models.SeepageContext{
		DateToBranch: map[string]string{
			"2026-04-13": "feat/A",
			"2026-04-14": "feat/B",
		},
		DateToHash: map[string]string{
			"2026-04-13": hA,
			"2026-04-14": hB,
		},
		BranchToParent: map[string]string{
			// All implicit
		},
		Strata:         []string{hA, hB},
		ParentOfStrata: hBase,
		Options:        models.Options{DryRun: false},
	}

	// We want to verify that feat/B correctly identifies feat/A as its parent
	// based on the date, even if not explicitly told.
	branchHeads, err := PerformSedimentation(SedimentationParams{
		Seepage:         seepage,
		Result:          &models.DiscoveryResult{Schedule: make(map[string]map[string][]string)},
		Repo:            repo,
		OrderedBranches: []string{"feat/A", "feat/B"},
		LinearCommits:   map[string]string{hA: hA, hB: hB},
	})
	if err != nil {
		t.Fatalf("Sedimentation failed: %v", err)
	}

	commitB, _ := repo.CommitObject(plumbing.NewHash(branchHeads["feat/B"]))

	if commitB.NumParents() == 0 || commitB.ParentHashes[0].String() != hA {
		t.Errorf("feat/B did not correctly use feat/A as its parent")
	}
}

func TestSharedHistoryOptimization(t *testing.T) {
	repo, wt := testutil.SetupMemRepo(t)
	d0 := time.Now()
	hBase, _ := testutil.CommitFile(t, testutil.CommitParams{
		Repo: repo, Worktree: wt, Path: "base.txt", Content: "base", Msg: "msg0", AuthorDate: d0,
	})

	// Linear chain: Base -> L1 -> L2
	hL1, _ := testutil.CommitFile(t, testutil.CommitParams{
		Repo: repo, Worktree: wt, Path: "a.txt", Content: "a", Msg: "msg1", AuthorDate: d0.Add(time.Hour),
	})
	hL2, _ := testutil.CommitFile(t, testutil.CommitParams{
		Repo: repo, Worktree: wt, Path: "b.txt", Content: "b", Msg: "msg2", AuthorDate: d0.Add(time.Hour * 2),
	})

	seepage := &models.SeepageContext{
		DateToBranch: map[string]string{
			"d1": "feat/A",
			"d2": "feat/B",
		},
		DateToHash: map[string]string{
			"d1": hL1,
			"d2": hL2,
		},
		BranchToParent: map[string]string{
			"feat/A": "",       // implicit base
			"feat/B": "feat/A", // explicit parent
		},
		Strata:              []string{hL1, hL2},
		ParentOfStrata:      hBase,
		OriginalBranchHeads: map[string]string{"feat/A": hL1, "feat/B": hL2},
		Options:             models.Options{DryRun: false},
	}

	// We'll use the linear hashes themselves as the reconstructed commits
	linearCommits := map[string]string{
		hL1: hL1,
		hL2: hL2,
	}

	branchHeads, err := PerformSedimentation(SedimentationParams{
		Seepage:         seepage,
		Result:          &models.DiscoveryResult{Schedule: make(map[string]map[string][]string)},
		Repo:            repo,
		OrderedBranches: []string{"feat/A", "feat/B"},
		LinearCommits:   linearCommits,
	})
	if err != nil {
		t.Fatalf("Sedimentation failed: %v", err)
	}

	// Verify feat/A uses hL1 exactly (since both have hBase as parent)
	if branchHeads["feat/A"] != hL1 {
		t.Errorf("feat/A did not reuse linear commit hash L1")
	}

	// Verify feat/B uses hL2 exactly (since feat/B's parent was set to feat/A which is hL1, and L2's parent is L1)
	if branchHeads["feat/B"] != hL2 {
		t.Errorf("feat/B did not reuse linear commit hash L2")
	}
}

func TestPerformSedimentation_HashStability(t *testing.T) {
	repo, wt := testutil.SetupMemRepo(t)
	now := time.Now()

	// 1. Create a bedrock commit h0
	h0, _ := testutil.CommitFile(t, testutil.CommitParams{
		Repo: repo, Worktree: wt, Path: "base.txt", Content: "base", Msg: "msg0", AuthorDate: now,
	})

	// 2. Create a reconstructed version of h0 (h0Prime)
	// Even with different content, we want to see if the sedimented branch preserves sigs.
	h0Prime, _ := testutil.CommitFile(t, testutil.CommitParams{
		Repo: repo, Worktree: wt, Path: "base.txt", Content: "reconstructed", Msg: "msg0", AuthorDate: now,
	})

	seepage := &models.SeepageContext{
		DateToBranch:        map[string]string{"d1": "feat_branch"},
		DateToHash:          map[string]string{"d1": h0},
		Strata:              []string{h0},
		ParentOfStrata:      h0,
		OriginalBranchHeads: map[string]string{"feat_branch": h0}, // Branch tip matches bedrock
		Options:             models.Options{DryRun: false},
	}

	orderedBranches := []string{"feat_branch"}
	linearCommits := map[string]string{h0: h0Prime}

	branchHeads, err := PerformSedimentation(SedimentationParams{
		Seepage:         seepage,
		Result:          &models.DiscoveryResult{Schedule: make(map[string]map[string][]string)},
		Repo:            repo,
		OrderedBranches: orderedBranches,
		LinearCommits:   linearCommits,
	})

	if err != nil {
		t.Fatalf("Sedimentation failed: %v", err)
	}

	// 3. VERIFY HASH PARITY
	// Since feat_branch pointed exactly to h0, and h0 was reconstructed to h0Prime,
	// and there are no other changes on feat_branch, it MUST reuse h0Prime exactly.
	newHead := branchHeads["feat_branch"]
	if newHead != h0Prime {
		t.Errorf("Expected branch to point to linear hash %s, but got %s (Stability/Parity failure)", h0Prime, newHead)
	}
}

func TestPerformSedimentation_ComplexTopology(t *testing.T) {
	repo, wt := testutil.SetupMemRepo(t)
	d0 := time.Now()

	hBase, _ := testutil.CommitFile(t, testutil.CommitParams{
		Repo: repo, Worktree: wt, Path: "base.txt", Content: "base", Msg: "msg0", AuthorDate: d0,
	})

	// feat/1 off base
	h1, _ := testutil.CommitFile(t, testutil.CommitParams{
		Repo: repo, Worktree: wt, Path: "f1.txt", Content: "c1", Msg: "msg1", AuthorDate: d0.Add(time.Hour),
	})
	_ = repo.Storer.SetReference(plumbing.NewHashReference(plumbing.NewBranchReferenceName("feat/1"), plumbing.NewHash(h1)))

	// feat/2 off base
	h2, _ := testutil.CommitFile(t, testutil.CommitParams{
		Repo: repo, Worktree: wt, Path: "f2.txt", Content: "c2", Msg: "msg2", AuthorDate: d0.Add(time.Hour * 2),
	})
	_ = repo.Storer.SetReference(plumbing.NewHashReference(plumbing.NewBranchReferenceName("feat/2"), plumbing.NewHash(h2)))

	// feat/3 off feat/1
	h3, _ := testutil.CommitFile(t, testutil.CommitParams{
		Repo: repo, Worktree: wt, Path: "f3.txt", Content: "c3", Msg: "msg3", AuthorDate: d0.Add(time.Hour * 3),
	})
	_ = repo.Storer.SetReference(plumbing.NewHashReference(plumbing.NewBranchReferenceName("feat/3"), plumbing.NewHash(h3)))

	seepage := &models.SeepageContext{
		DateToBranch: map[string]string{
			"d1": "feat/1",
			"d2": "feat/2",
			"d3": "feat/3",
		},
		DateToHash: map[string]string{
			"d1": h1,
			"d2": h2,
			"d3": h3,
		},
		BranchToParent: map[string]string{
			"feat/3": "feat/1",
		},
		Strata:         []string{h1, h2, h3},
		ParentOfStrata: hBase,
		Options:        models.Options{DryRun: false},
	}

	linearCommits := map[string]string{
		h1: "new1",
		h2: "new2",
		h3: "new3",
	}

	t.Run("Ordered Branch Execution", func(t *testing.T) {
		// Test with different ordering to ensure stability
		ordered := []string{"feat/1", "feat/2", "feat/3"}
		heads, err := PerformSedimentation(SedimentationParams{
			Seepage:         seepage,
			Result:          &models.DiscoveryResult{Schedule: make(map[string]map[string][]string)},
			Repo:            repo,
			OrderedBranches: ordered,
			LinearCommits:   linearCommits,
		})
		if err != nil {
			t.Fatalf("Sedimentation failed: %v", err)
		}
		if len(heads) != 3 {
			t.Errorf("Expected 3 branch heads, got %d", len(heads))
		}
	})
}

func TestPerformSedimentation_DanglingBranch(t *testing.T) {
	repo, wt := testutil.SetupMemRepo(t)
	now := time.Now()

	h0, _ := testutil.CommitFile(t, testutil.CommitParams{
		Repo: repo, Worktree: wt, Path: "base.txt", Content: "base", Msg: "base", AuthorDate: now,
	})
	hF, _ := testutil.CommitFile(t, testutil.CommitParams{
		Repo: repo, Worktree: wt, Path: "feat.txt", Content: "original", Msg: "feat", AuthorDate: now.Add(time.Hour),
	})

	seepage := &models.SeepageContext{
		DateToBranch:        map[string]string{"d1": "feat_branch"},
		DateToHash:          map[string]string{"d1": h0},
		Strata:              []string{h0},
		ParentOfStrata:      h0,
		OriginalBranchHeads: map[string]string{"feat_branch": hF},
		ResolvedRules: map[string][]string{
			h0: {"some_other_path.txt"},
		},
		Options: models.Options{DryRun: false},
	}

	orderedBranches := []string{"feat_branch"}

	// Simulate reconstruction: h0 is reconstructed to a new commit h0Prime
	h0Prime, _ := testutil.CommitFile(t, testutil.CommitParams{
		Repo: repo, Worktree: wt, Path: "base.txt", Content: "reconstructed base", Msg: "base", AuthorDate: now,
	})

	linearCommits := map[string]string{h0: h0Prime}

	branchHeads, err := PerformSedimentation(SedimentationParams{
		Seepage:         seepage,
		Result:          &models.DiscoveryResult{Schedule: make(map[string]map[string][]string)},
		Repo:            repo,
		OrderedBranches: orderedBranches,
		LinearCommits:   linearCommits,
	})

	if err != nil {
		t.Fatalf("Sedimentation failed: %v", err)
	}

	newHead := branchHeads["feat_branch"]
	if newHead == "" || newHead == hF {
		t.Errorf("Expected a NEW reconstructed head for dangling branch, got %s", newHead)
	}

	commit, _ := repo.CommitObject(plumbing.NewHash(newHead))
	if commit == nil {
		t.Fatal("Reconstructed commit not found")
	}
	tree, _ := commit.Tree()
	if _, err := tree.File("feat.txt"); err != nil {
		t.Errorf("feat.txt lost in reconstruction: %v", err)
	}
}

func TestSedimentationWithMigratedFiles(t *testing.T) {
	repo, wt := testutil.SetupMemRepo(t)
	now := time.Now()

	// 1. Bedrock A
	hA, _ := testutil.CommitFile(t, testutil.CommitParams{
		Repo: repo, Worktree: wt, Path: "a.txt", Content: "old a", Msg: "msg A", AuthorDate: now,
	})

	// Create a dummy commit on the branch to force reconstruction (branch tip != bedrock)
	_ = wt.Checkout(&git.CheckoutOptions{Hash: plumbing.NewHash(hA)})
	hTip, _ := testutil.CommitFile(t, testutil.CommitParams{
		Repo: repo, Worktree: wt, Path: "dummy.txt", Content: "d", Msg: "tip", AuthorDate: now.Add(time.Minute),
	})
	_ = repo.Storer.SetReference(plumbing.NewHashReference(plumbing.NewBranchReferenceName("feat/A"), plumbing.NewHash(hTip)))

	// 2. Future commit with a change to a.txt (managed by A)
	_ = wt.Checkout(&git.CheckoutOptions{Hash: plumbing.NewHash(hA)}) // back to main line
	hFuture, _ := testutil.CommitFile(t, testutil.CommitParams{
		Repo: repo, Worktree: wt, Path: "a.txt", Content: "new a", Msg: "msg future", AuthorDate: now.Add(time.Hour),
	})

	seepage := &models.SeepageContext{
		DateToBranch:        map[string]string{"d1": "feat/A"},
		DateToHash:          map[string]string{"d1": hA},
		Strata:              []string{hA, hFuture},
		OriginalBranchHeads: map[string]string{"feat/A": hTip},
		ResolvedRules: map[string][]string{
			hA: {"a.txt"},
		},
		Options:        models.Options{DryRun: false},
		ParentOfStrata: hA,
	}
	seepage.EnsureMatcher()

	// Discovery Result showing that a.txt from hFuture migrated into hA
	dr := &models.DiscoveryResult{
		Schedule: map[string]map[string][]string{
			hA: {
				hA:      {"a.txt"},
				hFuture: {"a.txt"},
			},
		},
	}

	// We assume hA is reconstructed to hA_Prime (which contains "new a")
	hA_Prime, _ := testutil.CommitFile(t, testutil.CommitParams{
		Repo: repo, Worktree: wt, Path: "a.txt", Content: "new a", Msg: "msg A", AuthorDate: now,
	})

	linearCommits := map[string]string{
		hA: hA_Prime,
	}

	branchHeads, err := PerformSedimentation(SedimentationParams{
		Seepage:         seepage,
		Result:          dr,
		Repo:            repo,
		OrderedBranches: []string{"feat/A"},
		LinearCommits:   linearCommits,
	})

	if err != nil {
		t.Fatalf("Sedimentation failed: %v", err)
	}

	// Verify feat/A has the NEW version of a.txt
	newHead := branchHeads["feat/A"]
	commit, _ := repo.CommitObject(plumbing.NewHash(newHead))
	tree, _ := commit.Tree()
	f, _ := tree.File("a.txt")
	content, _ := f.Contents()

	if content != "new a" {
		t.Errorf("Expected feat/A to have 'new a', but got '%s'. Migration failed in sedimentation.", content)
	}
}

func TestSedimentationPreservesManagedFiles(t *testing.T) {
	repo, wt := testutil.SetupMemRepo(t)
	now := time.Now()

	// 1. Bedrock A (Empty except for .gitignore)
	hA, _ := testutil.CommitFile(t, testutil.CommitParams{
		Repo: repo, Worktree: wt, Path: ".gitignore", Content: "*", Msg: "msg A", AuthorDate: now,
	})
	_ = repo.Storer.SetReference(plumbing.NewHashReference(plumbing.NewBranchReferenceName("feat/A"), plumbing.NewHash(hA)))

	// 2. A managed file 'managed.txt' that exists in original bedrock but IS being projected
	hManaged, _ := testutil.CommitFile(t, testutil.CommitParams{
		Repo: repo, Worktree: wt, Path: "managed.txt", Content: "v1", Msg: "managed", AuthorDate: now.Add(time.Minute),
	})

	seepage := &models.SeepageContext{
		DateToBranch:        map[string]string{"d1": "feat/A"},
		DateToHash:          map[string]string{"d1": hA},
		Strata:              []string{hA, hManaged},
		OriginalBranchHeads: map[string]string{"feat/A": hA},
		ResolvedRules: map[string][]string{
			hA: {"managed.txt"},
		},
		Options:        models.Options{DryRun: false},
		ParentOfStrata: hA,
	}
	seepage.EnsureMatcher()

	// Discovery shows managed.txt belongs to hA
	dr := &models.DiscoveryResult{
		Schedule: map[string]map[string][]string{
			hA: {
				hManaged: {"managed.txt"},
			},
		},
	}

	// Reconstructed main (hA_Prime) does NOT have managed.txt (simulating it being percolated elsewhere in main)
	hA_Prime, _ := testutil.CommitFile(t, testutil.CommitParams{
		Repo: repo, Worktree: wt, Path: ".gitignore", Content: "*", Msg: "msg A prime", AuthorDate: now,
	})

	linearCommits := map[string]string{hA: hA_Prime}

	branchHeads, err := PerformSedimentation(SedimentationParams{
		Seepage:         seepage,
		Result:          dr,
		Repo:            repo,
		OrderedBranches: []string{"feat/A"},
		LinearCommits:   linearCommits,
	})

	if err != nil {
		t.Fatalf("Sedimentation failed: %v", err)
	}

	// Verify feat/A HAS managed.txt even though hA_Prime doesn't!
	newHead := branchHeads["feat/A"]
	commit, _ := repo.CommitObject(plumbing.NewHash(newHead))
	tree, _ := commit.Tree()
	if _, err := tree.File("managed.txt"); err != nil {
		t.Errorf("managed.txt missing from feature branch. Explicit projection failed.")
	}
}

func TestSedimentationSyncsUnmanagedWithMain(t *testing.T) {
	repo, wt := testutil.SetupMemRepo(t)
	now := time.Now()

	// 1. Bedrock A with 'shared.txt'
	hA, _ := testutil.CommitFile(t, testutil.CommitParams{
		Repo: repo, Worktree: wt, Path: "shared.txt", Content: "old", Msg: "msg A", AuthorDate: now,
	})

	// Branch has a version of 'shared.txt' with a header (simulating user modification)
	_ = wt.Checkout(&git.CheckoutOptions{Hash: plumbing.NewHash(hA)})
	hTip, _ := testutil.CommitFile(t, testutil.CommitParams{
		Repo: repo, Worktree: wt, Path: "shared.txt", Content: "HEADER\nold", Msg: "tip", AuthorDate: now.Add(time.Minute),
	})
	_ = repo.Storer.SetReference(plumbing.NewHashReference(plumbing.NewBranchReferenceName("feat/A"), plumbing.NewHash(hTip)))

	seepage := &models.SeepageContext{
		DateToBranch:        map[string]string{"d1": "feat/A"},
		DateToHash:          map[string]string{"d1": hA},
		Strata:              []string{hA},
		OriginalBranchHeads: map[string]string{"feat/A": hTip},
		Options:             models.Options{DryRun: false},
		ParentOfStrata:      hA,
	}
	seepage.EnsureMatcher()

	// Reconstructed main has 'shared.txt' BUT WITHOUT the header (parity goal)
	hA_Prime, _ := testutil.CommitFile(t, testutil.CommitParams{
		Repo: repo, Worktree: wt, Path: "shared.txt", Content: "old", Msg: "msg A prime", AuthorDate: now,
	})

	linearCommits := map[string]string{hA: hA_Prime}

	branchHeads, err := PerformSedimentation(SedimentationParams{
		Seepage:         seepage,
		Result:          &models.DiscoveryResult{Schedule: make(map[string]map[string][]string)},
		Repo:            repo,
		OrderedBranches: []string{"feat/A"},
		LinearCommits:   linearCommits,
	})

	if err != nil {
		t.Fatalf("Sedimentation failed: %v", err)
	}

	// Verify feat/A has 'shared.txt' WITHOUT the header (matches main)
	newHead := branchHeads["feat/A"]
	commit, _ := repo.CommitObject(plumbing.NewHash(newHead))
	tree, _ := commit.Tree()
	f, _ := tree.File("shared.txt")
	content, _ := f.Contents()

	if content == "HEADER\nold" {
		t.Errorf("Feature branch kept branch-specific version of unmanaged file. Parity with main failed.")
	}
	if content != "old" {
		t.Errorf("Expected 'old', got '%s'", content)
	}
}

func TestSedimentationRecursiveInheritance(t *testing.T) {
	repo, wt := testutil.SetupMemRepo(t)
	now := time.Now()

	// 1. Bedrock A (Empty repo baseline)
	hA, _ := testutil.CommitFile(t, testutil.CommitParams{
		Repo: repo, Worktree: wt, Path: ".gitignore", Content: "*", Msg: "msg A", AuthorDate: now,
	})

	// 2. Branch A adds a unique file
	_ = wt.Checkout(&git.CheckoutOptions{Hash: plumbing.NewHash(hA)})
	hTipA, _ := testutil.CommitFile(t, testutil.CommitParams{
		Repo: repo, Worktree: wt, Path: "branchA.txt", Content: "vA", Msg: "tip A", AuthorDate: now.Add(time.Minute),
	})
	_ = repo.Storer.SetReference(plumbing.NewHashReference(plumbing.NewBranchReferenceName("feat/A"), plumbing.NewHash(hTipA)))

	// 3. Branch B (Child of A) adds another unique file
	_ = wt.Checkout(&git.CheckoutOptions{Hash: plumbing.NewHash(hTipA)})
	hTipB, _ := testutil.CommitFile(t, testutil.CommitParams{
		Repo: repo, Worktree: wt, Path: "branchB.txt", Content: "vB", Msg: "tip B", AuthorDate: now.Add(time.Minute * 2),
	})
	_ = repo.Storer.SetReference(plumbing.NewHashReference(plumbing.NewBranchReferenceName("feat/B"), plumbing.NewHash(hTipB)))

	seepage := &models.SeepageContext{
		DateToBranch: map[string]string{
			"d1": "feat/A",
			"d2": "feat/B",
		},
		DateToHash: map[string]string{
			"d1": hA,
			"d2": hA, // Both anchored to same bedrock for this test
		},
		BranchToParent: map[string]string{
			"feat/B": "feat/A",
		},
		Strata:              []string{hA},
		OriginalBranchHeads: map[string]string{"feat/A": hTipA, "feat/B": hTipB},
		Options:             models.Options{DryRun: false},
		ParentOfStrata:      hA,
	}
	seepage.EnsureMatcher()

	// Reconstructed main (hA_Prime)
	hA_Prime, _ := testutil.CommitFile(t, testutil.CommitParams{
		Repo: repo, Worktree: wt, Path: ".gitignore", Content: "*", Msg: "msg A prime", AuthorDate: now,
	})

	linearCommits := map[string]string{hA: hA_Prime}

	branchHeads, err := PerformSedimentation(SedimentationParams{
		Seepage:         seepage,
		Result:          &models.DiscoveryResult{Schedule: make(map[string]map[string][]string)},
		Repo:            repo,
		OrderedBranches: []string{"feat/A", "feat/B"},
		LinearCommits:   linearCommits,
	})

	if err != nil {
		t.Fatalf("Sedimentation failed: %v", err)
	}

	// Verify feat/B has both files
	newHeadB := branchHeads["feat/B"]
	commitB, _ := repo.CommitObject(plumbing.NewHash(newHeadB))
	treeB, _ := commitB.Tree()

	if _, err := treeB.File("branchA.txt"); err != nil {
		t.Errorf("branchA.txt lost in recursive inheritance")
	}
	if _, err := treeB.File("branchB.txt"); err != nil {
		t.Errorf("branchB.txt lost in recursive inheritance")
	}
}

func TestSedimentationBranchSpecificDeletion(t *testing.T) {
	repo, wt := testutil.SetupMemRepo(t)
	now := time.Now()

	// 1. Bedrock A (Empty repo baseline)
	hA, _ := testutil.CommitFile(t, testutil.CommitParams{
		Repo: repo, Worktree: wt, Path: ".gitignore", Content: "*", Msg: "msg A", AuthorDate: now,
	})

	// 2. Branch A adds a unique file
	_ = wt.Checkout(&git.CheckoutOptions{Hash: plumbing.NewHash(hA)})
	hTipA, _ := testutil.CommitFile(t, testutil.CommitParams{
		Repo: repo, Worktree: wt, Path: "branchA.txt", Content: "vA", Msg: "tip A", AuthorDate: now.Add(time.Minute),
	})
	_ = repo.Storer.SetReference(plumbing.NewHashReference(plumbing.NewBranchReferenceName("feat/A"), plumbing.NewHash(hTipA)))

	// 3. Branch B (Child of A) DELETES branchA.txt
	_ = wt.Checkout(&git.CheckoutOptions{Hash: plumbing.NewHash(hTipA)})
	_, _ = wt.Remove("branchA.txt")
	hTipB, _ := testutil.CommitFile(t, testutil.CommitParams{
		Repo: repo, Worktree: wt, Path: "branchB.txt", Content: "vB", Msg: "tip B delete A", AuthorDate: now.Add(time.Minute * 2),
	})
	_ = repo.Storer.SetReference(plumbing.NewHashReference(plumbing.NewBranchReferenceName("feat/B"), plumbing.NewHash(hTipB)))

	seepage := &models.SeepageContext{
		DateToBranch: map[string]string{
			"d1": "feat/A",
			"d2": "feat/B",
		},
		DateToHash: map[string]string{
			"d1": hA,
			"d2": hA,
		},
		BranchToParent: map[string]string{
			"feat/B": "feat/A",
		},
		Strata:              []string{hA},
		OriginalBranchHeads: map[string]string{"feat/A": hTipA, "feat/B": hTipB},
		Options:             models.Options{DryRun: false},
		ParentOfStrata:      hA,
	}
	seepage.EnsureMatcher()

	// Reconstructed main
	hA_Prime, _ := testutil.CommitFile(t, testutil.CommitParams{
		Repo: repo, Worktree: wt, Path: ".gitignore", Content: "*", Msg: "msg A prime", AuthorDate: now,
	})

	linearCommits := map[string]string{hA: hA_Prime}

	branchHeads, err := PerformSedimentation(SedimentationParams{
		Seepage:         seepage,
		Result:          &models.DiscoveryResult{Schedule: make(map[string]map[string][]string)},
		Repo:            repo,
		OrderedBranches: []string{"feat/A", "feat/B"},
		LinearCommits:   linearCommits,
	})

	if err != nil {
		t.Fatalf("Sedimentation failed: %v", err)
	}

	// Verify feat/B DOES NOT have branchA.txt (deletion preserved)
	newHeadB := branchHeads["feat/B"]
	commitB, _ := repo.CommitObject(plumbing.NewHash(newHeadB))
	treeB, _ := commitB.Tree()

	if _, err := treeB.File("branchA.txt"); err == nil {
		t.Errorf("branchA.txt unexpectedly resurrected in feat/B")
	}
	if _, err := treeB.File("branchB.txt"); err != nil {
		t.Errorf("branchB.txt lost in feature branch")
	}
}

func TestSedimentationPatchProjection(t *testing.T) {
	repo, wt := testutil.SetupMemRepo(t)
	now := time.Now()

	// 1. Bedrock A (Gnome)
	hGnome, _ := testutil.CommitFile(t, testutil.CommitParams{
		Repo: repo, Worktree: wt, Path: "gnome.txt", Content: "v1", Msg: "gnome", AuthorDate: now,
	})
	_ = repo.Storer.SetReference(plumbing.NewHashReference(plumbing.NewBranchReferenceName("feat/gnome"), plumbing.NewHash(hGnome)))

	// 2. Bedrock B (Antigravity) - Skipped by our feature branch
	hAntigravity, _ := testutil.CommitFile(t, testutil.CommitParams{
		Repo: repo, Worktree: wt, Path: "antigravity.txt", Content: "v1", Msg: "antigravity", AuthorDate: now.Add(time.Hour),
	})

	// 3. Bedrock C (ASfP)
	hASfP, _ := testutil.CommitFile(t, testutil.CommitParams{
		Repo: repo, Worktree: wt, Path: "asfp.txt", Content: "v1", Msg: "asfp", AuthorDate: now.Add(time.Hour * 2),
	})

	// 4. Excluded File (Should be removed by Global Exclusion logic)
	hExcluded, _ := testutil.CommitFile(t, testutil.CommitParams{
		Repo: repo, Worktree: wt, Path: "excluded.txt", Content: "gone", Msg: "to be excluded", AuthorDate: now.Add(time.Hour * 3),
	})

	seepage := &models.SeepageContext{
		DateToBranch: map[string]string{
			"d1": "feat/gnome",
			"d3": "feat/asfp",
		},
		DateToHash: map[string]string{
			"d1": hGnome,
			"d2": hAntigravity,
			"d3": hASfP,
			"d4": hExcluded,
		},
		BranchToParent: map[string]string{
			"feat/asfp": "feat/gnome",
		},
		Strata: []string{hGnome, hAntigravity, hASfP, hExcluded},
		ExcludedPaths: map[string]bool{
			"excluded.txt": true,
		},
		Options:        models.Options{DryRun: false},
		ParentOfStrata: hGnome,
	}
	seepage.EnsureMatcher()

	// Reconstructed main (all strata)
	hGnomePrime := hGnome
	hAntigravityPrime := hAntigravity
	hASfPPrime := hASfP
	hExcludedPrime := hExcluded

	linearCommits := map[string]string{
		hGnome:       hGnomePrime,
		hAntigravity: hAntigravityPrime,
		hASfP:        hASfPPrime,
		hExcluded:    hExcludedPrime,
	}

	branchHeads, err := PerformSedimentation(SedimentationParams{
		Seepage:         seepage,
		Result:          &models.DiscoveryResult{Schedule: make(map[string]map[string][]string)},
		Repo:            repo,
		OrderedBranches: []string{"feat/gnome", "feat/asfp"},
		LinearCommits:   linearCommits,
	})

	if err != nil {
		t.Fatalf("Sedimentation failed: %v", err)
	}

	// 5. Verify feat/asfp
	newHeadASfP := branchHeads["feat/asfp"]
	commitASfP, _ := repo.CommitObject(plumbing.NewHash(newHeadASfP))

	// Check parent (should be feat/gnome, NOT antigravity)
	if commitASfP.NumParents() != 1 || commitASfP.ParentHashes[0].String() != hGnomePrime {
		t.Errorf("feat/asfp has incorrect parent. Expected %s, got %v", hGnomePrime[:7], commitASfP.ParentHashes)
	}

	// Check tree contents
	tree, _ := commitASfP.Tree()
	if _, err := tree.File("asfp.txt"); err != nil {
		t.Errorf("asfp.txt missing from feature branch")
	}
	if _, err := tree.File("gnome.txt"); err != nil {
		t.Errorf("gnome.txt missing from feature branch")
	}
	if _, err := tree.File("antigravity.txt"); err == nil {
		t.Errorf("antigravity.txt unexpectedly present in feature branch (it was skipped)")
	}
	if _, err := tree.File("excluded.txt"); err == nil {
		t.Errorf("excluded.txt unexpectedly present in feature branch")
	}

	// 6. Verify DIFF (the critical part)
	// The diff relative to parent (feat/gnome) should match main's diff relative to its parent (antigravity)
	changes, _ := gitutil.GetChangedFiles(repo, hGnomePrime, newHeadASfP)
	if len(changes) != 1 || changes[0] != "asfp.txt" {
		t.Errorf("feat/asfp diff is incorrect. Expected [asfp.txt], got %v", changes)
	}
}

func TestSedimentationHonorsExplicitParent(t *testing.T) {
	repo, wt := testutil.SetupMemRepo(t)
	now := time.Now()

	// 1. Root commit
	hRoot, _ := testutil.CommitFile(t, testutil.CommitParams{
		Repo: repo, Worktree: wt, Path: "root.txt", Content: "root", Msg: "root", AuthorDate: now,
	})

	// 2. Linear stratum 1 (Intermediate, should be skipped)
	hLinear1, _ := testutil.CommitFile(t, testutil.CommitParams{
		Repo: repo, Worktree: wt, Path: "linear1.txt", Content: "v1", Msg: "linear1", AuthorDate: now.Add(time.Hour),
	})

	// 3. Requested Parent Branch (Also anchored to root, but conceptually a branch)
	_ = wt.Checkout(&git.CheckoutOptions{Hash: plumbing.NewHash(hRoot)})
	hParent, _ := testutil.CommitFile(t, testutil.CommitParams{
		Repo: repo, Worktree: wt, Path: "parent.txt", Content: "v1", Msg: "requested parent", AuthorDate: now.Add(time.Minute),
	})
	_ = repo.Storer.SetReference(plumbing.NewHashReference(plumbing.NewBranchReferenceName("feat/parent"), plumbing.NewHash(hParent)))

	// 4. Linear stratum 2 (ASfP - Target)
	_ = wt.Checkout(&git.CheckoutOptions{Hash: plumbing.NewHash(hLinear1)})
	hASfP, _ := testutil.CommitFile(t, testutil.CommitParams{
		Repo: repo, Worktree: wt, Path: "asfp.txt", Content: "v1", Msg: "asfp", AuthorDate: now.Add(time.Hour * 2),
	})

	seepage := &models.SeepageContext{
		DateToBranch: map[string]string{
			"d_parent": "feat/parent",
			"d_asfp":   "feat/asfp",
		},
		DateToHash: map[string]string{
			"d_parent":  hParent,
			"d_linear1": hLinear1,
			"d_asfp":    hASfP,
		},
		BranchToParent: map[string]string{
			"feat/asfp": "feat/parent", // Explicitly requested hierarchy
		},
		Strata:         []string{hRoot, hLinear1, hASfP},
		Options:        models.Options{DryRun: false},
		ParentOfStrata: hRoot,
	}
	seepage.EnsureMatcher()

	linearCommits := map[string]string{
		hRoot:    hRoot,
		hLinear1: hLinear1,
		hASfP:    hASfP,
		hParent:  hParent,
	}

	branchHeads, err := PerformSedimentation(SedimentationParams{
		Seepage:         seepage,
		Result:          &models.DiscoveryResult{Schedule: make(map[string]map[string][]string)},
		Repo:            repo,
		OrderedBranches: []string{"feat/parent", "feat/asfp"},
		LinearCommits:   linearCommits,
	})

	if err != nil {
		t.Fatalf("Sedimentation failed: %v", err)
	}

	// 5. Verify Parentage
	newHeadASfP := branchHeads["feat/asfp"]
	commitASfP, _ := repo.CommitObject(plumbing.NewHash(newHeadASfP))

	// It MUST have feat/parent as parent, NOT linear1
	if commitASfP.NumParents() != 1 || commitASfP.ParentHashes[0].String() != hParent {
		t.Errorf("feat/asfp used incorrect parent. Expected %s (feat/parent), got %v", hParent[:7], commitASfP.ParentHashes)
	}

	// 6. Verify Content
	tree, _ := commitASfP.Tree()
	if _, err := tree.File("asfp.txt"); err != nil {
		t.Errorf("asfp.txt missing")
	}
	if _, err := tree.File("parent.txt"); err != nil {
		t.Errorf("parent.txt missing (not inherited from parent branch)")
	}
	if _, err := tree.File("linear1.txt"); err == nil {
		t.Errorf("linear1.txt unexpectedly present (it should have been skipped)")
	}
}

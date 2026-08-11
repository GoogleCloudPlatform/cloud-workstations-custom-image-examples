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
	"testing"
	"time"

	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/git"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/testutil"
)

func TestGitHistorySuite(t *testing.T) {
	repo, wt := testutil.SetupMemRepo(t)

	d0, _ := time.Parse("2006-01-02", "2026-04-14")
	d1, _ := time.Parse("2006-01-02", "2026-04-15")
	d2, _ := time.Parse("2006-01-02 15:04:05 -0700", "2026-04-16 10:00:00 +0000")

	h0, _ := testutil.CommitFile(t, testutil.CommitParams{
		Repo: repo, Worktree: wt, Path: "base.txt", Content: "base", Msg: "base msg", AuthorDate: d0,
	})
	h1, _ := testutil.CommitFile(t, testutil.CommitParams{
		Repo: repo, Worktree: wt, Path: "file1.txt", Content: "content1", Msg: "msg1", AuthorDate: d1,
	})
	h2, _ := testutil.CommitFile(t, testutil.CommitParams{
		Repo: repo, Worktree: wt, Path: "file2.txt", Content: "content2", Msg: "msg2", AuthorDate: d2,
	})

	t.Run("GetSeepageHistory Batch Resolution", func(t *testing.T) {
		testutil.InitI18n()
		dates := []string{"2026-04-15", "2026-04-16 10:00:00 +0000"}
		res, err := git.GetSeepageHistory(git.NewTestGitService(repo), dates, "")
		if err != nil {
			t.Fatalf("GetSeepageHistory failed: %v", err)
		}

		if res.ResolvedDates["2026-04-15"] != h1 {
			t.Errorf("Expected h1 for d1, got %s", res.ResolvedDates["2026-04-15"])
		}
		if res.ResolvedDates["2026-04-16 10:00:00 +0000"] != h2 {
			t.Errorf("Expected h2 for d2, got %s", res.ResolvedDates["2026-04-16 10:00:00 +0000"])
		}

		if res.OriginalHead != h2 {
			t.Errorf("Expected head %s, got %s", h2, res.OriginalHead)
		}
		if res.ParentOfStrata != h0 {
			t.Errorf("Expected parent %s (of h1), got %s", h0, res.ParentOfStrata)
		}
		if res.BaseHash != "" {
			t.Errorf("Expected empty baseH when no baseRef provided, got %s", res.BaseHash)
		}
		if len(res.Strata) != 2 || res.Strata[0] != h1 || res.Strata[1] != h2 {
			t.Errorf("Expected strata [h1, h2], got %v", res.Strata)
		}
		if res.CurrentBranch == "" || res.CurrentBranch == "HEAD" {
			t.Errorf("Expected a current branch name, got %s", res.CurrentBranch)
		}
	})

	t.Run("GetSeepageHistory with BaseRef", func(t *testing.T) {
		testutil.InitI18n()
		// If we use h1 as base, strata should only include h2
		res, err := git.GetSeepageHistory(git.NewTestGitService(repo), nil, h1)
		if err != nil {
			t.Fatalf("GetSeepageHistory failed: %v", err)
		}

		if len(res.ResolvedDates) != 0 {
			t.Errorf("Expected no resolved dates")
		}
		if res.ParentOfStrata != h1 {
			t.Errorf("Expected parent %s, got %s", h1, res.ParentOfStrata)
		}
		if res.BaseHash != h1 {
			t.Errorf("Expected baseH %s, got %s", h1, res.BaseHash)
		}
		if len(res.Strata) != 1 || res.Strata[0] != h2 {
			t.Errorf("Expected strata [h2], got %v", res.Strata)
		}
	})
}

func TestHistory_Coverage_Extra(t *testing.T) {
	repo, wt := testutil.SetupMemRepo(t)
	now := time.Now()
	h1, _ := testutil.CommitFile(t, testutil.CommitParams{
		Repo: repo, Worktree: wt, Path: "f1", Content: "c1", Msg: "m1", AuthorDate: now,
	})

	t.Run("GetCommit", func(t *testing.T) {
		c, err := git.GetCommit(repo, h1)
		if err != nil {
			t.Fatalf("GetCommit failed: %v", err)
		}
		if c.Hash.String() != h1 {
			t.Errorf("Expected %s, got %s", h1, c.Hash.String())
		}
	})

	t.Run("GetCurrentBranch", func(t *testing.T) {
		b, err := git.GetCurrentBranch(repo)
		if err != nil {
			t.Fatalf("GetCurrentBranch failed: %v", err)
		}
		if b == "" || b == "HEAD" {
			t.Errorf("Expected a branch name, got %s", b)
		}
	})
}

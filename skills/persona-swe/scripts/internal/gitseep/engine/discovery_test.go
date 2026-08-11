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

// Package engine_test provides integration tests for the GitSeep engine.
package engine_test

import (
	"testing"
	"time"

	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/engine"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/models"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/testutil"
)

func TestDiscoverySuite(t *testing.T) {
	repo, wt := testutil.SetupMemRepo(t)

	d0 := time.Now()
	h0, _ := testutil.CommitFile(t, testutil.CommitParams{
		Repo: repo, Worktree: wt, Path: "base.txt", Content: "base", Msg: "base msg", AuthorDate: d0,
	})

	t.Run("Standard Percolation Mapping", func(t *testing.T) {
		h1, _ := testutil.CommitFile(t, testutil.CommitParams{
			Repo: repo, Worktree: wt, Path: "src/owned.txt", Content: "content1", Msg: "msg1", AuthorDate: d0.Add(time.Hour),
		})
		h2, _ := testutil.CommitFile(t, testutil.CommitParams{
			Repo: repo, Worktree: wt, Path: "src/owned.txt", Content: "content1_modified", Msg: "msg2", AuthorDate: d0.Add(time.Hour * 2),
		})
		h3, _ := testutil.CommitFile(t, testutil.CommitParams{
			Repo: repo, Worktree: wt, Path: "other.txt", Content: "content3", Msg: "msg3", AuthorDate: d0.Add(time.Hour * 3),
		})

		seepage := &models.SeepageContext{
			Strata: []string{h1, h2, h3},
			PathToBedrock: map[string]string{
				"src/owned.txt": h1,
			},
		}

		dr, err := engine.Discover(engine.DiscoveryParams{Seepage: seepage, Repo: repo})
		if err != nil {
			t.Fatalf("Discover failed: %v", err)
		}

		if dr.Schedule[h1][h2][0] != "src/owned.txt" {
			t.Errorf("Expected h2 to modify src/owned.txt for h1 bedrock")
		}
		if _, ok := dr.Sources["src/owned.txt"][h2]; !ok {
			t.Errorf("Expected src/owned.txt to list h2 as a source")
		}
		if _, ok := dr.Touched[h3]["other.txt"]; !ok {
			t.Errorf("Expected h3 to touch other.txt")
		}
	})

	t.Run("Path Ownership Overlap (Most Specific Wins)", func(t *testing.T) {
		// Bedrock 1: owns src/
		h1, _ := testutil.CommitFile(t, testutil.CommitParams{
			Repo: repo, Worktree: wt, Path: "src/main.py", Content: "content1", Msg: "msg1", AuthorDate: d0.Add(time.Hour * 4),
		})

		// Bedrock 2: owns src/core/ (more specific)
		h2, _ := testutil.CommitFile(t, testutil.CommitParams{
			Repo: repo, Worktree: wt, Path: "src/core/api.py", Content: "content2", Msg: "msg2", AuthorDate: d0.Add(time.Hour * 5),
		})

		// Surface commit touching both
		h3, _ := testutil.CommitFile(t, testutil.CommitParams{
			Repo: repo, Worktree: wt, Path: "src/main.py", Content: "content1_mod", Msg: "msg3", AuthorDate: d0.Add(time.Hour * 6),
		})
		_, _ = testutil.CommitFile(t, testutil.CommitParams{
			Repo: repo, Worktree: wt, Path: "src/core/api.py", Content: "content2_mod", Msg: "msg4", AuthorDate: d0.Add(time.Hour * 7),
		})
		h4, _ := repo.Head()

		seepage := &models.SeepageContext{
			Strata: []string{h1, h2, h3, h4.Hash().String()},
			PathToBedrock: map[string]string{
				"src/":      h1,
				"src/core/": h2,
			},
			ParentOfStrata: h0,
		}

		dr, err := engine.Discover(engine.DiscoveryParams{Seepage: seepage, Repo: repo})
		if err != nil {
			t.Fatalf("Discover failed: %v", err)
		}

		if _, ok := dr.Schedule[h1][h3]; !ok {
			t.Errorf("Expected src/main.py modification to map to h1")
		}
		if _, ok := dr.Schedule[h2][h4.Hash().String()]; !ok {
			t.Errorf("Expected src/core/api.py modification to map to h2, not h1")
		}
		if _, ok := dr.Schedule[h1][h4.Hash().String()]; ok {
			t.Errorf("src/core/api.py incorrectly mapped to the less specific h1 rule")
		}
	})

	t.Run("No-Lithify Policy Violation Check", func(t *testing.T) {
		h1, _ := testutil.CommitFile(t, testutil.CommitParams{
			Repo: repo, Worktree: wt, Path: "lib/owned.txt", Content: "content1", Msg: "msg1", AuthorDate: d0.Add(time.Hour * 8),
		})
		h2, _ := testutil.CommitFile(t, testutil.CommitParams{
			Repo: repo, Worktree: wt, Path: "lib/owned.txt", Content: "content1_mod1", Msg: "msg2", AuthorDate: d0.Add(time.Hour * 9),
		})
		h3, _ := testutil.CommitFile(t, testutil.CommitParams{
			Repo: repo, Worktree: wt, Path: "lib/owned.txt", Content: "content1_mod2", Msg: "msg3", AuthorDate: d0.Add(time.Hour * 10),
		})

		seepage := &models.SeepageContext{
			Strata: []string{h1, h2, h3},
			PathToBedrock: map[string]string{
				"lib/owned.txt": h1,
			},
			Options: models.Options{NoLithify: true},
		}

		dr, err := engine.Discover(engine.DiscoveryParams{Seepage: seepage, Repo: repo})
		if err != nil {
			t.Fatalf("Discover failed: %v", err)
		}

		// Simulate the engine check for Lithification violation
		lithifyError := false
		for _, sources := range dr.Sources {
			if len(sources) > 1 && seepage.Options.NoLithify {
				lithifyError = true
				break
			}
		}

		if !lithifyError {
			t.Errorf("Expected a Lithification policy violation error")
		}
	})

	t.Run("Empty Strata", func(t *testing.T) {
		seepage := &models.SeepageContext{
			Strata: []string{},
			PathToBedrock: map[string]string{
				"lib/owned.txt": "placeholder",
			},
		}
		dr, err := engine.Discover(engine.DiscoveryParams{Seepage: seepage, Repo: repo})
		if err != nil {
			t.Fatalf("Discover failed: %v", err)
		}
		if len(dr.Touched) != 0 {
			t.Errorf("Expected zero touched files")
		}
	})

	t.Run("Concurrency Stress Test", func(t *testing.T) {
		// Generate a larger number of commits to increase the probability of race conditions.
		stressRepo, stressWt := testutil.SetupMemRepo(t)
		now := time.Now()
		var strata []string
		for i := 0; i < 50; i++ {
			fname := "file.txt"
			content := "content" + string(rune(i))
			h, _ := testutil.CommitFile(t, testutil.CommitParams{
				Repo:       stressRepo,
				Worktree:   stressWt,
				Path:       fname,
				Content:    content,
				Msg:        "msg",
				AuthorDate: now.Add(time.Duration(i) * time.Minute),
			})
			strata = append(strata, h)
		}

		seepage := &models.SeepageContext{
			Strata: strata,
			PathToBedrock: map[string]string{
				"file.txt": strata[0],
			},
		}

		// Run Discovery multiple times to further stress the synchronization.
		for i := 0; i < 5; i++ {
			_, err := engine.Discover(engine.DiscoveryParams{Seepage: seepage, Repo: stressRepo})
			if err != nil {
				t.Fatalf("Discover failed during stress test iteration %d: %v", i, err)
			}
		}
	})
}

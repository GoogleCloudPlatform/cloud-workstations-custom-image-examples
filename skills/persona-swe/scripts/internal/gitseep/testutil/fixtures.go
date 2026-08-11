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

package testutil

import (
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/models"
)

// SetupMinimalSeepageContext returns a basic context for testing.
func SetupMinimalSeepageContext() *models.SeepageContext {
	seepageCtx := &models.SeepageContext{
		RepoRoot:       "/root",
		OriginalHead:   "head123",
		BaseHash:       "base123",
		HashToSubject:  make(map[string]string),
		DateToHash:     make(map[string]string),
		DateToBranch:   make(map[string]string),
		HashToDate:     make(map[string]string),
		BranchToParent: make(map[string]string),
		ResolvedRules:  make(map[string][]string),
		PathToBedrock:  make(map[string]string),
	}
	seepageCtx.EnsureMatcher()
	return seepageCtx
}

// SetupLinearHistory creates a simple linear history in a mock repository.
func SetupLinearHistory(t *testing.T) (*git.Repository, *models.SeepageContext, []string) {
	t.Helper()
	repo, wt := SetupMemRepo(t)
	now := time.Now()

	h1, _ := CommitFile(t, CommitParams{
		Repo: repo, Worktree: wt, Path: "file1.txt", Content: "content1", Msg: "initial", AuthorDate: now.Add(-24 * time.Hour),
	})
	h2, _ := CommitFile(t, CommitParams{
		Repo: repo, Worktree: wt, Path: "file2.txt", Content: "content2", Msg: "second", AuthorDate: now.Add(-12 * time.Hour),
	})

	seepage := SetupMinimalSeepageContext()
	seepage.OriginalHead = h2
	seepage.BaseHash = h1
	seepage.ParentOfStrata = h1   // Base is parent of first stratum
	seepage.Strata = []string{h2} // Strata usually excludes the base
	seepage.HashToSubject[h1] = "initial"
	seepage.HashToSubject[h2] = "second"
	seepage.DateToHash["2026-01-01"] = h1
	seepage.DateToHash["2026-01-02"] = h2
	seepage.DateToBranch["2026-01-01"] = "main"
	seepage.DateToBranch["2026-01-02"] = "feat"
	seepage.HashToDate[h1] = "2026-01-01"
	seepage.HashToDate[h2] = "2026-01-02"
	seepage.BranchToParent["feat"] = "main"

	return repo, seepage, []string{h1, h2}
}

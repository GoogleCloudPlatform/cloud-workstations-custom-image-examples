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
	"context"
	"testing"
	"time"

	gitseepGit "github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/git"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/models"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/testutil"
)

func TestTaskExecution_Coverage(t *testing.T) {
	testutil.InitI18n()
	repo, seepage, _ := testutil.SetupLinearHistory(t)
	gitSvc := gitseepGit.NewTestGitService(repo)
	events := make(chan models.PipelineEvent, 100)
	confirm := make(chan []string, 1)

	// Initialize matcher to avoid nil dereference in Discover
	seepage.EnsureMatcher()

	tc := NewTaskContext(TaskContextParams{
		Ctx:         context.Background(),
		Repo:        repo,
		Git:         gitSvc,
		Shell:       &mockShellService{},
		Events:      events,
		ConfirmChan: confirm,
	})
	tc.seepage = seepage
	tc.surfaceErrChan <- nil // Pre-unblock validation

	t.Run("DiscoveryTask", func(t *testing.T) {
		task := discoveryTask{}
		err := task.Execute(tc)
		if err != nil {
			t.Errorf("DiscoveryTask failed: %v", err)
		}
		if tc.discoveryResult == nil {
			t.Error("DiscoveryTask did not set discoveryResult")
		}
	})

	t.Run("StratigraphyTask", func(t *testing.T) {
		task := stratigraphyTask{}
		err := task.Execute(tc)
		if err != nil {
			t.Errorf("StratigraphyTask failed: %v", err)
		}
		if len(tc.orderedBranches) == 0 {
			t.Error("StratigraphyTask did not set orderedBranches")
		}
	})

	t.Run("ReconstructionTask", func(t *testing.T) {
		task := reconstructionTask{}
		err := task.Execute(tc)
		if err != nil {
			t.Errorf("ReconstructionTask failed: %v", err)
		}
		if tc.linearHead == "" {
			t.Error("ReconstructionTask did not set linearHead")
		}
	})

	t.Run("SedimentationTask", func(t *testing.T) {
		task := sedimentationTask{}
		err := task.Execute(tc)
		if err != nil {
			t.Errorf("SedimentationTask failed: %v", err)
		}
		if tc.branchHeads == nil {
			t.Error("SedimentationTask did not set branchHeads")
		}
	})

	t.Run("ValidationTask", func(t *testing.T) {
		tc.seepage.Options.SkipPreCommit = true
		task := validationTask{}
		err := task.Execute(tc)
		if err != nil {
			t.Errorf("ValidationTask failed: %v", err)
		}
	})

	t.Run("FinalizationTask", func(t *testing.T) {
		tc.seepage.Options.DryRun = true
		task := finalizationTask{}
		err := task.Execute(tc)
		if err != nil {
			t.Errorf("FinalizationTask failed: %v", err)
		}
	})
}

func TestValidateSurface_Coverage(t *testing.T) {
	testutil.InitI18n()
	repo, wt := testutil.SetupMemRepo(t)
	gitSvc := gitseepGit.NewTestGitService(repo)
	now := time.Now()

	h1, _ := testutil.CommitFile(t, testutil.CommitParams{
		Repo: repo, Worktree: wt, Path: "file1.txt", Content: "content1", Msg: "msg1", AuthorDate: now,
	})

	t.Run("Surface check skipped", func(t *testing.T) {
		p := ValidationParams{Skip: true}
		_, err := ValidateSurface(p)
		if err != nil {
			t.Errorf("ValidateSurface failed when skipped: %v", err)
		}
	})

	t.Run("Surface check no pre-commit config", func(t *testing.T) {
		p := ValidationParams{
			Repo:         repo,
			GitSvc:       gitSvc,
			OriginalHead: h1,
			RepoRoot:     t.TempDir(), // No config here
		}
		_, err := ValidateSurface(p)
		if err != nil {
			t.Errorf("ValidateSurface failed without config: %v", err)
		}
	})
}

func TestTasks_Coverage(t *testing.T) {
	t.Run("Task metadata", func(t *testing.T) {
		tasks := []PipelineTask{
			contextResolutionTask{},
			discoveryTask{},
			stratigraphyTask{},
			reviewTask{},
			reconstructionTask{},
			sedimentationTask{},
			validationTask{},
			finalizationTask{},
		}
		for _, task := range tasks {
			if task.Name() == "" {
				t.Error("Task name should not be empty")
			}
			// Just verify it doesn't panic and returns a value
			_ = task.Phase()
		}
	})
}

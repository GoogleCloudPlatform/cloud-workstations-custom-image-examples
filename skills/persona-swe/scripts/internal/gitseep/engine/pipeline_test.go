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
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/config"
	gitseepGit "github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/git"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/models"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/testutil"
)

func TestHistoryPipeline_NoChangesSkipValidation(t *testing.T) {
	testutil.InitI18n()
	repoDir, h1, _, cleanupRepo := testutil.SetupTestRepoSystemGit(t)
	defer cleanupRepo()

	// 1. Create a mock pre-commit that tracks calls
	countFile := filepath.Join(repoDir, "count.txt")
	testutil.SetupMockPreCommit(t, 0, countFile)

	repo, _ := git.PlainOpen(repoDir)

	// 2. Setup config where no files need migration (everything is already in order)
	cfg := &config.GitSeepConfig{
		Rules: []config.Rule{}, // Empty rules means no migrations predicted
	}

	events := make(chan models.PipelineEvent, 100)
	confirm := make(chan []string, 1)
	confirm <- nil // No exclusions

	reporter := &HeadlessReporter{Quiet: true}
	pipeline := NewHistoryPipeline(cfg, models.Options{BaseCommit: h1}, repo,
		WithGitService(gitseepGit.NewTestGitService(repo)),
		WithEvents(events),
		WithConfirmation(confirm),
		WithReporter(reporter),
	)

	// Use a goroutine to drain events and log them
	done := make(chan bool)
	go func() {
		for ev := range events {
			t.Logf("Event: Phase=%d, Type=%d, Msg=%s", ev.Phase, ev.Type, ev.Message)
		}
		done <- true
	}()

	pipeline.Execute()
	<-done

	// 3. Verify validation was skipped
	// Surface validation runs once on h2 (OriginalHead).
	// If historical validation were to run, it would run on h1.
	// But since hasChanges is false, h1 should be skipped.

	data, _ := os.ReadFile(countFile)
	count := strings.Count(string(data), "ran")

	// Expect 1 run (Surface Check) instead of 2 (h1 + h2)
	if count != 1 {
		t.Errorf("Expected exactly 1 pre-commit run (Surface Check), but ran %d times", count)
	}
}

type mockReporter struct {
	summaryReported bool
	errorReported   bool
}

func (r *mockReporter) ReportLog(msg string)                                  {}
func (r *mockReporter) ReportProgress(phase models.PipelinePhase, msg string) {}
func (r *mockReporter) ReportSummary(seepage *models.SeepageContext, dr *models.DiscoveryResult) {
	r.summaryReported = true
}
func (r *mockReporter) ReportFailure(ev models.PipelineEvent) {
	r.errorReported = true
}

func TestPipeline_Execute_Abort(t *testing.T) {
	p := &HistoryPipeline{
		events:      make(chan models.PipelineEvent, 100),
		confirmChan: make(chan []string, 1),
		reporter:    &mockReporter{},
	}

	tc := &TaskContext{
		events:      p.events,
		confirmChan: p.confirmChan,
	}

	err := p.runTask(tc, abortingTask{})
	if err != nil {
		t.Errorf("runTask should consume ErrPipelineAborted and return nil, got %v", err)
	}
}

type abortingTask struct{}

func (t abortingTask) Name() string                { return "Abort" }
func (t abortingTask) Phase() models.PipelinePhase { return models.PhaseReview }
func (t abortingTask) Execute(tc *TaskContext) error {
	return models.ErrPipelineAborted
}

func (p *HistoryPipeline) runTask(tc *TaskContext, task PipelineTask) error {
	if err := task.Execute(tc); err != nil {
		if models.IsAborted(err) {
			p.reporter.ReportSummary(tc.seepage, tc.discoveryResult)
			return nil
		}
		return err
	}
	return nil
}

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

	"github.com/go-git/go-git/v5"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/config"
	gitseepGit "github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/git"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/models"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/sys"
)

// HistoryPipeline encapsulates the full 8-phase geological reconstruction lifecycle.
type HistoryPipeline struct {
	cfg         *config.GitSeepConfig
	opts        models.Options
	repo        *git.Repository
	git         gitseepGit.GitService
	shell       sys.ShellService
	events      chan<- models.PipelineEvent
	confirmChan <-chan []string
	reporter    Reporter
}

// PipelineOption defines a functional option for configuring the HistoryPipeline.
type PipelineOption func(*HistoryPipeline)

// WithShellService sets the system shell service provider.
func WithShellService(shell sys.ShellService) PipelineOption {
	return func(p *HistoryPipeline) {
		p.shell = shell
	}
}

// WithGitService sets the Git service provider.
func WithGitService(git gitseepGit.GitService) PipelineOption {
	return func(p *HistoryPipeline) {
		p.git = git
	}
}

// WithEvents sets the pipeline events channel.
func WithEvents(events chan<- models.PipelineEvent) PipelineOption {
	return func(p *HistoryPipeline) {
		p.events = events
	}
}

// WithConfirmation sets the channel for receiving user approval/rejections.
func WithConfirmation(confirmChan <-chan []string) PipelineOption {
	return func(p *HistoryPipeline) {
		p.confirmChan = confirmChan
	}
}

// WithReporter sets the reporter for delivering pipeline updates.
func WithReporter(reporter Reporter) PipelineOption {
	return func(p *HistoryPipeline) {
		p.reporter = reporter
	}
}

// NewHistoryPipeline creates a new pipeline instance.
func NewHistoryPipeline(cfg *config.GitSeepConfig, opts models.Options, repo *git.Repository, options ...PipelineOption) *HistoryPipeline {
	p := &HistoryPipeline{
		cfg:   cfg,
		opts:  opts,
		repo:  repo,
		shell: &sys.RealShellService{}, // Default
	}

	for _, opt := range options {
		opt(p)
	}

	return p
}

// Execute runs the full pipeline sequentially, coordinating parallel validation and user reviews.
func (p *HistoryPipeline) Execute() {
	defer close(p.events)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tc := NewTaskContext(TaskContextParams{
		Ctx:         ctx,
		Cfg:         p.cfg,
		Opts:        p.opts,
		Repo:        p.repo,
		Git:         p.git,
		Shell:       p.shell,
		Events:      p.events,
		ConfirmChan: p.confirmChan,
	})

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
		if err := task.Execute(tc); err != nil {
			if models.IsAborted(err) {
				p.reporter.ReportSummary(tc.seepage, tc.discoveryResult)
				return
			}

			// If the error already carries a rich event, use it
			if re, ok := err.(*models.RichPipelineError); ok {
				p.reporter.ReportFailure(re.Event)
				return
			}

			failureEv := models.PipelineEvent{
				Phase:   task.Phase(),
				Type:    models.EventFailure,
				Message: err.Error(),
			}
			p.events <- failureEv
			p.reporter.ReportFailure(failureEv)
			return
		}
	}

	// Final Report
	p.reporter.ReportSummary(tc.seepage, tc.discoveryResult)
}

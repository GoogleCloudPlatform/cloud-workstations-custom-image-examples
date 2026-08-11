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
	"sync"

	"github.com/go-git/go-git/v5"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/config"
	gitseepGit "github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/git"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/models"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/sys"
)

// TaskContext carries the evolving state of a geological reconstruction pipeline.
type TaskContext struct {
	ctx         context.Context
	cfg         *config.GitSeepConfig
	opts        models.Options
	repo        *git.Repository
	git         gitseepGit.GitService
	shell       sys.ShellService
	repoMu      sync.Mutex // Ensures thread-safe access to the non-thread-safe go-git instance
	events      chan<- models.PipelineEvent
	confirmChan <-chan []string

	// Evolving state
	seepage         *models.SeepageContext
	discoveryResult *models.DiscoveryResult
	orderedBranches []string
	linearCommits   map[string]string
	linearHead      string
	branchHeads     map[string]string
	linearHashes    []string
	surfaceErrChan  chan error
}

// TaskContextParams defines the dependencies and configuration required to initialize a TaskContext.
type TaskContextParams struct {
	Ctx         context.Context
	Cfg         *config.GitSeepConfig
	Opts        models.Options
	Repo        *git.Repository
	Git         gitseepGit.GitService
	Shell       sys.ShellService
	Events      chan<- models.PipelineEvent
	ConfirmChan <-chan []string
}

// NewTaskContext creates a new context for the geological reconstruction pipeline.
func NewTaskContext(p TaskContextParams) *TaskContext {
	return &TaskContext{
		ctx:            p.Ctx,
		cfg:            p.Cfg,
		opts:           p.Opts,
		repo:           p.Repo,
		git:            p.Git,
		shell:          p.Shell,
		events:         p.Events,
		confirmChan:    p.ConfirmChan,
		surfaceErrChan: make(chan error, 1),
	}
}

// StartBackgroundSurfaceValidation starts the workspace validation in a separate goroutine.
func (tc *TaskContext) StartBackgroundSurfaceValidation() {
	go func() {
		select {
		case <-tc.ctx.Done():
			tc.surfaceErrChan <- tc.ctx.Err()
			return
		default:
			_, err := ValidateSurface(ValidationParams{
				RepoRoot:     tc.seepage.RepoRoot,
				Repo:         tc.repo,
				OriginalHead: tc.seepage.OriginalHead,
				Skip:         tc.seepage.Options.SkipPreCommit,
				Events:       tc.events,
				RepoMu:       &tc.repoMu,
				GitSvc:       tc.git,
				Shell:        tc.shell,
			})
			tc.surfaceErrChan <- err
		}
	}()
}

// WaitSurfaceValidation blocks until the background surface validation completes.
func (tc *TaskContext) WaitSurfaceValidation() error {
	return <-tc.surfaceErrChan
}

// PipelineTask defines a single atomic phase of the GitSeep reconstruction lifecycle.
type PipelineTask interface {
	Name() string
	Phase() models.PipelinePhase
	Execute(tc *TaskContext) error
}

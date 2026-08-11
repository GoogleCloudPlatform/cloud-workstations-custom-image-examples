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
	"fmt"

	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/i18n"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/logger"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/models"
)

// contextResolutionTask handles the initial parsing of rules and repository context.
type contextResolutionTask struct{}

func (t contextResolutionTask) Name() string                { return "Context Resolution" }
func (t contextResolutionTask) Phase() models.PipelinePhase { return models.PhaseDiscovery }
func (t contextResolutionTask) Execute(tc *TaskContext) error {
	l := logger.WithPhase(logger.IconSearch, "Discovery")
	l.Info("Resolving geological strata...")
	seepage, err := NewContext(tc.git, tc.cfg, tc.opts)
	if err != nil {
		return fmt.Errorf("context resolution failed: %w", err)
	}
	tc.seepage = seepage
	tc.events <- models.PipelineEvent{Type: models.EventContextResolved, Payload: seepage}
	l.Success("Geological strata resolved.")
	return nil
}

// discoveryTask handles scanning the history and starting surface validation.
type discoveryTask struct{}

func (t discoveryTask) Name() string                { return "Discovery" }
func (t discoveryTask) Phase() models.PipelinePhase { return models.PhaseDiscovery }
func (t discoveryTask) Execute(tc *TaskContext) error {
	tc.StartBackgroundSurfaceValidation()

	dr, err := Discover(DiscoveryParams{
		Seepage: tc.seepage,
		Repo:    tc.repo,
		Events:  tc.events,
		RepoMu:  &tc.repoMu,
	})
	if err != nil {
		return err
	}
	tc.discoveryResult = dr
	return nil
}

// stratigraphyTask handles dependency analysis and conflict prediction.
type stratigraphyTask struct{}

func (t stratigraphyTask) Name() string                { return "Stratigraphy" }
func (t stratigraphyTask) Phase() models.PipelinePhase { return models.PhaseStratigraphy }
func (t stratigraphyTask) Execute(tc *TaskContext) error {
	l := logger.WithPhase(logger.IconBedrock, "Stratigraphy")
	l.Info("Analyzing geological dependencies...")
	orderedBranches, err := ValidateStratigraphyAndPredictConflicts(tc.seepage, tc.discoveryResult)
	if err != nil {
		return err
	}
	tc.orderedBranches = orderedBranches
	l.Success("Stratigraphy analysis complete.")
	return nil
}

// reviewTask handles user confirmation and early exit short-circuits.
type reviewTask struct{}

func (t reviewTask) Name() string                { return "Review" }
func (t reviewTask) Phase() models.PipelinePhase { return models.PhaseReview }
func (t reviewTask) Execute(tc *TaskContext) error {
	l := logger.WithPhase(logger.IconReview, "Review")
	l.Info("Waiting for user approval of file migrations...")
	tc.events <- models.PipelineEvent{Phase: models.PhaseReview, Type: models.EventStart}
	excludedFiles := <-tc.confirmChan
	tc.events <- models.PipelineEvent{Phase: models.PhaseReview, Type: models.EventSuccess}

	if len(excludedFiles) > 0 {
		tc.seepage.ExcludedPaths = make(map[string]bool)
		for _, f := range excludedFiles {
			tc.seepage.ExcludedPaths[f] = true
		}
		ApplyExclusions(tc.discoveryResult, excludedFiles)
	}

	if !tc.discoveryResult.HasMigrations() {
		l.Success("No geological changes approved. Finishing.")
		// Wait for surface check to finish before ending the pipeline
		_ = tc.WaitSurfaceValidation()
		return models.ErrPipelineAborted
	}
	return nil
}

// reconstructionTask handles the mathematical projection of files into history.
type reconstructionTask struct{}

func (t reconstructionTask) Name() string                { return "Execution" }
func (t reconstructionTask) Phase() models.PipelinePhase { return models.PhaseExecution }
func (t reconstructionTask) Execute(tc *TaskContext) error {
	linearCommits, linearHead, err := ReconstructHistory(ReconstructionParams{
		Seepage: tc.seepage,
		Result:  tc.discoveryResult,
		Repo:    tc.repo,
		Events:  tc.events,
	})
	if err != nil {
		return err
	}
	tc.linearCommits = linearCommits
	tc.linearHead = linearHead
	tc.seepage.OriginalToReconstructed = linearCommits
	tc.linearHashes = tc.seepage.GetReconstructedStrata()

	return nil
}

// sedimentationTask handles updating the feature branch DAG.
type sedimentationTask struct{}

func (t sedimentationTask) Name() string                { return "Sedimentation" }
func (t sedimentationTask) Phase() models.PipelinePhase { return models.PhaseSedimentation }
func (t sedimentationTask) Execute(tc *TaskContext) error {
	branchHeads, err := PerformSedimentation(SedimentationParams{
		Seepage:         tc.seepage,
		Result:          tc.discoveryResult,
		Repo:            tc.repo,
		OrderedBranches: tc.orderedBranches,
		LinearCommits:   tc.linearCommits,
		Events:          tc.events,
	})
	if err != nil {
		return err
	}
	tc.branchHeads = branchHeads
	return nil
}

// validationTask handles wait for surface check and running historical validation.
type validationTask struct{}

func (t validationTask) Name() string                { return "Validation" }
func (t validationTask) Phase() models.PipelinePhase { return models.PhaseValidation }
func (t validationTask) Execute(tc *TaskContext) error {
	// Wait for Surface Validation
	if err := tc.WaitSurfaceValidation(); err != nil {
		return err
	}

	// Historical Validation
	return ValidateHistory(ValidationParams{
		Seepage:      tc.seepage,
		RepoRoot:     tc.seepage.RepoRoot,
		Repo:         tc.repo,
		OriginalHead: tc.seepage.OriginalHead,
		LinearHashes: tc.linearHashes,
		BranchHeads:  tc.branchHeads,
		Skip:         tc.seepage.Options.SkipPreCommit,
		SkipHash:     tc.seepage.BaseHash,
		Events:       tc.events,
		RepoMu:       &tc.repoMu,
		Shell:        tc.shell,
	})
}

// finalizationTask handles the final safe pointer move.
type finalizationTask struct{}

func (t finalizationTask) Name() string                { return "Finalization" }
func (t finalizationTask) Phase() models.PipelinePhase { return models.PhaseFinalization }
func (t finalizationTask) Execute(tc *TaskContext) error {
	l := logger.WithPhase(logger.IconFinalize, "Finalize")
	tc.events <- models.PipelineEvent{Phase: models.PhaseFinalization, Type: models.EventStart}
	targetBranch, err := FinalizeReferences(tc.seepage, tc.git, tc.linearHead, tc.branchHeads)
	if err != nil {
		return err
	}
	tc.events <- models.PipelineEvent{Phase: models.PhaseFinalization, Type: models.EventSuccess}

	if tc.seepage.Options.StageOnly {
		l.Success("%s", i18n.HistoryStaged(targetBranch))
	} else {
		l.Success("Geological history reconstructed and finalized successfully.")
	}
	return nil
}

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
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	gitseepGit "github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/git"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/gitutil"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/i18n"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/logger"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/models"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/sys"
)

// ValidationParams carries the configuration and state for geological validation.
type ValidationParams struct {
	RepoRoot     string
	Repo         *git.Repository
	GitSvc       gitseepGit.GitService
	RepoMu       *sync.Mutex
	Events       chan<- models.PipelineEvent
	OriginalHead string
	Skip         bool

	// History-specific
	Seepage      *models.SeepageContext
	LinearHashes []string
	BranchHeads  map[string]string
	SkipHash     string
	Shell        sys.ShellService
}

// ValidateSurface runs pre-commit checks on the current HEAD, but only for files that have changed
// in the HEAD commit itself.
func ValidateSurface(p ValidationParams) (string, error) {
	if p.Events != nil {
		p.Events <- models.PipelineEvent{Phase: models.PhaseSurface, Type: models.EventStart}
	}
	if p.Skip {
		if p.Events != nil {
			p.Events <- models.PipelineEvent{Phase: models.PhaseSurface, Type: models.EventSkipped}
		}
		return "", nil
	}
	l := logger.WithPhase(logger.IconLab, "Surface Check")
	l.Info("Validating current workspace (HEAD)...")

	repoRoot := p.RepoRoot
	if repoRoot == "" {
		repoRoot = "."
	}

	if p.RepoMu != nil {
		p.RepoMu.Lock()
	}
	// 1. Identify changed files in the HEAD commit (Per-Stratum logic)
	headCommit, _ := p.Repo.CommitObject(plumbing.NewHash(p.OriginalHead))
	var commitDate, commitMsg string
	var parentHash string
	if headCommit != nil {
		commitDate = headCommit.Committer.When.Format("2006-01-02 15:04")
		commitMsg = strings.Split(headCommit.Message, "\n")[0]
		if headCommit.NumParents() > 0 {
			parentHash = headCommit.ParentHashes[0].String()
		}
	}
	if p.RepoMu != nil {
		p.RepoMu.Unlock()
	}

	var changedFiles []string
	var err error
	if parentHash != "" {
		if p.RepoMu != nil {
			p.RepoMu.Lock()
		}
		changedFiles, err = p.GitSvc.GetChangedFiles(parentHash, p.OriginalHead)
		if p.RepoMu != nil {
			p.RepoMu.Unlock()
		}
		if err != nil {
			return "", err
		}
	} else if headCommit != nil {
		// Root commit, get all files
		if p.RepoMu != nil {
			p.RepoMu.Lock()
		}
		tree, _ := headCommit.Tree()
		filesMap, _ := gitutil.GetAllEntries(p.Repo, tree)
		if p.RepoMu != nil {
			p.RepoMu.Unlock()
		}
		for f := range filesMap {
			changedFiles = append(changedFiles, f)
		}
	}

	if len(changedFiles) == 0 {
		l.Success("No changes detected on surface (HEAD).")
		if p.Events != nil {
			p.Events <- models.PipelineEvent{Phase: models.PhaseSurface, Type: models.EventSuccess}
		}
		return "", nil
	}

	// 2. Pre-flight check: pre-commit config must exist at root
	configPath := filepath.Join(repoRoot, ".pre-commit-config.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if p.Events != nil {
			p.Events <- models.PipelineEvent{Phase: models.PhaseSurface, Type: models.EventSkipped, Message: i18n.T("error_precommit_not_configured")}
		}
		return "", nil
	}

	if p.Events != nil {
		p.Events <- models.PipelineEvent{Phase: models.PhaseSurface, Type: models.EventProgress, Message: i18n.ValidatingSurface(p.OriginalHead)}
	}

	// 3. Initialize Task Runner
	runner := NewPreCommitRunner(PreCommitParams{
		RepoRoot:     repoRoot,
		Repo:         p.Repo,
		OriginalHead: p.OriginalHead,
		Events:       p.Events,
		Shell:        p.Shell,
	})
	if err := runner.Setup(); err != nil {
		if p.Events != nil {
			p.Events <- models.PipelineEvent{Phase: models.PhaseSurface, Type: models.EventFailure, Message: err.Error()}
		}
		return "", err
	}

	// 4. Run tests for the surface stratum
	logOut, _, skipped, troubleshoot, err := runner.Run(p.OriginalHead, "", models.PhaseSurface)
	if err != nil {
		logger.Error("%s", i18n.T("error_surface_validation_failed"))
		ev := models.PipelineEvent{
			Phase:        models.PhaseSurface,
			Type:         models.EventFailure,
			Commit:       p.OriginalHead,
			CommitDate:   commitDate,
			CommitMsg:    commitMsg,
			Message:      err.Error(),
			LogOutput:    logOut,
			Files:        changedFiles,
			TempDir:      runner.WorktreeDir(),
			Troubleshoot: troubleshoot,
		}
		if p.Events != nil {
			p.Events <- ev
		}
		return logOut, &models.RichPipelineError{Event: ev}
	}

	if skipped {
		if p.Events != nil {
			p.Events <- models.PipelineEvent{Phase: models.PhaseSurface, Type: models.EventSkipped, Message: i18n.T("error_precommit_not_configured")}
		}
		_ = runner.Cleanup()
		return "", nil
	}

	_ = runner.Cleanup()
	l.Success("%s", i18n.T("surface_validation_success"))
	if p.Events != nil {
		p.Events <- models.PipelineEvent{Phase: models.PhaseSurface, Type: models.EventSuccess}
	}

	return logOut, nil
}

// ValidateHistory creates an isolated git worktree and runs pre-commit tests on reconstructed commits.
func ValidateHistory(p ValidationParams) error {
	if p.Events != nil {
		p.Events <- models.PipelineEvent{
			Phase: models.PhaseValidation,
			Type:  models.EventStart,
			Payload: map[string]interface{}{
				"heads":  p.BranchHeads,
				"linear": p.LinearHashes,
			},
		}
	}
	l := logger.WithPhase(logger.IconLab, "Validation")
	l.Info("Validating reconstructed geological strata...")
	if p.Skip {
		if p.Events != nil {
			p.Events <- models.PipelineEvent{Phase: models.PhaseValidation, Type: models.EventSkipped}
		}
		return nil
	}

	repoRoot := p.RepoRoot
	if repoRoot == "" {
		repoRoot = "."
	}

	// Pre-flight check: if pre-commit isn't installed, don't bother.
	if _, err := exec.LookPath("pre-commit"); err != nil {
		if p.Events != nil {
			p.Events <- models.PipelineEvent{Phase: models.PhaseValidation, Type: models.EventSkipped, Message: i18n.T("error_precommit_not_found")}
		}
		return nil
	}

	// Pre-flight check: if HEAD doesn't have a config, we can't run evergreen validation.
	configPath := filepath.Join(repoRoot, ".pre-commit-config.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if p.Events != nil {
			p.Events <- models.PipelineEvent{Phase: models.PhaseValidation, Type: models.EventSkipped, Message: i18n.T("error_precommit_not_configured")}
		}
		return nil
	}

	// 1. Initialize Task Runner
	runner := NewPreCommitRunner(PreCommitParams{
		RepoRoot:     repoRoot,
		Repo:         p.Repo,
		OriginalHead: p.OriginalHead,
		Events:       p.Events,
		Shell:        p.Shell,
	})
	if err := runner.Setup(); err != nil {
		if p.Events != nil {
			p.Events <- models.PipelineEvent{Phase: models.PhaseValidation, Type: models.EventFailure, Message: err.Error()}
		}
		return err
	}

	success := false
	defer func() {
		if success {
			_ = runner.Cleanup()
		}
	}()

	validatedHashes := make(map[string]bool)

	// Reverse mapping for fast resolution: reconstructed -> original
	newToOld := make(map[string]string)
	for oldH, newH := range p.Seepage.OriginalToReconstructed {
		newToOld[newH] = oldH
	}

	// Testing Feature Branches - Newest to Oldest
	var featureBranchNames []string
	for name := range p.BranchHeads {
		featureBranchNames = append(featureBranchNames, name)
	}
	sort.Slice(featureBranchNames, func(i, j int) bool {
		var di, dj string
		for d, b := range p.Seepage.DateToBranch {
			if b == featureBranchNames[i] {
				di = d
			}
			if b == featureBranchNames[j] {
				dj = d
			}
		}
		return models.ParseRuleDate(di).After(models.ParseRuleDate(dj))
	})

	totalBranches := len(featureBranchNames)
	for i, branchName := range featureBranchNames {
		hash := p.BranchHeads[branchName]
		if validatedHashes[hash] {
			if p.Events != nil {
				p.Events <- models.PipelineEvent{Phase: models.PhaseValidation, Type: models.EventSuccess, Commit: hash, Branch: branchName, Message: i18n.T("already_validated")}
			}
			continue
		}

		if hash == p.SkipHash {
			if p.Events != nil {
				p.Events <- models.PipelineEvent{Phase: models.PhaseValidation, Type: models.EventSkipped, Commit: hash, Branch: branchName, Message: i18n.T("base_ref_skipped")}
			}
			validatedHashes[hash] = true
			continue
		}

		// SHARED HISTORY OPTIMIZATION
		if origH, ok := newToOld[hash]; ok && origH == hash {
			if p.Events != nil {
				p.Events <- models.PipelineEvent{Phase: models.PhaseValidation, Type: models.EventSuccess, Commit: hash, Branch: branchName, Message: i18n.T("already_sedimented")}
			}
			validatedHashes[hash] = true
			continue
		}

		if p.Events != nil {
			p.Events <- models.PipelineEvent{
				Phase:   models.PhaseValidation,
				Type:    models.EventProgress,
				Commit:  hash,
				Branch:  branchName,
				Message: i18n.ValidatingFeatureBranch(i+1, totalBranches, hash, branchName),
			}
		}
		logOut, files, skipped, troubleshoot, err := runner.Run(hash, branchName, models.PhaseValidation)
		if err != nil {
			origH := newToOld[hash]
			if origH == "" {
				origH = hash
			}
			ev := models.PipelineEvent{
				Phase:        models.PhaseValidation,
				Type:         models.EventFailure,
				Commit:       hash,
				CommitDate:   p.Seepage.HashToDate[origH],
				CommitMsg:    p.Seepage.HashToSubject[origH],
				Branch:       branchName,
				Message:      err.Error(),
				LogOutput:    logOut,
				TempDir:      runner.WorktreeDir(),
				Files:        files,
				Troubleshoot: troubleshoot,
			}
			if p.Events != nil {
				p.Events <- ev
			}
			return &models.RichPipelineError{Event: ev}
		}
		validatedHashes[hash] = true
		if p.Events != nil {
			if skipped {
				p.Events <- models.PipelineEvent{Phase: models.PhaseValidation, Type: models.EventSkipped, Commit: hash, Branch: branchName, Message: i18n.ErrorPrecommitMissing(hash)}
			} else {
				p.Events <- models.PipelineEvent{Phase: models.PhaseValidation, Type: models.EventSuccess, Commit: hash, Branch: branchName}
			}
		}
	}

	// Testing Linear History - Newest to Oldest (Tips to Bedrock)
	for i := len(p.LinearHashes) - 1; i >= 0; i-- {
		hash := p.LinearHashes[i]
		if validatedHashes[hash] {
			continue
		}

		if hash == p.SkipHash {
			if p.Events != nil {
				p.Events <- models.PipelineEvent{Phase: models.PhaseValidation, Type: models.EventSkipped, Commit: hash, Message: i18n.T("base_ref_skipped")}
			}
			validatedHashes[hash] = true
			continue
		}

		// SHARED HISTORY OPTIMIZATION
		if origH, ok := newToOld[hash]; ok && origH == hash {
			for j := i; j >= 0; j-- {
				h := p.LinearHashes[j]
				if !validatedHashes[h] {
					if p.Events != nil {
						p.Events <- models.PipelineEvent{Phase: models.PhaseValidation, Type: models.EventSuccess, Commit: h, Message: i18n.T("already_sedimented")}
					}
					validatedHashes[h] = true
				}
			}
			break
		}

		totalLinear := len(p.LinearHashes)
		if p.Events != nil {
			p.Events <- models.PipelineEvent{
				Phase:   models.PhaseValidation,
				Type:    models.EventProgress,
				Commit:  hash,
				Message: i18n.ValidatingLinearStratum(totalLinear-i, totalLinear, hash),
			}
		}
		logOut, files, skipped, troubleshoot, err := runner.Run(hash, "", models.PhaseValidation)
		if err != nil {
			origH := newToOld[hash]
			if origH == "" {
				origH = hash
			}
			ev := models.PipelineEvent{
				Phase:        models.PhaseValidation,
				Type:         models.EventFailure,
				Commit:       hash,
				CommitDate:   p.Seepage.HashToDate[origH],
				CommitMsg:    p.Seepage.HashToSubject[origH],
				Message:      err.Error(),
				LogOutput:    logOut,
				TempDir:      runner.WorktreeDir(),
				Files:        files,
				Troubleshoot: troubleshoot,
			}
			if p.Events != nil {
				p.Events <- ev
			}
			return &models.RichPipelineError{Event: ev}
		}
		validatedHashes[hash] = true
		if p.Events != nil {
			if skipped {
				p.Events <- models.PipelineEvent{Phase: models.PhaseValidation, Type: models.EventSkipped, Commit: hash, Message: i18n.ErrorPrecommitMissing(hash)}
			} else {
				p.Events <- models.PipelineEvent{Phase: models.PhaseValidation, Type: models.EventSuccess, Commit: hash}
			}
		}
	}

	success = true
	l.Success("History validation completed.")
	if p.Events != nil {
		p.Events <- models.PipelineEvent{Phase: models.PhaseValidation, Type: models.EventSuccess}
	}
	return nil
}

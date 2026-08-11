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
	"sort"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/gitutil"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/i18n"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/logger"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/models"
)

// ReconstructionParams carries the configuration and state for geological projection.
type ReconstructionParams struct {
	Seepage  *models.SeepageContext
	Result   *models.DiscoveryResult
	Repo     *git.Repository
	Events   chan<- models.PipelineEvent
	Strategy ProjectionStrategy
}

// ProjectionStrategy defines how file modifications are projected into a new geological stratum.
type ProjectionStrategy interface {
	Project(p ReconstructionParams, commitHashStr string, parentHash plumbing.Hash) (plumbing.Hash, error)
}

// StateProjectionStrategy implements the default "State Mirror-and-Project" logic.
// It mirrors unmanaged files from original history and projects managed migrations.
type StateProjectionStrategy struct{}

func (s *StateProjectionStrategy) Project(p ReconstructionParams, commitHashStr string, parentHash plumbing.Hash) (plumbing.Hash, error) {
	origHash := plumbing.NewHash(commitHashStr)
	origCommit, err := p.Repo.CommitObject(origHash)
	if err != nil {
		return plumbing.ZeroHash, fmt.Errorf("%s", i18n.TF("error_commit_object", map[string]interface{}{"Hash": commitHashStr, "Error": err}))
	}

	origTree, err := origCommit.Tree()
	if err != nil {
		return plumbing.ZeroHash, fmt.Errorf("%s", i18n.TF("error_orig_tree", map[string]interface{}{"Hash": commitHashStr, "Error": err}))
	}

	// 1. Start with the tree of the NEW linear parent
	parentCommit, err := p.Repo.CommitObject(parentHash)
	if err != nil {
		return plumbing.ZeroHash, fmt.Errorf("failed to get parent commit %s: %w", parentHash.String(), err)
	}
	parentTree, _ := parentCommit.Tree()
	newEntries, _ := gitutil.GetAllEntries(p.Repo, parentTree)

	// 2. Mirror unmanaged files from original history
	origEntries, _ := gitutil.GetAllEntries(p.Repo, origTree)
	MirrorUnmanaged(p.Seepage, newEntries, origEntries, true)

	// 3. Project managed migrations into this stratum.
	// We MUST iterate over source commits in chronological order to ensure newer states win.
	if sourcesMap, ok := p.Result.Schedule[commitHashStr]; ok {
		var sortedSources []string
		for srcH := range sourcesMap {
			sortedSources = append(sortedSources, srcH)
		}
		sort.Slice(sortedSources, func(i, j int) bool {
			return models.IndexOf(p.Seepage.Strata, sortedSources[i]) < models.IndexOf(p.Seepage.Strata, sortedSources[j])
		})

		for _, srcH := range sortedSources {
			files := sourcesMap[srcH]
			srcCommit, _ := p.Repo.CommitObject(plumbing.NewHash(srcH))
			srcTree, _ := srcCommit.Tree()
			srcEntries, _ := gitutil.GetAllEntries(p.Repo, srcTree)

			for _, f := range files {
				if entry, ok := srcEntries[f]; ok {
					newEntries[f] = entry
				} else {
					delete(newEntries, f)
				}
			}
		}
	}

	// 4. Build the new tree
	newTreeHash, err := gitutil.BuildTree(p.Repo, newEntries)
	if err != nil {
		return plumbing.ZeroHash, err
	}

	// 5. Synthesize the reconstructed commit
	obj := p.Repo.Storer.NewEncodedObject()
	newCommit := &object.Commit{
		Author:       origCommit.Author,
		Committer:    origCommit.Committer,
		Message:      origCommit.Message,
		TreeHash:     newTreeHash,
		ParentHashes: []plumbing.Hash{parentHash},
	}

	err = newCommit.Encode(obj)
	if err != nil {
		return plumbing.ZeroHash, fmt.Errorf("failed to encode reconstructed commit: %w", err)
	}

	newHash, err := p.Repo.Storer.SetEncodedObject(obj)
	if err != nil {
		return plumbing.ZeroHash, fmt.Errorf("failed to store reconstructed commit: %w", err)
	}

	return newHash, nil
}

// ReconstructHistory performs the mathematical projection of file modifications
// into a new linear chronological stack, following the per-stratum principle.
func ReconstructHistory(p ReconstructionParams) (map[string]string, string, error) {
	if p.Events != nil {
		p.Events <- models.PipelineEvent{Phase: models.PhaseExecution, Type: models.EventStart}
	}
	l := logger.WithPhase(logger.IconExecution, "Execution")
	l.Info("Projecting file modifications into new geological strata...")

	if p.Strategy == nil {
		p.Strategy = &StateProjectionStrategy{}
	}

	linearCommits := make(map[string]string)
	currentParentHash := plumbing.NewHash(p.Seepage.ParentOfStrata)

	total := len(p.Seepage.Strata)
	for i, commitHashStr := range p.Seepage.Strata {
		if p.Events != nil {
			p.Events <- models.PipelineEvent{
				Phase:   models.PhaseExecution,
				Type:    models.EventProgress,
				Message: i18n.ReconstructingStrata(i+1, total, commitHashStr, p.Seepage.HashToSubject[commitHashStr]),
			}
		}

		newHash, err := p.Strategy.Project(p, commitHashStr, currentParentHash)
		if err != nil {
			return nil, "", err
		}

		linearCommits[commitHashStr] = newHash.String()
		currentParentHash = newHash
	}

	l.Success("History reconstruction completed.")
	if p.Events != nil {
		p.Events <- models.PipelineEvent{Phase: models.PhaseExecution, Type: models.EventSuccess}
	}
	return linearCommits, currentParentHash.String(), nil
}

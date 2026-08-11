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
	"sync"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/utils/merkletrie"

	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/i18n"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/logger"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/models"
)

// DiscoveryParams carries the configuration and state for geological discovery.
type DiscoveryParams struct {
	Seepage *models.SeepageContext
	Repo    *git.Repository
	Events  chan<- models.PipelineEvent
	RepoMu  *sync.Mutex
}

// Discover scans the history strata to identify which files were touched in which commits.
// History scanning is performed sequentially to ensure deterministic stability with the
// non-thread-safe go-git object resolution.
func Discover(p DiscoveryParams) (*models.DiscoveryResult, error) {
	p.Seepage.EnsureMatcher()
	if p.Events != nil {
		p.Events <- models.PipelineEvent{Phase: models.PhaseDiscovery, Type: models.EventStart}
	}
	l := logger.WithPhase(logger.IconSearch, "Discovery")
	l.Info("Scanning geological history for file changes...")

	result := &models.DiscoveryResult{
		Schedule: make(map[string]map[string][]string),
		Sources:  make(map[string]map[string]struct{}),
		Touched:  make(map[string]map[string]struct{}),
	}

	total := len(p.Seepage.Strata)
	for i, commitHash := range p.Seepage.Strata {
		if p.RepoMu != nil {
			p.RepoMu.Lock()
		}
		// Resolve commit and tree
		_, cTree, pTree, err := resolveCommitTrees(p.Repo, commitHash)
		if err != nil {
			if p.RepoMu != nil {
				p.RepoMu.Unlock()
			}
			return nil, err
		}

		// Perform an in-memory diff between the current commit and its parent.
		changes, err := object.DiffTree(pTree, cTree)
		if p.RepoMu != nil {
			p.RepoMu.Unlock()
		}
		if err != nil {
			return nil, fmt.Errorf("%s", i18n.TF("error_diff_tree", map[string]interface{}{"Hash": commitHash, "Error": err}))
		}

		// Extract list of touched files
		files := getTouchedFiles(changes)

		if result.Touched[commitHash] == nil {
			result.Touched[commitHash] = make(map[string]struct{})
		}

		for _, f := range files {
			targetBedrock, _ := p.Seepage.Matcher.ResolveTarget(f)

			result.Touched[commitHash][f] = struct{}{}

			if result.Sources[f] == nil {
				result.Sources[f] = make(map[string]struct{})
			}
			result.Sources[f][commitHash] = struct{}{}

			if targetBedrock != "" {
				if result.Schedule[targetBedrock] == nil {
					result.Schedule[targetBedrock] = make(map[string][]string)
				}
				result.Schedule[targetBedrock][commitHash] = append(result.Schedule[targetBedrock][commitHash], f)
			}
		}

		if p.Events != nil {
			p.Events <- models.PipelineEvent{
				Phase:   models.PhaseDiscovery,
				Type:    models.EventProgress,
				Message: i18n.ScanningHistory(i+1, total, commitHash, p.Seepage.HashToSubject[commitHash]),
			}
		}
	}

	l.Success("Discovery completed.")
	if p.Events != nil {
		p.Events <- models.PipelineEvent{Phase: models.PhaseDiscovery, Type: models.EventSuccess, Payload: result}
	}

	return result, nil
}

// resolveCommitTrees handles object resolution from the go-git repository.
// Note: This function is not thread-safe and must be called under a lock.
func resolveCommitTrees(repo *git.Repository, hash string) (*object.Commit, *object.Tree, *object.Tree, error) {
	c, err := repo.CommitObject(plumbing.NewHash(hash))
	if err != nil {
		return nil, nil, nil, fmt.Errorf("%s", i18n.TF("error_find_commit", map[string]interface{}{"Hash": hash, "Error": err}))
	}

	cTree, err := c.Tree()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("%s", i18n.TF("error_get_tree", map[string]interface{}{"Hash": hash, "Error": err}))
	}

	var pTree *object.Tree
	if c.NumParents() > 0 {
		p, err := c.Parent(0)
		if err == nil {
			pTree, _ = p.Tree()
		}
	}

	return c, cTree, pTree, nil
}

// getTouchedFiles extracts the file paths from a set of Git changes.
func getTouchedFiles(changes object.Changes) []string {
	var files []string
	for _, change := range changes {
		action, err := change.Action()
		if err != nil {
			continue
		}
		if action == merkletrie.Delete {
			files = append(files, change.From.Name)
		} else {
			files = append(files, change.To.Name)
		}
	}
	return files
}

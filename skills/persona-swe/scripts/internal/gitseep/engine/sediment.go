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
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/gitutil"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/i18n"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/logger"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/models"
)

// SedimentationParams carries the configuration and state for feature branch updates.
type SedimentationParams struct {
	Seepage         *models.SeepageContext
	Result          *models.DiscoveryResult
	Repo            *git.Repository
	OrderedBranches []string
	LinearCommits   map[string]string
	Events          chan<- models.PipelineEvent
}

// PerformSedimentation force-updates feature branches to their new reconstructed heads.
func PerformSedimentation(params SedimentationParams) (map[string]string, error) {
	if params.Events != nil {
		params.Events <- models.PipelineEvent{Phase: models.PhaseSedimentation, Type: models.EventStart}
	}
	l := logger.WithPhase(logger.IconBranch, "Sedimentation")
	l.Info("Sedimenting feature branches into reconstructed history...")

	// 1. Calculate Repository Universe
	// We build a set of all paths that ever existed in the original repository history
	// across the strata we are processing. This allows us to distinguish between
	// "Repo Files" (which follow main) and "Branch Files" (which are preserved).
	repoUniverse := make(map[string]bool)
	strataToScan := append([]string{}, params.Seepage.Strata...)
	if params.Seepage.ParentOfStrata != "" {
		strataToScan = append(strataToScan, params.Seepage.ParentOfStrata)
	}

	for _, h := range strataToScan {
		c, err := params.Repo.CommitObject(plumbing.NewHash(h))
		if err != nil {
			continue
		}
		t, _ := c.Tree()
		entries, _ := gitutil.GetAllEntries(params.Repo, t)
		for f := range entries {
			repoUniverse[f] = true
		}
	}

	result := make(map[string]string)
	branchHeads := make(map[string]plumbing.Hash)

	total := len(params.OrderedBranches)
	for i, branchName := range params.OrderedBranches {
		if params.Events != nil {
			params.Events <- models.PipelineEvent{
				Phase:   models.PhaseSedimentation,
				Type:    models.EventProgress,
				Message: i18n.SedimentingBranch(i+1, total, branchName),
			}
		}

		var bedrockH string
		for d, b := range params.Seepage.DateToBranch {
			if b == branchName {
				bedrockH = params.Seepage.DateToHash[d]
				break
			}
		}

		if bedrockH == "" {
			l.Warn("Could not identify bedrock for branch %s, skipping.", branchName)
			continue
		}

		var newBedrockHash plumbing.Hash
		if newH, ok := params.LinearCommits[bedrockH]; ok {
			newBedrockHash = plumbing.NewHash(newH)
		} else {
			l.Warn("Bedrock %s for branch %s not found in linear commits.", bedrockH[:7], branchName)
			newBedrockHash = plumbing.NewHash(bedrockH)
		}

		// Determine parent
		parentBranch := params.Seepage.BranchToParent[branchName]
		var parentHash plumbing.Hash
		if parentBranch != "" {
			if h, ok := branchHeads[parentBranch]; ok {
				parentHash = h
			} else {
				if ref, err := params.Repo.Reference(plumbing.NewBranchReferenceName(parentBranch), true); err == nil {
					parentHash = ref.Hash()
				}
			}
		}

		if parentHash == plumbing.ZeroHash {
			parentHash = resolveImplicitParent(params.Seepage, branchName, branchHeads, params.LinearCommits)
		}

		origHeadStr := params.Seepage.OriginalBranchHeads[branchName]
		// SHARED HISTORY OPTIMIZATION
		// If the branch tip was part of the original core history (stratum), we can reuse
		// the reconstructed linear commit directly. This ensures absolute parity
		// in both the tree AND the diff (file list).
		if origHeadStr != "" {
			if newH, ok := params.LinearCommits[origHeadStr]; ok {
				l.Info("Optimizing branch %s: reusing reconstructed commit %s", branchName, newH[:7])
				branchHeads[branchName] = plumbing.NewHash(newH)
				result[branchName] = newH
				continue
			}
		} else {
			// Initialize new branch at bedrock if it purely follows linear history.
			if parentHash == plumbing.ZeroHash || parentHash == newBedrockHash {
				l.Info("Initializing branch %s at bedrock %s", branchName, newBedrockHash.String()[:7])
				branchHeads[branchName] = newBedrockHash
				result[branchName] = newBedrockHash.String()
				continue
			}
		}

		// FULL RECONSTRUCTION
		// We use a "Patch Projection" strategy to ensure diff parity with 'main'
		// while correctly skipping intermediate strata.
		// 1. Calculate the patch of the current stratum in reconstructed linear history.
		// 2. Apply that patch to the current feature branch parent.
		// 3. Enforce global exclusions.
		// 4. Preserve branch-specific additions.
		if parentHash == plumbing.ZeroHash {
			parentHash = newBedrockHash
		}
		l.Info("Reconstructing branch %s onto parent %s", branchName, parentHash.String()[:7])

		parentCommit, err := params.Repo.CommitObject(parentHash)
		if err != nil {
			l.Error("Failed to get parent commit %s: %v", parentHash.String()[:7], err)
			branchHeads[branchName] = newBedrockHash
			result[branchName] = newBedrockHash.String()
			continue
		}
		parentTree, _ := parentCommit.Tree()
		newEntries, _ := gitutil.GetAllEntries(params.Repo, parentTree)

		// 1. Resolve Linear Patch (Change Set of main)
		// Find the reconstructed linear parent (predistratum)
		var linearParentHash plumbing.Hash
		for i, h := range params.Seepage.Strata {
			if h == bedrockH {
				if i > 0 {
					prevH := params.Seepage.Strata[i-1]
					if newPrevH, ok := params.LinearCommits[prevH]; ok {
						linearParentHash = plumbing.NewHash(newPrevH)
					}
				} else if params.Seepage.ParentOfStrata != "" {
					linearParentHash = plumbing.NewHash(params.Seepage.ParentOfStrata)
				}
				break
			}
		}

		linearParentCommit, _ := params.Repo.CommitObject(linearParentHash)
		var linearParentEntries map[string]object.TreeEntry
		if linearParentCommit != nil {
			pt, _ := linearParentCommit.Tree()
			linearParentEntries, _ = gitutil.GetAllEntries(params.Repo, pt)
		} else {
			linearParentEntries = make(map[string]object.TreeEntry)
		}

		bedrockCommit, _ := params.Repo.CommitObject(newBedrockHash)
		mainTree, _ := bedrockCommit.Tree()
		mainEntries, _ := gitutil.GetAllEntries(params.Repo, mainTree)

		// 2. Apply Patch
		// Additions and Modifications
		for f, entry := range mainEntries {
			prevEntry, wasInMainParent := linearParentEntries[f]
			if !wasInMainParent || prevEntry.Hash != entry.Hash {
				newEntries[f] = entry
			}
		}
		// Deletions
		for f := range linearParentEntries {
			if _, exists := mainEntries[f]; !exists {
				delete(newEntries, f)
			}
		}

		// 3. Enforce Global Exclusions
		// If a file is excluded in the UI, it MUST be gone from all reconstructed commits,
		// regardless of whether the patch deleted it.
		for f := range newEntries {
			if params.Seepage.ExcludedPaths != nil && params.Seepage.ExcludedPaths[f] {
				delete(newEntries, f)
			}
		}

		// 4. Branch Preservation: Carry over truly branch-specific content.
		// A file is branch-specific if it was NEVER part of the original repository history
		// AND it is not managed by any rules.
		if origHeadStr != "" {
			oldTipCommit, _ := params.Repo.CommitObject(plumbing.NewHash(origHeadStr))
			oldTipTree, _ := oldTipCommit.Tree()
			oldTipEntries, _ := gitutil.GetAllEntries(params.Repo, oldTipTree)

			// Sync unmanaged, non-repo files with the old tip.
			for f := range newEntries {
				if params.Seepage.ResolveBedrockForPath(f) == "" && !repoUniverse[f] {
					if _, exists := oldTipEntries[f]; !exists {
						delete(newEntries, f)
					}
				}
			}
			for f, entry := range oldTipEntries {
				if params.Seepage.ResolveBedrockForPath(f) == "" && !repoUniverse[f] {
					newEntries[f] = entry
				}
			}
		}
		// NOTE: If origHeadStr == "", the branch is new and simply inherits
		// all content (including branch-specific additions) from its parent.

		// METADATA ALIGNMENT: Inherit author/message from main to maximize hash stability.
		sourceAuthor := bedrockCommit.Author
		sourceCommitter := bedrockCommit.Committer
		sourceMsg := bedrockCommit.Message

		newTreeHash, err := gitutil.BuildTree(params.Repo, newEntries)
		if err != nil {
			l.Error("Failed to build tree for branch %s: %v", branchName, err)
			continue
		}

		// STABILITY: Preserve full signatures to maintain hash parity with linear history.
		newCommit := &object.Commit{
			Author:       sourceAuthor,
			Committer:    sourceCommitter,
			Message:      sourceMsg,
			TreeHash:     newTreeHash,
			ParentHashes: []plumbing.Hash{parentHash},
		}

		obj := params.Repo.Storer.NewEncodedObject()
		err = newCommit.Encode(obj)
		if err != nil {
			l.Error("Failed to encode commit for branch %s: %v", branchName, err)
			continue
		}
		newHeadHash, err := params.Repo.Storer.SetEncodedObject(obj)
		if err != nil {
			l.Error("Failed to store commit for branch %s: %v", branchName, err)
			continue
		}

		l.Info("Successfully sedimented branch %s -> %s", branchName, newHeadHash.String()[:7])
		branchHeads[branchName] = newHeadHash
		result[branchName] = newHeadHash.String()
	}

	l.Success("Branch sedimentation completed.")
	return result, nil
}

// resolveImplicitParent identifies the geologically most recent commit to use as a parent.
// It considers both already-sedimented feature branches and the reconstructed linear history (main).
func resolveImplicitParent(seepage *models.SeepageContext, branchName string, branchHeads map[string]plumbing.Hash, linearCommits map[string]string) plumbing.Hash {
	var date string
	for d, b := range seepage.DateToBranch {
		if b == branchName {
			date = d
			break
		}
	}

	targetDate := models.ParseRuleDate(date)
	var bestParent plumbing.Hash
	var bestDate time.Time

	// 1. Check other feature branches
	for d, b := range seepage.DateToBranch {
		if b == branchName {
			continue
		}
		dTime := models.ParseRuleDate(d)
		if dTime.Before(targetDate) && (bestDate.IsZero() || dTime.After(bestDate)) {
			if h, ok := branchHeads[b]; ok {
				bestDate = dTime
				bestParent = h
			}
		}
	}

	// 2. Check linear history (main strata)
	// We want to pick the latest linear stratum that is geologically BEFORE our current branch.
	for d, origH := range seepage.DateToHash {
		dTime := models.ParseRuleDate(d)
		if dTime.Before(targetDate) && (bestDate.IsZero() || dTime.After(bestDate)) {
			// Get the reconstructed version of this stratum
			if newH, ok := linearCommits[origH]; ok {
				bestDate = dTime
				bestParent = plumbing.NewHash(newH)
			}
		}
	}

	return bestParent
}

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

// Package git provides helpers for interacting with the Git repository using go-git.
package git

import (
	"fmt"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/i18n"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/models"
)

// HistoryResult carries the resolved geological history state.
type HistoryResult struct {
	ResolvedDates  map[string]string
	Strata         []string
	ParentOfStrata string
	BaseHash       string
	OriginalHead   string
	CurrentBranch  string
}

// GetSeepageHistory performs a high-performance single-pass traversal of Git history to resolve all bedrock dates
// and identify the linear stack of strata commits.
func GetSeepageHistory(gitSvc GitService, targetDates []string, baseRef string) (*HistoryResult, error) {
	repo := gitSvc.GetRepo()
	head, err := repo.Head()
	if err != nil {
		return nil, err
	}
	origHead := head.Hash().String()
	currentBranch, _ := gitSvc.GetCurrentBranch()

	// Resolve base commit hash if provided
	var baseHash string
	if baseRef != "" {
		h, err := gitSvc.ResolveRevision(baseRef)
		if err != nil {
			// Try as hash literal
			tempH := plumbing.NewHash(baseRef)
			h = tempH
		}
		baseHash = h.String()
	}

	// Prepare date resolution state
	resolvedDates := make(map[string]string)
	bestCommits := make(map[string]*object.Commit)
	parsedTargetDates := make(map[string]time.Time)
	for _, d := range targetDates {
		parsedTargetDates[d] = models.ParseRuleDate(d)
	}

	cIter, err := repo.Log(&git.LogOptions{From: head.Hash()})
	if err != nil {
		return nil, err
	}

	var fullReverseHistory []string
	var parentOfStrata string
	baseIdx := -1

	// Single pass iteration from HEAD downwards
	_ = cIter.ForEach(func(c *object.Commit) error {
		h := c.Hash.String()
		fullReverseHistory = append(fullReverseHistory, h)

		// 1. Resolve dates
		for dStr, targetDate := range parsedTargetDates {
			if c.Author.When.After(targetDate) || c.Author.When.Equal(targetDate) {
				if bestCommits[dStr] == nil || c.Author.When.Before(bestCommits[dStr].Author.When) {
					bestCommits[dStr] = c
				}
			}
		}

		// 2. Identify base/parent
		if baseHash != "" && h == baseHash {
			baseIdx = len(fullReverseHistory) - 1
			parentOfStrata = h // Anchor ONTO the base commit
			// We DON'T stop here because we might need to resolve older bedrock dates
		}

		return nil
	})

	// Truncate history to identify strata
	var reverseStrata []string
	if baseRef != "" {
		if baseIdx != -1 {
			// Strata includes everything AFTER the base commit up to HEAD
			reverseStrata = fullReverseHistory[:baseIdx]
		} else {
			return nil, fmt.Errorf("%s", i18n.TF("error_invalid_base", map[string]interface{}{"Base": baseRef, "Error": "not found in history"}))
		}
	} else {
		// If no base was provided, we use the oldest bedrock found as the anchor
		var oldestBedrock *object.Commit
		for _, c := range bestCommits {
			if oldestBedrock == nil || c.Author.When.Before(oldestBedrock.Author.When) {
				oldestBedrock = c
			}
		}

		if oldestBedrock != nil {
			oldestHash := oldestBedrock.Hash.String()
			idx := -1
			for i, h := range fullReverseHistory {
				if h == oldestHash {
					idx = i
					break
				}
			}
			if idx != -1 {
				reverseStrata = fullReverseHistory[:idx+1]
				if oldestBedrock.NumParents() > 0 {
					parentOfStrata = oldestBedrock.ParentHashes[0].String()
				}
			}
		} else {
			// No rules, no base? Just take the whole history
			reverseStrata = fullReverseHistory
		}
	}

	for dStr, c := range bestCommits {
		resolvedDates[dStr] = c.Hash.String()
	}

	// Final verification
	for _, dStr := range targetDates {
		if _, ok := resolvedDates[dStr]; !ok {
			return nil, fmt.Errorf("%s", i18n.TF("error_no_commits_for_date", map[string]interface{}{"Date": dStr}))
		}
	}

	// Reverse for chronological order
	var strata []string
	for i := len(reverseStrata) - 1; i >= 0; i-- {
		strata = append(strata, reverseStrata[i])
	}

	return &HistoryResult{
		ResolvedDates:  resolvedDates,
		Strata:         strata,
		ParentOfStrata: parentOfStrata,
		BaseHash:       baseHash,
		OriginalHead:   origHead,
		CurrentBranch:  currentBranch,
	}, nil
}

// GetCommit returns a Git object.Commit from a hexadecimal hash string.
func GetCommit(repo *git.Repository, hashStr string) (*object.Commit, error) {
	hash := plumbing.NewHash(hashStr)
	return repo.CommitObject(hash)
}

// GetCurrentBranch returns the name of the branch currently checked out at HEAD.
func GetCurrentBranch(repo *git.Repository) (string, error) {
	head, err := repo.Head()
	if err != nil {
		return "", err
	}
	if head.Name().IsBranch() {
		return head.Name().Short(), nil
	}
	return "HEAD", nil
}

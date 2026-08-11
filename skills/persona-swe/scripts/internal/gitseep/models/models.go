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

// Package models defines shared data structures for GitSeep.
package models

import (
	"sort"
)

// Options contains global configuration flags for a GitSeep execution.
type Options struct {
	// keep-sorted start
	Amend             bool
	AutoApprove       bool
	BaseCommit        string
	CheckMode         bool
	DryRun            bool
	ExperimentalGoGit bool
	Headless          bool // Skip TUI and run synchronously
	NoLithify         bool
	OriginalHead      string // Override the detected HEAD (used for virtual amendments)
	Quiet             bool
	SkipPreCommit     bool
	StageOnly         bool
	TargetBranch      string
	// keep-sorted end
}

// SeepageContext represents the resolved state of a GitSeep session.
type SeepageContext struct {
	// keep-sorted start
	BaseHash                string
	BaseRef                 string
	BranchToParent          map[string]string // branch name -> parent branch
	CurrentBranch           string
	DateToBranch            map[string]string // date -> branch name
	DateToHash              map[string]string // date -> hash
	ExcludedPaths           map[string]bool   // path -> true
	HashToDate              map[string]string // hash -> ISO date
	HashToSubject           map[string]string // hash -> commit message
	Matcher                 *GeologicalMatcher
	Options                 Options
	OriginalBranchHeads     map[string]string // branch name -> original hash
	OriginalHead            string
	OriginalToReconstructed map[string]string // original hash -> reconstructed hash
	ParentOfStrata          string
	PathToBedrock           map[string]string // path -> hash
	RepoRoot                string
	ResolvedRules           map[string][]string // hash -> paths
	Strata                  []string            // Chronological list of commit hashes
	TargetBranch            string
	// keep-sorted end
}

// ResolveBedrockForPath returns the target bedrock hash for a given file path.
func (c *SeepageContext) ResolveBedrockForPath(path string) string {
	if c.Matcher == nil {
		return ""
	}
	hash, _ := c.Matcher.ResolveTarget(path)
	return hash
}

// EnsureMatcher ensures that the geological matcher is initialized.
func (c *SeepageContext) EnsureMatcher() {
	if c.Matcher == nil {
		c.Matcher = NewGeologicalMatcher(c.PathToBedrock)
	}
}

// ResolveSubject returns the first line of the commit message for a given hash.
func (c *SeepageContext) ResolveSubject(hash string) string {
	if s, ok := c.HashToSubject[hash]; ok {
		return s
	}
	return ""
}

// GetReconstructedStrata returns the list of new commit hashes for each stratum.
func (c *SeepageContext) GetReconstructedStrata() []string {
	var res []string
	for _, s := range c.Strata {
		if h, ok := c.OriginalToReconstructed[s]; ok {
			res = append(res, h)
		}
	}
	return res
}

func (ev PipelineEvent) GetTroubleshootLog() string {
	return ev.LogOutput
}

func (ev PipelineEvent) GetTroubleshootManualCmd() string {
	if ev.Troubleshoot != nil {
		return ev.Troubleshoot.ReproductionCmd
	}
	return ""
}

func (ev PipelineEvent) GetTroubleshootCleanupCmd() string {
	if ev.Troubleshoot != nil {
		return ev.Troubleshoot.CleanupCmd
	}
	return ""
}

// SeepageSummary accumulates telemetry and results for the final report.
type SeepageSummary struct {
	// keep-sorted start
	BedrockFiles       int
	LithifiedFiles     map[string][]string
	ParityPassed       bool
	PercolateFiles     int
	PercolatePaths     map[string]bool
	SedimentedBranches map[string]bool
	SeepFiles          int
	SeepPaths          map[string]bool
	StrataProcessed    int
	// keep-sorted end
}

// DiscoveryResult holds the mapping of files to their respective bedrock targets.
type DiscoveryResult struct {
	// keep-sorted start
	Schedule map[string]map[string][]string // bedrock_hash -> source_hash -> [files]
	Sources  map[string]map[string]struct{} // file_path -> set(source_hashes)
	Touched  map[string]map[string]struct{} // source_hash -> set(file_paths)
	// keep-sorted end
}

// PipelinePhase defines the major stages of the GitSeep execution.
type PipelinePhase int

const (
	// keep-sorted start
	PhaseDiscovery PipelinePhase = iota
	PhaseExecution
	PhaseFinalization
	PhaseReview
	PhaseSedimentation
	PhaseStratigraphy
	PhaseSurface
	PhaseValidation
	// keep-sorted end
)

// PipelineEventType defines the status of a pipeline task.
type PipelineEventType int

const (
	// keep-sorted start
	EventContextResolved PipelineEventType = iota // New type for async context initialization
	EventDiagnosticLine                           // New type for streaming command output
	EventFailure
	EventFinished
	EventLog // New type for broadcasting log messages
	EventProgress
	EventSkipped
	EventStart
	EventSuccess
	// keep-sorted end
)

// PipelineEvent represents a real-time update from any phase of the GitSeep pipeline.
type PipelineEvent struct {
	// keep-sorted start
	Branch       string      // Branch name (if applicable)
	Commit       string      // Commit hash (if applicable)
	CommitDate   string      // ISO Date of commit
	CommitMsg    string      // First line of commit message
	Files        []string    // Files involved in the event (e.g. for pre-commit)
	LogOutput    string      // Full command output on failure
	Message      string      // Summary or status message
	Payload      interface{} // Phase-specific data (e.g., DiscoveryResult)
	Phase        PipelinePhase
	TempDir      string                // Path to worktree if preserved
	Troubleshoot *TroubleshootMetadata // High-fidelity guidance on failure
	Type         PipelineEventType
	// keep-sorted end
}

// TroubleshootMetadata provides high-fidelity guidance and commands upon a pipeline failure.
type TroubleshootMetadata struct {
	// keep-sorted start
	Category        FailureCategory
	CleanupCmd      string
	LogOutput       string
	ReproductionCmd string
	WorktreeDir     string
	// keep-sorted end
}

// FailureCategory classifies the type of failure for better guidance.
type FailureCategory int

const (
	FailureUnknown FailureCategory = iota
	// keep-sorted start
	FailureCheckout
	FailurePreCommit
	FailureSetup
	// keep-sorted end
)

// HasMigrations returns true if there are any files being moved between strata.
func (dr *DiscoveryResult) HasMigrations() bool {
	for bedrockH, sourcesMap := range dr.Schedule {
		if len(sourcesMap) > 1 {
			return true
		}
		// Also check if the single source is different from bedrock (though Schedule usually groups by bedrock)
		for srcH := range sourcesMap {
			if srcH != bedrockH {
				return true
			}
		}
	}
	return false
}

// MigrationItem represents a single file migration task.
type MigrationItem struct {
	Path    string
	Bedrock string
	Sources []string
}

// GetMigrationItems returns a flattened list of files that require migration (seep or percolate).
func (dr *DiscoveryResult) GetMigrationItems() []MigrationItem {
	pathData := make(map[string]map[string][]string)
	for bedrockH, sourcesMap := range dr.Schedule {
		for srcH, files := range sourcesMap {
			if srcH == bedrockH {
				continue
			}
			for _, f := range files {
				if pathData[f] == nil {
					pathData[f] = make(map[string][]string)
				}
				pathData[f][bedrockH] = append(pathData[f][bedrockH], srcH)
			}
		}
	}

	var items []MigrationItem
	var paths []string
	for p := range pathData {
		paths = append(paths, p)
	}
	sort.Strings(paths)

	for _, path := range paths {
		bedrocks := pathData[path]
		for bedrockH, sources := range bedrocks {
			items = append(items, MigrationItem{
				Path:    path,
				Bedrock: bedrockH,
				Sources: sources,
			})
		}
	}
	return items
}

// GetSummary calculates the final geological statistics.
func (dr *DiscoveryResult) GetSummary(seepage *SeepageContext) SeepageSummary {
	summary := SeepageSummary{
		LithifiedFiles: make(map[string][]string),
		PercolatePaths: make(map[string]bool),
		SeepPaths:      make(map[string]bool),
	}

	// 1. Lithifications: Files with > 1 source commit for the same bedrock
	for _, sourcesMap := range dr.Schedule {
		for _, files := range sourcesMap {
			for _, f := range files {
				sources := dr.Sources[f]
				if len(sources) > 1 {
					// Count each file only once
					if _, ok := summary.LithifiedFiles[f]; !ok {
						var srcList []string
						for s := range sources {
							srcList = append(srcList, s)
						}
						summary.LithifiedFiles[f] = srcList
					}
				}
			}
		}
	}

	// 2. Seeps & Percolations
	strataIdx := make(map[string]int)
	for i, h := range seepage.Strata {
		strataIdx[h] = i
	}

	for bedrockH, sourcesMap := range dr.Schedule {
		bedIdx, hasBedrock := strataIdx[bedrockH]
		if !hasBedrock {
			continue
		}

		for srcH, files := range sourcesMap {
			srcIdx, hasSrc := strataIdx[srcH]
			if !hasSrc {
				continue
			}

			for _, f := range files {
				if srcIdx > bedIdx {
					// Moving from newer source to older bedrock = DOWN = Percolate
					summary.PercolateFiles++
					summary.PercolatePaths[f] = true
				} else if srcIdx < bedIdx {
					// Moving from older source to newer bedrock = UP = Seep
					summary.SeepFiles++
					summary.SeepPaths[f] = true
				}
			}
		}
	}

	summary.StrataProcessed = len(seepage.Strata)
	return summary
}

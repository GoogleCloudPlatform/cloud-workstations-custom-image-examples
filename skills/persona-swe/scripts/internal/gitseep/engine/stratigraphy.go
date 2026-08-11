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

// Package engine orchestrates the core GitSeep logic including discovery,
// reconstruction, and sedimentation phases.
package engine

import (
	"fmt"
	"sort"
	"strings"

	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/i18n"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/logger"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/models"
)

// ValidateStratigraphyAndPredictConflicts runs cycle detection and conflict prediction to guarantee execution certitude.
func ValidateStratigraphyAndPredictConflicts(seepage *models.SeepageContext, dr *models.DiscoveryResult) ([]string, error) {
	l := logger.WithPhase(logger.IconBedrock, "Stratigraphy")
	l.Info("Validating geological stratigraphy and predicting conflicts...")
	if len(seepage.DateToBranch) == 0 {
		return []string{}, nil
	}

	branchNodes, orderedBranches, err := buildStratigraphyGraph(seepage)
	if err != nil {
		return nil, err
	}

	if err := predictCherryPickConflicts(seepage, dr, branchNodes, orderedBranches); err != nil {
		return nil, err
	}

	l.Success("Stratigraphy validated successfully. No mathematical conflicts predicted.")
	return orderedBranches, nil
}

func buildStratigraphyGraph(seepage *models.SeepageContext) (map[string]string, []string, error) {
	branchNodes := make(map[string]string)
	var dates []string
	for d := range seepage.DateToBranch {
		dates = append(dates, d)
	}
	sort.Slice(dates, func(i, j int) bool {
		return models.ParseRuleDate(dates[i]).Before(models.ParseRuleDate(dates[j]))
	})

	for _, dateStr := range dates {
		branchName := seepage.DateToBranch[dateStr]
		parentBranch := seepage.BranchToParent[branchName]
		if parentBranch == "" {
			// If not explicit, chronological predecessor in config is implicit parent
			idx := models.IndexOf(dates, dateStr)
			if idx > 0 {
				prevDate := dates[idx-1]
				parentBranch = seepage.DateToBranch[prevDate]
			}
		}
		branchNodes[branchName] = parentBranch
	}

	// Topological sort for ordered sedimentation
	visited := make(map[string]bool)
	tempMark := make(map[string]bool)
	var ordered []string

	var visit func(n string) error
	visit = func(n string) error {
		if tempMark[n] {
			// Cycle detected
			var cycle []string
			for p, inCycle := range tempMark {
				if inCycle {
					cycle = append(cycle, p)
				}
			}
			cycle = append(cycle, n)
			return fmt.Errorf("%s", i18n.TF("error_dependency_cycle", map[string]interface{}{"Cycle": strings.Join(cycle, " -> ")}))
		}

		if !visited[n] {
			tempMark[n] = true
			parent := branchNodes[n]
			if parent != "" && containsNode(branchNodes, parent) {
				if err := visit(parent); err != nil {
					return err
				}
			}
			tempMark[n] = false
			visited[n] = true
			ordered = append(ordered, n)
		}
		return nil
	}

	for n := range branchNodes {
		if !visited[n] {
			if err := visit(n); err != nil {
				return nil, nil, err
			}
		}
	}

	return branchNodes, ordered, nil
}

func predictCherryPickConflicts(seepage *models.SeepageContext, dr *models.DiscoveryResult, branchNodes map[string]string, orderedBranches []string) error {
	// A conflict is predicted if a branch 'B' (at bedrock 'S') modifies files
	// that were also modified in a commit 'K' that is SKIPPED.
	// However, if the file is OWNED by bedrock 'S', it's NOT a conflict because
	// we pull the final state from HEAD (lithification).

	// 1. Identify all skipped commits
	skippedStrataFiles := make(map[string]map[string]struct{})
	linearStrata := make(map[string]bool)
	for _, s := range seepage.Strata {
		linearStrata[s] = true
	}
	for commitH, filesMap := range dr.Touched {
		if !linearStrata[commitH] {
			skippedStrataFiles[commitH] = filesMap
		}
	}

	// 2. Predict conflicts for each branch/bedrock
	for _, branchName := range orderedBranches {
		var s string
		for d, b := range seepage.DateToBranch {
			if b == branchName {
				s = seepage.DateToHash[d]
				break
			}
		}
		if s == "" {
			continue
		}

		// Files touched by commit S
		touchedInS := dr.Touched[s]

		// Files OWNED by bedrock S
		ownedByS := make(map[string]struct{})
		if sourcesMap, ok := dr.Schedule[s]; ok {
			for _, files := range sourcesMap {
				for _, f := range files {
					ownedByS[f] = struct{}{}
				}
			}
		}

		// Foreign changes in S: Touched - Owned
		foreignInS := make(map[string]struct{})
		for f := range touchedInS {
			if _, owned := ownedByS[f]; !owned {
				foreignInS[f] = struct{}{}
			}
		}

		for k, filesInSkipped := range skippedStrataFiles {
			intersection := intersectSets(foreignInS, filesInSkipped)
			if len(intersection) > 0 {
				var files []string
				for f := range intersection {
					files = append(files, f)
				}
				sort.Strings(files)
				return fmt.Errorf(
					"%s", i18n.TF("error_conflict_predicted", map[string]interface{}{
						"Branch":  branchName,
						"Bedrock": logger.ColorHash(s),
						"Stratum": logger.ColorHash(k),
						"Files":   strings.Join(files, ", "),
					}),
				)
			}
		}
	}
	return nil
}

func intersectSets(s1, s2 map[string]struct{}) map[string]struct{} {
	intersection := make(map[string]struct{})
	for k := range s1 {
		if _, ok := s2[k]; ok {
			intersection[k] = struct{}{}
		}
	}
	return intersection
}

func containsNode(m map[string]string, key string) bool {
	_, ok := m[key]
	return ok
}

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

package ui

import (
	"sort"

	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/models"
)

// TreeContext provides shared utilities for rendering lane-based geological trees.
type TreeContext struct {
	OrderedHashes    []string            // Hashes sorted in post-order (Tips to Base).
	HashToParent     map[string]string   // Mapping of a commit hash to its geological parent.
	ParentToChildren map[string][]string // Mapping of a parent hash to its child dependents.
	HashToColumn     map[string]int      // The assigned vertical lane column for each hash.
	MaxColumn        int                 // The maximum column index used in the tree.
}

// NewTreeContext builds the structural geometry for a geological tree.
func NewTreeContext(seepage *models.SeepageContext) *TreeContext {
	// 1. Setup Maps
	branchToHash := make(map[string]string)
	for d, b := range seepage.DateToBranch {
		h := seepage.DateToHash[d]
		branchToHash[b] = h
	}

	hashToParent := make(map[string]string)
	parentToChildren := make(map[string][]string)

	baseH := seepage.BaseHash
	if baseH == "" {
		baseH = "BASE"
	}

	branchNodes := buildGraph(seepage)
	for b, p := range branchNodes {
		h := branchToHash[b]
		ph := branchToHash[p]
		if ph == "" {
			ph = baseH
		}
		if h != "" && h != baseH {
			hashToParent[h] = ph
			parentToChildren[ph] = append(parentToChildren[ph], h)
		}
	}

	// 2. Build ordered nodes (Tips to Base - Post-Order)
	var orderedHashes []string
	visited := make(map[string]bool)
	var collect func(h string)
	collect = func(h string) {
		if h == "" || visited[h] {
			return
		}
		visited[h] = true
		// Process children (dependents) first for post-order traversal
		children := parentToChildren[h]
		sort.Slice(children, func(i, j int) bool {
			di, dj := "", ""
			for d, hash := range seepage.DateToHash {
				if hash == children[i] {
					di = d
				}
				if hash == children[j] {
					dj = d
				}
			}
			return models.ParseRuleDate(di).After(models.ParseRuleDate(dj))
		})
		for _, child := range children {
			collect(child)
		}
		orderedHashes = append(orderedHashes, h)
	}

	// Start collection from the root anchor
	collect(baseH)

	// 3. Lane Assignment
	hashToColumn := make(map[string]int)
	nextCol := 0
	var assignLane func(h string, col int)
	assignLane = func(h string, col int) {
		if h == "" {
			return
		}
		if _, ok := hashToColumn[h]; ok {
			return
		}
		hashToColumn[h] = col
		if col >= nextCol {
			nextCol = col + 1
		}

		children := parentToChildren[h]
		for i, child := range children {
			childCol := col
			if i > 0 {
				childCol = nextCol
				nextCol++
			}
			assignLane(child, childCol)
		}
	}

	assignLane(baseH, 0)

	maxCol := 0
	for _, c := range hashToColumn {
		if c > maxCol {
			maxCol = c
		}
	}

	return &TreeContext{
		OrderedHashes:    orderedHashes,
		HashToParent:     hashToParent,
		ParentToChildren: parentToChildren,
		HashToColumn:     hashToColumn,
		MaxColumn:        maxCol,
	}
}

func buildGraph(seepage *models.SeepageContext) map[string]string {
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
		// If no parent is specified, it builds off the BASE anchor (the previous linear bedrock).
		// We no longer force a chain of feature branches.
		branchNodes[branchName] = parentBranch
	}
	return branchNodes
}

func (tc *TreeContext) RenderConnectors(h string, col int, activeLanes map[int]string) string {
	var mergingCols []int
	for c, ah := range activeLanes {
		if tc.HashToParent[ah] == h && c != col {
			mergingCols = append(mergingCols, c)
		}
	}
	sort.Ints(mergingCols)

	maxC := tc.MaxColumn
	hasMerges := len(mergingCols) > 0
	maxM := -1
	for _, mc := range mergingCols {
		if mc > maxM {
			maxM = mc
		}
	}

	conn := ""
	for c := 0; c <= maxC; c++ {
		if hasMerges && c >= col && c <= maxM {
			if c == col {
				conn += "├─"
			} else if c == maxM {
				conn += "┘ "
			} else {
				isM := false
				for _, mc := range mergingCols {
					if mc == c {
						isM = true
						break
					}
				}
				if isM {
					conn += "┴─"
				} else {
					conn += "──"
				}
			}
		} else if _, ok := activeLanes[c]; ok {
			conn += "│ "
		} else {
			conn += "  "
		}
	}
	return conn
}

func (tc *TreeContext) RenderLanes(col int, activeLanes map[int]string) string {
	lanes := ""
	for c := 0; c <= tc.MaxColumn; c++ {
		if c == col {
			lanes += "● "
		} else {
			if _, ok := activeLanes[c]; ok {
				lanes += "│ "
			} else {
				lanes += "  "
			}
		}
	}
	return lanes
}

func (tc *TreeContext) RenderAnchorLanes(col int, activeLanes map[int]string) string {
	// For the anchor, we want to connect ALL currently active lanes that have converged here.
	// In a post-order tree, these are the lanes that haven't been merged yet.
	maxTerminating := col
	for c := range activeLanes {
		if c > maxTerminating {
			maxTerminating = c
		}
	}

	lanes := ""
	for c := 0; c <= tc.MaxColumn; c++ {
		if c == col {
			if maxTerminating > col {
				lanes += "●─"
			} else {
				lanes += "● "
			}
		} else if c == maxTerminating {
			lanes += "┴─"
		} else if c < maxTerminating {
			if _, ok := activeLanes[c]; ok {
				lanes += "┴─"
			} else {
				lanes += "──"
			}
		} else {
			// Lanes beyond the anchor connection
			if _, ok := activeLanes[c]; ok {
				lanes += "│ "
			} else {
				lanes += "  "
			}
		}
	}
	return lanes
}

func (tc *TreeContext) RenderMetaLanes(col int, activeLanes map[int]string) string {
	lanes := ""
	for c := 0; c <= tc.MaxColumn; c++ {
		if _, ok := activeLanes[c]; ok || c == col {
			lanes += "│ "
		} else {
			lanes += "  "
		}
	}
	return lanes
}

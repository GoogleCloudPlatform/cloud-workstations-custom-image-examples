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
	"fmt"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-git/go-git/v5"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/i18n"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/logger"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/models"
)

// stratigraphySubModel handles the visualization of the geological history tree.
type stratigraphySubModel struct {
	seepage      *models.SeepageContext
	repo         *git.Repository
	treeCtx      *TreeContext
	width        int
	height       int
	scrollOffset int
	cursor       int
	expanded     map[string]bool
	nodes        []string // Flattened list of bedrock hashes for cursor navigation.
}

func newStratigraphySubModel(seepage *models.SeepageContext, repo *git.Repository) *stratigraphySubModel {
	var tc *TreeContext
	if seepage != nil {
		tc = NewTreeContext(seepage)
	}
	return &stratigraphySubModel{
		seepage:  seepage,
		repo:     repo,
		treeCtx:  tc,
		expanded: make(map[string]bool),
	}
}

func (s *stratigraphySubModel) handleEvent(ev models.PipelineEvent) tea.Cmd {
	if ev.Type == models.EventContextResolved {
		s.seepage = ev.Payload.(*models.SeepageContext)
		s.treeCtx = NewTreeContext(s.seepage)
	}
	return nil
}

func (s *stratigraphySubModel) pageSize() int {
	// Stratigraphy nodes are 2 lines each (node + vertical connector above)
	// We use the internal height which is the content area.
	return s.height / 2
}

func (s *stratigraphySubModel) update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		nav := NewNavigationController(&s.cursor, len(s.nodes), s.pageSize())
		if nav.HandleKey(msg) {
			return nil
		}

		switch msg.String() {
		case "a": // Toggle All
			count := 0
			for _, exp := range s.expanded {
				if exp {
					count++
				}
			}
			if count == len(s.nodes) {
				// All expanded, collapse all
				s.expanded = make(map[string]bool)
			} else {
				// Expand all
				for _, h := range s.nodes {
					s.expanded[h] = true
				}
			}
		case "v", "V": // Toggle Validation
			s.seepage.Options.SkipPreCommit = !s.seepage.Options.SkipPreCommit
		case " ": // Toggle expansion with SPACE
			if s.cursor < len(s.nodes) {
				hash := s.nodes[s.cursor]
				s.expanded[hash] = !s.expanded[hash]
			}
		case "enter": // Proceed to next tab
			return func() tea.Msg { return tea.KeyMsg{Type: tea.KeyTab} }
		}
	}
	return nil
}

func (s *stratigraphySubModel) view(availableHeight int) string {
	s.height = availableHeight
	if s.treeCtx == nil {
		return "  " + i18n.T("resolving_context") + "\n"
	}
	tc := s.treeCtx

	hashToBranches := make(map[string][]string)
	for d, b := range s.seepage.DateToBranch {
		h := s.seepage.DateToHash[d]
		hashToBranches[h] = append(hashToBranches[h], b)
	}
	for h := range hashToBranches {
		sort.Strings(hashToBranches[h])
	}

	var lines []string
	var nodeList []string
	nodeToLineIdx := make(map[string]int)
	activeLanes := make(map[int]string) // col -> hash

	for _, h := range tc.OrderedHashes {
		nodeList = append(nodeList, h)
		col := tc.HashToColumn[h]

		// 1. Identify Merges (for RenderConnectors and for clearing lanes)
		var mergingCols []int
		for c, ah := range activeLanes {
			if tc.HashToParent[ah] == h && c != col {
				mergingCols = append(mergingCols, c)
			}
		}

		// 2. Vertical connectors padding (Gravity effect)
		if len(activeLanes) > 0 {
			conn := "    " + tc.RenderConnectors(h, col, activeLanes)
			lines = append(lines, conn)
		}

		// 3. POST-CONNECTOR / PRE-NODE: Finalize the merges for the node line
		for _, mc := range mergingCols {
			delete(activeLanes, mc)
		}
		activeLanes[col] = h

		// 4. Node line
		nodeToLineIdx[h] = len(lines)
		line := "  "
		if len(nodeList)-1 == s.cursor {
			line = "\033[7m>\033[0m "
		}

		isAnchor := h == s.seepage.BaseHash
		expIcon := "▸ "
		if s.expanded[h] {
			expIcon = "▾ "
		}
		if isAnchor {
			expIcon = "  " // Anchors don't expand
		}

		lanes := ""
		if isAnchor {
			lanes = tc.RenderAnchorLanes(col, activeLanes)
		} else {
			lanes = tc.RenderLanes(col, activeLanes)
		}
		line += expIcon + lanes

		branches := hashToBranches[h]
		branchPart := ""
		if len(branches) > 0 {
			branchPart = fmt.Sprintf(" (%s)", logger.StyleBold.Render(strings.Join(branches, ", ")))
		} else if isAnchor {
			baseName := s.seepage.BaseRef
			if baseName == "" {
				baseName = i18n.T("base_ref_name")
			}
			branchPart = fmt.Sprintf(" (%s)", logger.StyleBold.Render(baseName))
		}

		subject := s.seepage.HashToSubject[h]
		if len(subject) > 40 {
			subject = subject[:37] + "..."
		}
		line += fmt.Sprintf("[%s]%s - %s", logger.ColorHash(h), branchPart, subject)
		lines = append(lines, line)

		// Metadata (Visible if expanded)
		if s.expanded[h] && !isAnchor {
			metaPrefix := "    " + tc.RenderMetaLanes(col, activeLanes)

			var date string
			for d, hash := range s.seepage.DateToHash {
				if hash == h {
					date = d
					break
				}
			}
			lines = append(lines, fmt.Sprintf("%s    🕒 %s", metaPrefix, date))
			paths := s.seepage.ResolvedRules[h]
			for _, p := range paths {
				lines = append(lines, fmt.Sprintf("%s      • %s", metaPrefix, p))
			}
		}
	}

	s.nodes = nodeList

	if len(lines) == 0 {
		return "  " + i18n.T("stratigraphy_empty") + "\n"
	}

	// Auto-scroll logic
	if s.cursor >= 0 && s.cursor < len(s.nodes) {
		h := s.nodes[s.cursor]
		if cursorLine, ok := nodeToLineIdx[h]; ok {
			if cursorLine < s.scrollOffset {
				s.scrollOffset = cursorLine
			} else if cursorLine >= s.scrollOffset+s.height {
				s.scrollOffset = cursorLine - s.height + 1
			}
		}
	}

	maxScroll := len(lines) - s.height
	if maxScroll < 0 {
		maxScroll = 0
	}
	if s.scrollOffset > maxScroll {
		s.scrollOffset = maxScroll
	}
	if s.scrollOffset < 0 {
		s.scrollOffset = 0
	}

	start := s.scrollOffset
	end := start + s.height
	if end > len(lines) {
		end = len(lines)
	}

	return strings.Join(lines[start:end], "\n")
}

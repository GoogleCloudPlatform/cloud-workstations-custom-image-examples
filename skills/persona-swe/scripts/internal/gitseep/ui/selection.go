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

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-git/go-git/v5"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/i18n"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/models"
)

// selectionSubModel handles the interactive selection of file migrations.
type selectionSubModel struct {
	seepage         *models.SeepageContext
	dr              *models.DiscoveryResult
	repo            *git.Repository
	items           []migrationItem
	cursor          int
	excluded        map[string]struct{}
	confirmed       bool // If true, the user has finalized their selection.
	engineUnblocked bool // If true, the selection has been sent to the pipeline engine.
	width           int
	height          int
}

func newSelectionSubModel(seepage *models.SeepageContext, dr *models.DiscoveryResult, repo *git.Repository) *selectionSubModel {
	migrationItems := dr.GetMigrationItems()
	var items []migrationItem
	for _, mi := range migrationItems {
		items = append(items, FormatMigrationItem(seepage, mi))
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].path < items[j].path
	})

	return &selectionSubModel{
		seepage:  seepage,
		dr:       dr,
		repo:     repo,
		items:    items,
		excluded: make(map[string]struct{}),
	}
}

func (s *selectionSubModel) pageSize() int {
	// We use 2 lines per item (Path + Bedrock line)
	// Subtract 2 for the status line and spacing
	available := s.height - 2
	if available < 2 {
		return 1
	}
	return available / 2
}

func (s *selectionSubModel) handleEvent(ev models.PipelineEvent) tea.Cmd {
	if ev.Phase == models.PhaseDiscovery && ev.Type == models.EventSuccess {
		// Re-initialize from new discovery result
		dr := ev.Payload.(*models.DiscoveryResult)
		w, h := s.width, s.height
		*s = *newSelectionSubModel(s.seepage, dr, s.repo)
		s.width, s.height = w, h
	}
	return nil
}

func (s *selectionSubModel) update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		nav := NewNavigationController(&s.cursor, len(s.items), s.pageSize())
		if nav.HandleKey(msg) {
			return nil
		}

		switch msg.String() {
		case "enter": // Confirm and Proceed
			if !s.confirmed {
				s.confirmed = true
			}
		case " ", "x": // Toggle with SPACE
			if !s.confirmed {
				path := s.items[s.cursor].path
				if _, ok := s.excluded[path]; ok {
					delete(s.excluded, path)
				} else {
					s.excluded[path] = struct{}{}
				}
			}
		case "a": // Toggle All
			if !s.confirmed {
				if len(s.excluded) == 0 {
					// Exclude all
					for _, it := range s.items {
						s.excluded[it.path] = struct{}{}
					}
				} else {
					// Select all
					s.excluded = make(map[string]struct{})
				}
			}
		case "v", "V": // Toggle Validation
			s.seepage.Options.SkipPreCommit = !s.seepage.Options.SkipPreCommit
		}
	}
	return nil
}

func (s *selectionSubModel) getExcludedPaths() []string {
	var paths []string
	for p := range s.excluded {
		paths = append(paths, p)
	}
	return paths
}

func (s *selectionSubModel) view(availableHeight int) string {
	if len(s.items) == 0 {
		return "\n  " + i18n.T("selection_empty") + "\n"
	}

	content, status := renderMigrationList(listRenderParams{
		items:       s.items,
		cursor:      s.cursor,
		height:      availableHeight,
		excluded:    s.excluded,
		interactive: !s.confirmed,
		dr:          s.dr,
		seepage:     s.seepage,
	})
	return content + status
}

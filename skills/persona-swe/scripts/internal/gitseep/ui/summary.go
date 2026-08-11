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

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-git/go-git/v5"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/i18n"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/models"
)

type summarySubModel struct {
	seepage  *models.SeepageContext
	dr       *models.DiscoveryResult
	repo     *git.Repository
	items    []migrationItem
	cursor   int
	width    int
	height   int
	excluded map[string]struct{}
}

func newSummarySubModel(seepage *models.SeepageContext, dr *models.DiscoveryResult, repo *git.Repository, excluded map[string]struct{}) *summarySubModel {
	migrationItems := dr.GetMigrationItems()
	var items []migrationItem
	for _, mi := range migrationItems {
		items = append(items, FormatMigrationItem(seepage, mi))
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].path < items[j].path
	})

	return &summarySubModel{
		seepage:  seepage,
		dr:       dr,
		repo:     repo,
		items:    items,
		excluded: excluded,
	}
}

func (s *summarySubModel) pageSize() int {
	available := s.height - 2
	if available < 2 {
		return 1
	}
	return available / 2
}

func (s *summarySubModel) update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		nav := NewNavigationController(&s.cursor, len(s.items), s.pageSize())
		nav.HandleKey(msg)
	}
	return nil
}

func (s *summarySubModel) view(availableHeight int) string {
	s.height = availableHeight
	if len(s.items) == 0 {
		return "\n  " + i18n.T("summary_empty") + "\n"
	}

	// Filter items based on LIVE excluded map
	var visibleItems []migrationItem
	for _, it := range s.items {
		if _, ok := s.excluded[it.path]; !ok {
			visibleItems = append(visibleItems, it)
		}
	}

	if len(visibleItems) == 0 {
		return "\n  " + i18n.T("summary_empty") + "\n"
	}

	stats := s.dr.GetSummary(s.seepage)
	statsView := s.renderStats(stats)

	listHeight := availableHeight - lipgloss.Height(statsView) - 2
	if listHeight < 1 {
		listHeight = 1
	}

	content, _ := renderMigrationList(listRenderParams{
		items:       visibleItems,
		cursor:      s.cursor,
		height:      listHeight,
		interactive: false,
		dr:          s.dr,
		seepage:     s.seepage,
	})
	status := "\n " + i18n.SummaryStatus(s.cursor+1, len(visibleItems))

	return content + status + "\n\n" + statsView
}

func (s *summarySubModel) renderStats(stats models.SeepageSummary) string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("33"))
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1).
		Margin(0, 1)

	content := fmt.Sprintf("%s\n\n", titleStyle.Render(i18n.T("report_geological")))
	for _, line := range i18n.GeologicalReportLines(stats.SeepFiles, stats.PercolateFiles, len(stats.LithifiedFiles)) {
		content += line + "\n"
	}

	return boxStyle.Render(content)
}

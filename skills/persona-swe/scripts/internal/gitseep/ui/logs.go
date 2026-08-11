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
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/logger"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/models"
)

// logGroup represents a titled group of log lines that can be expanded or collapsed.
type logGroup struct {
	title    string
	lines    []string
	expanded bool
	commit   string
}

// flatLogLine tracks metadata for a single rendered line in the logs viewer.
type flatLogLine struct {
	group    *logGroup
	isHeader bool
}

// logsSubModel handles the scrollable and interactive application log viewer.
type logsSubModel struct {
	viewport       viewport.Model
	items          []*logGroup
	flattened      []flatLogLine
	activeByCommit map[string]*logGroup
	cursor         int
	width          int
	height         int
}

func newLogsSubModel() *logsSubModel {
	return &logsSubModel{
		viewport:       viewport.New(0, 0),
		activeByCommit: make(map[string]*logGroup),
	}
}

func (l *logsSubModel) handleEvent(ev models.PipelineEvent) tea.Cmd {
	if l.activeByCommit == nil {
		l.activeByCommit = make(map[string]*logGroup)
	}

	if ev.Type == models.EventLog {
		l.items = append(l.items, &logGroup{title: ev.Message})
	} else if ev.Type == models.EventDiagnosticLine {
		if ev.Commit != "" {
			if group, ok := l.activeByCommit[ev.Commit]; ok {
				group.lines = append(group.lines, ev.Message)
			} else {
				// Fallback group if start event was missed
				group = &logGroup{
					title:    fmt.Sprintf("Validation [%s]", logger.ColorHash(ev.Commit)),
					commit:   ev.Commit,
					lines:    []string{ev.Message},
					expanded: true, // Auto-expand new streaming diagnostics
				}
				l.items = append(l.items, group)
				l.activeByCommit[ev.Commit] = group
			}
		} else {
			// Phase-level diagnostic
			l.items = append(l.items, &logGroup{title: ev.Message})
		}
	} else if (ev.Type == models.EventStart || ev.Type == models.EventProgress) && (ev.Phase == models.PhaseValidation || ev.Phase == models.PhaseSurface) {
		if ev.Commit != "" {
			if _, ok := l.activeByCommit[ev.Commit]; !ok {
				title := ev.Message
				if title == "" {
					title = fmt.Sprintf("Validation [%s]", logger.ColorHash(ev.Commit))
				}
				group := &logGroup{
					title:  title,
					commit: ev.Commit,
				}
				l.items = append(l.items, group)
				l.activeByCommit[ev.Commit] = group
			} else {
				// Update title if it was a fallback
				if ev.Message != "" {
					l.activeByCommit[ev.Commit].title = ev.Message
				}
			}
		}
	}

	// Limit to a reasonable number of groups
	if len(l.items) > 500 {
		l.items = l.items[1:]
	}

	l.refresh()
	return nil
}

// refresh regenerates the flattened view of the logs based on the current expansion state
// of each log group and synchronizes the viewport position.
func (l *logsSubModel) refresh() {
	var displayLines []string
	l.flattened = nil

	currentLineIdx := 0
	for _, g := range l.items {
		// 1. Process Header: Every group starts with a header line.
		l.flattened = append(l.flattened, flatLogLine{group: g, isHeader: true})
		prefix := "  "
		if currentLineIdx == l.cursor {
			prefix = "\033[7m>\033[0m "
		}

		icon := "▸"
		if g.expanded {
			icon = "▾"
		}
		if len(g.lines) == 0 {
			icon = "●"
		}

		titleStyle := lipgloss.NewStyle()
		if currentLineIdx == l.cursor {
			titleStyle = titleStyle.Bold(true)
		}

		displayLines = append(displayLines, fmt.Sprintf("%s %s %s", prefix, icon, titleStyle.Render(g.title)))
		currentLineIdx++

		// 2. Process Body: If the group is expanded, add its indented lines to the flat list.
		if g.expanded {
			for _, line := range g.lines {
				l.flattened = append(l.flattened, flatLogLine{group: g, isHeader: false})
				prefix := "      "
				if currentLineIdx == l.cursor {
					prefix = "\033[7m>\033[0m     "
				}
				displayLines = append(displayLines, prefix+logger.StyleGrey.Render(line))
				currentLineIdx++
			}
		}
	}

	l.viewport.SetContent(strings.Join(displayLines, "\n"))

	// 3. Viewport Synchronization: Ensure the logical cursor remains within the visible viewport bounds.
	if l.viewport.Height > 0 {
		if l.cursor < l.viewport.YOffset {
			l.viewport.YOffset = l.cursor
		} else if l.cursor >= l.viewport.YOffset+l.viewport.Height {
			l.viewport.YOffset = l.cursor - l.viewport.Height + 1
		}

		// Cap YOffset at the bottom to avoid showing empty lines at the end.
		maxOffset := len(displayLines) - l.viewport.Height
		if maxOffset < 0 {
			maxOffset = 0
		}
		if l.viewport.YOffset > maxOffset {
			l.viewport.YOffset = maxOffset
		}
	}

	if l.viewport.YOffset < 0 {
		l.viewport.YOffset = 0
	}
}

func (l *logsSubModel) update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		nav := NewNavigationController(&l.cursor, len(l.flattened), l.viewport.Height)
		if nav.HandleKey(msg) {
			l.refresh()
			return nil
		}

		switch msg.String() {
		case " ":
			if l.cursor >= 0 && l.cursor < len(l.flattened) {
				l.flattened[l.cursor].group.expanded = !l.flattened[l.cursor].group.expanded
			}
		case "a":
			// Toggle all: if any are expanded, collapse all. Else expand all.
			anyExpanded := false
			for _, it := range l.items {
				if it.expanded && len(it.lines) > 0 {
					anyExpanded = true
					break
				}
			}
			for _, it := range l.items {
				if len(it.lines) > 0 {
					it.expanded = !anyExpanded
				}
			}
		}
		l.refresh()
	}
	return nil
}

func (l *logsSubModel) view(availableHeight int) string {
	l.height = availableHeight
	l.viewport.Width = l.width - 4
	l.viewport.Height = l.height - 2
	l.refresh()

	logStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("33")).
		Padding(0, 1).
		Width(l.width - 4).
		Height(l.height - 2)

	return "\n " + logStyle.Render(l.viewport.View())
}

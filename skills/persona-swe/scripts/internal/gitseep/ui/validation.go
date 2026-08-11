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

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-git/go-git/v5"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/i18n"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/logger"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/models"
)

type validationSubModel struct {
	repo         *git.Repository
	seepage      *models.SeepageContext
	treeCtx      *TreeContext
	total        int
	completed    int
	statuses     []*validationStatus
	diagnostics  []string
	failure      *models.PipelineEvent
	spinner      spinner.Model
	progress     progress.Model
	width        int
	height       int
	scrollOffset int
}

type validationStatus struct {
	isSurface    bool
	isAnchor     bool
	commit       string
	originalHash string
	branch       string
	status       models.PipelineEventType
	details      string
	subject      string
	date         string
	commitMsg    string
}

func newValidationSubModel(repo *git.Repository, seepage *models.SeepageContext) *validationSubModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("33"))

	subject := resolveSurfaceSubject(seepage)
	var date string
	if seepage != nil && seepage.OriginalHead != "" {
		for d, h := range seepage.HashToDate {
			if h == seepage.OriginalHead {
				date = d
				break
			}
		}
	}

	var origHead string
	if seepage != nil {
		origHead = seepage.OriginalHead
	}

	var tc *TreeContext
	if seepage != nil {
		tc = NewTreeContext(seepage)
	}

	v := &validationSubModel{
		repo:     repo,
		seepage:  seepage,
		treeCtx:  tc,
		spinner:  s,
		progress: progress.New(progress.WithDefaultGradient(), progress.WithoutPercentage()),
		statuses: []*validationStatus{
			{
				isSurface:    true,
				status:       models.EventStart,
				commit:       origHead,
				originalHash: origHead,
				subject:      subject,
				date:         date,
				commitMsg:    i18n.T("surface_check_label"),
			},
		},
		total: 1,
	}
	return v
}

func resolveSurfaceSubject(seepage *models.SeepageContext) string {
	return i18n.T("surface_check_label")
}

func (v *validationSubModel) populateHistory(linearHashes []string, branchHeads map[string]string, seepage *models.SeepageContext) {
	v.treeCtx = NewTreeContext(seepage)
	tc := v.treeCtx

	// Map: reconstructedHash -> []branchName
	hashToBranches := make(map[string][]string)
	for name, hash := range branchHeads {
		hashToBranches[hash] = append(hashToBranches[hash], name)
	}
	// Sort branches within each hash for deterministic display
	for h := range hashToBranches {
		sort.Strings(hashToBranches[h])
	}

	// Keep existing Surface Check (index 0)
	var newStatuses []*validationStatus
	v.statuses[0].subject = resolveSurfaceSubject(seepage)
	v.statuses[0].originalHash = seepage.OriginalHead
	v.statuses[0].commit = seepage.OriginalHead
	newStatuses = append(newStatuses, v.statuses[0])

	// Project the TreeContext geometry into validation statuses
	for _, h := range tc.OrderedHashes {
		reconstructedH := seepage.OriginalToReconstructed[h]
		if reconstructedH == "" {
			reconstructedH = h // Fallback to original if not (yet) reconstructed
		}

		branches := hashToBranches[reconstructedH]
		branchLabel := strings.Join(branches, ", ")

		var date string
		for d, hash := range seepage.HashToDate {
			if hash == h {
				date = d
				break
			}
		}

		newStatuses = append(newStatuses, &validationStatus{
			commit:       reconstructedH,
			originalHash: h,
			branch:       branchLabel,
			status:       models.EventStart,
			isAnchor:     h == seepage.BaseHash,
			subject:      seepage.HashToSubject[h],
			date:         date,
			commitMsg:    seepage.HashToSubject[h],
		})
	}

	v.statuses = newStatuses
	v.total = len(v.statuses)

	// Re-sync completed count
	v.completed = 0
	for _, s := range v.statuses {
		if isTerminal(s.status) {
			v.completed++
		}
	}
}
func isTerminal(t models.PipelineEventType) bool {
	return t == models.EventSuccess || t == models.EventFailure || t == models.EventSkipped
}

func (v *validationSubModel) handleEvent(ev models.PipelineEvent) tea.Cmd {
	if ev.Type == models.EventContextResolved {
		v.seepage = ev.Payload.(*models.SeepageContext)
		v.treeCtx = NewTreeContext(v.seepage)

		// Sync Surface Check metadata immediately upon resolution
		if len(v.statuses) > 0 && v.statuses[0].isSurface {
			v.statuses[0].commit = v.seepage.OriginalHead
			v.statuses[0].originalHash = v.seepage.OriginalHead
			v.statuses[0].subject = resolveSurfaceSubject(v.seepage)

			// Try to find the date for the OriginalHead
			for d, h := range v.seepage.DateToHash {
				if h == v.seepage.OriginalHead {
					v.statuses[0].date = d
					break
				}
			}
		}
	}

	if ev.Phase == models.PhaseValidation && ev.Type == models.EventStart {
		// Discovery 2.0 (Reconstructed hashes)
		if payload, ok := ev.Payload.(map[string]interface{}); ok {
			linear := payload["linear"].([]string)
			heads := payload["heads"].(map[string]string)
			v.populateHistory(linear, heads, v.seepage)
		}
	}

	if ev.Phase == models.PhaseValidation && ev.Type == models.EventSuccess && ev.Commit == "" {
		// ALL FINISHED: Mark everything as successful to reach 100%
		// We do NOT clear diagnostics here to allow the user to see the final logs.
		for _, s := range v.statuses {
			if !isTerminal(s.status) {
				s.status = models.EventSuccess
				v.completed++
			}
		}
		return nil
	}

	if ev.Type == models.EventDiagnosticLine {
		v.diagnostics = append(v.diagnostics, ev.Message)
		if len(v.diagnostics) > 100 {
			v.diagnostics = v.diagnostics[len(v.diagnostics)-100:]
		}
		return nil
	}

	if ev.Type == models.EventProgress || ev.Type == models.EventSuccess || ev.Type == models.EventFailure || ev.Type == models.EventSkipped {
		if ev.Type == models.EventFailure {
			v.failure = &ev
		}
		for _, s := range v.statuses {
			// Strict Phase Matching
			matches := false
			if ev.Phase == models.PhaseSurface && s.isSurface {
				matches = true
			} else if ev.Phase == models.PhaseValidation && !s.isSurface {
				if s.commit == ev.Commit && (ev.Branch == "" || s.branch == ev.Branch) {
					matches = true
				}
			}

			if matches {
				// State Transition Logic for Progress Tracking
				if isTerminal(ev.Type) && !isTerminal(s.status) {
					v.completed++
				}
				s.status = ev.Type

				if ev.CommitDate != "" {
					s.date = ev.CommitDate
				}
				if ev.CommitMsg != "" {
					s.commitMsg = ev.CommitMsg
				}

				// Auto-scroll logic: keep the latest action in view (2 lines per item)
				availableHeight := v.height - 12
				listHeight := availableHeight - 10
				if listHeight < 1 {
					listHeight = 1
				}
				maxItems := listHeight / 2
				if maxItems < 1 {
					maxItems = 1
				}

				idx := 0
				for i, st := range v.statuses {
					if st == s {
						idx = i
						break
					}
				}
				if idx >= v.scrollOffset+maxItems {
					v.scrollOffset = idx - maxItems + 1
				}

				if ev.Message != "" {
					s.details = ev.Message
				}
				break
			}
		}
	}
	return nil
}

func (v *validationSubModel) update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		nav := NewNavigationController(&v.scrollOffset, len(v.statuses), v.height-15)
		nav.HandleKey(msg)
	}
	return nil
}

func (v *validationSubModel) view(availableHeight int) string {
	v.height = availableHeight
	if v.treeCtx == nil {
		return "  " + i18n.T("resolving_context") + "\n"
	}
	tc := v.treeCtx

	barWidth := v.width - 20
	if barWidth > 40 {
		barWidth = 40
	}
	v.progress.Width = barWidth
	prog := v.progress.ViewAs(float64(v.completed) / float64(v.total))
	header := " " + i18n.ValidationStatus(
		prog,
		int(float64(v.completed)/float64(v.total)*100),
		v.completed,
		v.total,
	) + "\n"

	// Responsive Layout Logic
	showDiagnostics := availableHeight >= 15
	diagHeight := 0
	if showDiagnostics {
		if availableHeight >= 30 {
			diagHeight = 16
		} else {
			// Shrink diagnostics if space is tight
			diagHeight = availableHeight - 14
			if diagHeight < 4 {
				diagHeight = 4
			}
		}
	}

	diagView := ""
	if showDiagnostics && diagHeight > 0 {
		maxDiagWidth := v.width - 6
		if maxDiagWidth < 10 {
			maxDiagWidth = 10
		}

		displayDiag := make([]string, diagHeight)
		for i := 0; i < diagHeight; i++ {
			displayDiag[i] = ""
		}

		// Fill top-down from the tail of history (chronological)
		startIdx := len(v.diagnostics) - diagHeight
		if startIdx < 0 {
			startIdx = 0
		}
		for i := 0; i < diagHeight; i++ {
			idx := startIdx + i
			if idx < len(v.diagnostics) {
				line := v.diagnostics[idx]
				if len(line) > maxDiagWidth {
					line = line[:maxDiagWidth-3] + "..."
				}
				displayDiag[i] = line
			}
		}

		diagView = "\n" + lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(0, 1).
			Width(v.width-4).
			Height(diagHeight).
			Render(strings.Join(displayDiag, "\n")) + "\n"
	}

	// Available space for list is height minus header lines and diagnostics
	listHeight := availableHeight - 2
	if showDiagnostics {
		listHeight -= lipgloss.Height(diagView)
	}
	if listHeight < 1 {
		listHeight = 1
	}

	// Dynamic sizing: use 1 line per item if space is tight, else 2 lines (connector + node)
	linesPerItem := 2
	if listHeight < 10 {
		linesPerItem = 1
	}
	maxItems := listHeight / linesPerItem
	if maxItems < 1 {
		maxItems = 1
	}

	displayList := v.statuses
	maxScroll := len(displayList) - maxItems
	if maxScroll < 0 {
		maxScroll = 0
	}
	if v.scrollOffset > maxScroll {
		v.scrollOffset = maxScroll
	}
	if v.scrollOffset < 0 {
		v.scrollOffset = 0
	}

	start := v.scrollOffset
	end := start + maxItems
	if end > len(displayList) {
		end = len(displayList)
	}

	var list []string
	activeLanes := make(map[int]string)

	for i, s := range displayList {
		h := s.originalHash
		col := 0
		if !s.isSurface {
			col = tc.HashToColumn[h]
		}

		// Identify Merges
		var mergingCols []int
		if !s.isSurface {
			for c, ah := range activeLanes {
				if tc.HashToParent[ah] == h && c != col {
					mergingCols = append(mergingCols, c)
				}
			}
		}

		// Simplified Status Icons (No Pointer column)
		statusIcon := "  "
		if !s.isAnchor {
			switch s.status {
			case models.EventSuccess:
				statusIcon = "✅"
			case models.EventFailure:
				statusIcon = "❌"
			case models.EventSkipped:
				statusIcon = "⏩"
			case models.EventStart:
				statusIcon = "⏳"
			case models.EventProgress:
				statusIcon = v.spinner.View()
			}
		}

		// Vertical connectors ABOVE node (except the very first surface check)
		if len(activeLanes) > 0 {
			conn := "   " // statusIcon(2) + space(1)
			if s.isSurface {
				// Surface starts the first lane, connect to anything previously active
				conn += tc.RenderMetaLanes(0, activeLanes)
			} else {
				conn += tc.RenderConnectors(h, col, activeLanes)
			}

			if i >= start && i < end {
				list = append(list, conn)
			}
		}

		// POST-CONNECTOR / PRE-NODE: Finalize state for current line
		if !s.isSurface {
			for _, mc := range mergingCols {
				delete(activeLanes, mc)
			}
			activeLanes[col] = h
		} else {
			// Surface starts the first lane
			activeLanes[0] = h
		}

		// Node line
		lanes := ""
		if s.isSurface {
			lanes = tc.RenderLanes(0, activeLanes)
		} else if s.isAnchor {
			lanes = tc.RenderAnchorLanes(col, activeLanes)
		} else {
			lanes = tc.RenderLanes(col, activeLanes)
		}

		hashFormatted := fmt.Sprintf("[%s]", logger.ColorHash(s.commit))
		branchPart := ""
		if s.branch != "" {
			branchPart = fmt.Sprintf(" (%s)", logger.StyleBold.Render(s.branch))
		} else if s.isSurface {
			branchPart = fmt.Sprintf(" (%s)", logger.StyleBold.Render("HEAD"))
		} else if s.isAnchor {
			baseName := v.seepage.BaseRef
			if baseName == "" {
				baseName = i18n.T("base_ref_name")
			}
			branchPart = fmt.Sprintf(" (%s)", logger.StyleBold.Render(baseName))
		}

		subject := s.subject
		if lipgloss.Width(subject) > 60 {
			// Fallback: only truncate if we can do so safely, otherwise just use a large enough limit
			if len(subject) > 60 && !strings.Contains(subject, "\x1b") {
				subject = subject[:57] + "..."
			}
		}

		if i >= start && i < end {
			list = append(list, fmt.Sprintf("%s %s%s%s - %s", statusIcon, lanes, hashFormatted, branchPart, subject))
		}
	}

	// If we only have the surface check and it's done, show waiting message
	if len(v.statuses) == 1 && v.statuses[0].status == models.EventSuccess {
		list = append(list, "\n "+logger.StyleGrey.Render("... "+i18n.T("waiting_for_reconstruction")))
	}

	return header + diagView + "\n" + strings.Join(list, "\n")
}

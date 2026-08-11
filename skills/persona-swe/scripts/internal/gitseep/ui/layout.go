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

// Package ui provides the terminal-based dashboard for GitSeep.
package ui

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/i18n"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/models"
)

// HeaderComponent renders the top application branding.
type HeaderComponent struct{}

func (h HeaderComponent) View() string {
	style := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("33")).
		Padding(0, 1)

	return "\n" + style.Render("GitSeep Dashboard") + "\n"
}

// TabsComponent renders the main navigation bar with dynamic status indicators.
type TabsComponent struct {
	ActiveTab    DashboardTab
	HasSelection bool
	Confirmed    bool
	SkipPre      bool
	ValidationH  bool // Validation Highlight (Failure or Running)
	HasSummary   bool
	PulseFrame   int
	CurrentPhase models.PipelinePhase
}

func (t TabsComponent) View() string {
	var tabs []string

	// Engine is busy during discovery, surface check, and execution (reconstruct/validation)
	isEngineBusy := (t.CurrentPhase != models.PhaseReview && t.CurrentPhase != models.PhaseStratigraphy && !t.HasSummary)

	for i := 0; i < 5; i++ {
		tab := DashboardTab(i)
		if tab == TabSummary && !t.HasSummary {
			continue
		}

		label := ""
		isActive := t.ActiveTab == tab
		isLogicalNext := false

		// Prefix with number to fix numbering as requested
		prefix := fmt.Sprintf("%d. ", i+1)

		switch tab {
		case TabLogs:
			label = prefix + i18n.T("tab_logs")
		case TabStratigraphy:
			label = prefix + i18n.T("tab_stratigraphy")
		case TabSelection:
			label = prefix + i18n.T("tab_selection")
			if t.HasSelection && !t.Confirmed {
				label += " ✋"
				isLogicalNext = true
			}
		case TabValidation:
			label = prefix + i18n.T("tab_validation")
			if t.SkipPre {
				label += " (" + i18n.T("skipped") + ")"
			}
			if t.Confirmed && !t.HasSummary {
				isLogicalNext = true
			}
		case TabSummary:
			label = prefix + i18n.T("tab_summary")
			if t.HasSummary {
				isLogicalNext = true
			}
		}

		style := lipgloss.NewStyle().Padding(0, 2)
		if isActive {
			style = style.Bold(true).Foreground(lipgloss.Color("33"))
		}

		// Animation: Pulsing underline for logical next or active engine
		if isActive || (isLogicalNext && isEngineBusy) {
			baseColor := "240" // Grey for logical next
			if isActive {
				baseColor = "33" // Blue for active
			}

			finalColor := baseColor
			if isEngineBusy {
				pulse := math.Sin(float64(t.PulseFrame)*0.15)*0.5 + 0.5
				if isActive {
					steps := []string{"25", "26", "27", "33", "39"}
					idx := int(pulse * float64(len(steps)-1))
					finalColor = steps[idx]
				} else {
					steps := []string{"236", "238", "240", "242", "244"}
					idx := int(pulse * float64(len(steps)-1))
					finalColor = steps[idx]
				}
			}

			style = style.Border(lipgloss.NormalBorder(), false, false, true, false).
				BorderForeground(lipgloss.Color(finalColor))
		} else {
			style = style.Border(lipgloss.NormalBorder(), false, false, true, false).
				BorderForeground(lipgloss.Color("235"))
		}

		tabs = append(tabs, style.Render(label))
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, tabs...) + "\n"
}

// FooterComponent renders the keyboard shortcut guidance.
type FooterComponent struct {
	ActiveTab  DashboardTab
	Confirmed  bool
	HasSummary bool
	CanProceed bool // If true, Enter switches to active tab
}

func (f FooterComponent) View() string {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Padding(1, 1)

	tabRange := "1-4"
	if f.HasSummary {
		tabRange = "1-5"
	}

	keys := []string{"q: quit", "?: help", tabRange + ": tabs"}

	// Contextual Actions
	if f.ActiveTab == TabSelection && !f.Confirmed {
		keys = append(keys, "space: toggle", "a: toggle all", "enter: confirm")
	} else if f.ActiveTab == TabStratigraphy {
		keys = append(keys, "space: toggle", "a: toggle all")
	} else if f.ActiveTab == TabLogs {
		keys = append(keys, "space: toggle", "a: toggle all")
	}

	// Global Proceed Logic
	if f.CanProceed {
		label := "enter: switch to active tab"
		if f.ActiveTab == TabSummary {
			label = "enter: exit"
		}
		keys = append(keys, label)
	}

	return style.Render(strings.Join(keys, " • "))
}

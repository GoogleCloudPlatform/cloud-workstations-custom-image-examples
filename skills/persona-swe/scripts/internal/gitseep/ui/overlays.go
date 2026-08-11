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

	"github.com/charmbracelet/lipgloss"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/i18n"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/logger"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/models"
)

// helpOverlay handles the interactive help dialog.
type helpOverlay struct {
	HasSummary bool
}

func (h helpOverlay) view() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("33")).
		MarginBottom(1).
		Align(lipgloss.Center)

	helpStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("33")).
		Padding(1, 4).
		Margin(1, 2).
		Align(lipgloss.Left)

	keyStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("33"))
	sectionStyle := lipgloss.NewStyle().Bold(true).Underline(true).MarginTop(1)

	tabRange := "1, 2, 3, 4"
	if h.HasSummary {
		tabRange = "1, 2, 3, 4, 5"
	}

	content := titleStyle.Render("GitSeep Dashboard Help") + "\n\n"

	content += sectionStyle.Render("Navigation") + "\n"
	content += fmt.Sprintf("  %s: %s\n", keyStyle.Render(tabRange), "Switch directly to tab")
	content += fmt.Sprintf("  %s: %s\n", keyStyle.Render("Tab / Shift+Tab"), "Cycle through tabs")
	content += fmt.Sprintf("  %s: %s\n", keyStyle.Render("Arrows / hjkl"), "Scroll lists / Switch tabs")

	content += sectionStyle.Render("Actions") + "\n"
	content += fmt.Sprintf("  %s: %s\n", keyStyle.Render("Space"), "Toggle selection or node")
	content += fmt.Sprintf("  %s: %s\n", keyStyle.Render("A"), "Toggle all (expand/select)")
	content += fmt.Sprintf("  %s: %s\n", keyStyle.Render("Enter"), "Confirm selection / Proceed to next phase")
	content += fmt.Sprintf("  %s: %s\n", keyStyle.Render("Q / Ctrl+C"), "Quit application")

	content += "\n" + lipgloss.NewStyle().Italic(true).Foreground(lipgloss.Color("240")).Render("Press any key to close help.")

	return "\n" + helpStyle.Render(content) + "\n"
}

// failureOverlay handles the high-fidelity error guidance dialog.
type failureOverlay struct {
	width   int
	failure *models.PipelineEvent
}

func (fo failureOverlay) view() string {
	dialogStyle := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color("9")).
		Padding(1, 2).
		Margin(1, 2).
		Width(fo.width - 8)

	f := fo.failure
	if f == nil {
		return ""
	}

	title := i18n.T("validation_failure_title")
	guidance := ""

	if f.Troubleshoot != nil {
		switch f.Troubleshoot.Category {
		case models.FailureCheckout:
			title = "⚠️  " + logger.StyleRed.Bold(true).Render("Internal Pipeline Error (Checkout)")
			guidance = "Git reported a system error during checkout. This often indicates a shallow clone or repository corruption.\n\n" +
				"Suggested Actions:\n" +
				"  - Verify integrity: git fsck\n" +
				"  - Unshallow repo:  git fetch --unshallow\n"
		case models.FailureSetup:
			title = "⚠️  " + logger.StyleRed.Bold(true).Render("Internal Pipeline Error (Setup)")
			guidance = "Failed to initialize the isolated validation environment.\n"
		case models.FailurePreCommit:
			title = "🧪 " + logger.StyleRed.Bold(true).Render(i18n.T("validation_failure_title"))
			guidance = "Code quality issues were detected. Please resolve these before continuing.\n"
		}
	}

	content := logger.StyleRed.Bold(true).Render("FAILED: ") + f.Message + "\n\n"

	if f.Commit != "" {
		content += fmt.Sprintf("Commit: [%s] %s\n", logger.ColorHash(f.Commit), f.CommitMsg)
		content += fmt.Sprintf("Date:   %s\n\n", f.CommitDate)
	}

	content += guidance + "\n"

	if f.Troubleshoot != nil {
		content += i18n.T("validation_failure_preserved") + "\n"
		content += logger.StyleGrey.Render(fmt.Sprintf("  cd %s", f.Troubleshoot.WorktreeDir)) + "\n\n"
		if f.Troubleshoot.ReproductionCmd != "" {
			content += "Reproduce locally:\n"
			content += logger.StyleGrey.Render(fmt.Sprintf("  %s", models.EnsureSkipEnv(f.Troubleshoot.ReproductionCmd))) + "\n\n"
		}
		if f.Troubleshoot.CleanupCmd != "" {
			content += i18n.T("validation_failure_cleanup_info") + "\n"
			content += logger.StyleGrey.Render(fmt.Sprintf("  %s", f.Troubleshoot.CleanupCmd)) + "\n\n"
		}
	}

	content += logger.StyleBold.Render(i18n.T("validation_failure_shortcuts"))

	return "\n" + dialogStyle.Render(title+"\n\n"+content) + "\n"
}

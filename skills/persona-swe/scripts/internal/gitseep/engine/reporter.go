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

// Reporter defines the interface for delivering geological reconstruction updates.
type Reporter interface {
	ReportLog(msg string)
	ReportProgress(phase models.PipelinePhase, msg string)
	ReportSummary(seepage *models.SeepageContext, dr *models.DiscoveryResult)
	ReportFailure(ev models.PipelineEvent)
}

// HeadlessReporter delivers updates directly to the terminal stdout.
type HeadlessReporter struct {
	Quiet bool
}

func (r *HeadlessReporter) ReportLog(msg string) {
	// Scoped logs already print to stdout via logger package
}

func (r *HeadlessReporter) ReportProgress(phase models.PipelinePhase, msg string) {
	if !r.Quiet && msg != "" {
		icon := ">>"
		name := ""
		switch phase {
		case models.PhaseDiscovery:
			icon = logger.IconSearch
			name = "Discovery"
		case models.PhaseExecution:
			icon = logger.IconExecution
			name = "Execution"
		case models.PhaseSedimentation:
			icon = logger.IconBranch
			name = "Sedimentation"
		case models.PhaseValidation:
			icon = logger.IconLab
			name = "Validation"
		case models.PhaseSurface:
			icon = logger.IconLab
			name = "Surface Check"
		}

		prefix := fmt.Sprintf("[%s]", icon)
		if name != "" {
			prefix = fmt.Sprintf("[%s %s]", icon, name)
		}
		fmt.Printf("%s %s\n", prefix, msg)
	}
}

func (r *HeadlessReporter) ReportSummary(seepage *models.SeepageContext, dr *models.DiscoveryResult) {
	if r.Quiet {
		return
	}

	items := dr.GetMigrationItems()
	if len(items) == 0 {
		return
	}

	fmt.Printf("\n%s\n", logger.StyleBold.Render(i18n.T("summary_view_title")))
	for _, it := range items {
		// Build high-fidelity details string
		sort.Slice(it.Sources, func(i, j int) bool {
			return models.IndexOf(seepage.Strata, it.Sources[i]) < models.IndexOf(seepage.Strata, it.Sources[j])
		})
		var details []string
		for _, srcH := range it.Sources {
			srcIdx := models.IndexOf(seepage.Strata, srcH)
			bedIdx := models.IndexOf(seepage.Strata, it.Bedrock)

			icon := logger.IconSeep
			dir := "↑" // Rising forward to surface
			if srcIdx > bedIdx {
				icon = logger.IconPercolate
				dir = "↓" // Falling back to foundation
			}
			details = append(details, fmt.Sprintf("%s %s %s", icon, dir, logger.ColorHash(srcH)))
		}
		detailStr := strings.Join(details, ", ")

		statusIcon := fmt.Sprintf("[%s] ", logger.ColorHash(it.Bedrock))
		fmt.Printf("  %s%s\n", statusIcon, logger.StylePath.Render(it.Path))
		fmt.Printf("      ⮑  %s\n", detailStr)
	}

	fmt.Printf("\n%s\n", i18n.TF("headless_summary_status", map[string]interface{}{
		"Total": len(items),
	}))

	stats := dr.GetSummary(seepage)
	fmt.Printf("\n%s\n", logger.StyleBold.Render(i18n.T("report_geological")))
	for _, line := range i18n.GeologicalReportLines(stats.SeepFiles, stats.PercolateFiles, len(stats.LithifiedFiles)) {
		fmt.Println(line)
	}
}

func (r *HeadlessReporter) ReportFailure(ev models.PipelineEvent) {
	if ev.Troubleshoot != nil {
		printTroubleshootingInfo(ev, ev.Troubleshoot.WorktreeDir)
	}
}

// TUIReporter broadcasts updates via the pipeline events channel.
type TUIReporter struct {
	Events chan<- models.PipelineEvent
}

func (r *TUIReporter) ReportLog(msg string) {
	r.Events <- models.PipelineEvent{Type: models.EventLog, Message: msg}
}

func (r *TUIReporter) ReportProgress(phase models.PipelinePhase, msg string) {
	r.Events <- models.PipelineEvent{Phase: phase, Type: models.EventProgress, Message: msg}
}

func (r *TUIReporter) ReportSummary(seepage *models.SeepageContext, dr *models.DiscoveryResult) {
	// TUI handles summary internally by listening for EventFinished
	r.Events <- models.PipelineEvent{Type: models.EventFinished}
}

func (r *TUIReporter) ReportFailure(ev models.PipelineEvent) {
	// TUI handles failure internally by listening for EventFailure channel events.
	// No explicit report needed here as the TUI is already listening.
}

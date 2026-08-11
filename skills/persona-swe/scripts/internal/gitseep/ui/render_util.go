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

	"github.com/charmbracelet/lipgloss"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/i18n"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/logger"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/models"
)

// migrationItem represents a single file migration for rendering in lists.
type migrationItem struct {
	path    string
	bedrock string
	sources []string
}

// FormatMigrationItem converts a raw DiscoveryResult item into a UI-optimized migrationItem.
func FormatMigrationItem(seepage *models.SeepageContext, mi models.MigrationItem) migrationItem {
	sort.Slice(mi.Sources, func(i, j int) bool {
		return models.IndexOf(seepage.Strata, mi.Sources[i]) < models.IndexOf(seepage.Strata, mi.Sources[j])
	})

	return migrationItem{
		path:    mi.Path,
		bedrock: mi.Bedrock,
		sources: mi.Sources,
	}
}

type listRenderParams struct {
	items       []migrationItem
	cursor      int
	height      int
	excluded    map[string]struct{}
	interactive bool
	dr          *models.DiscoveryResult
	seepage     *models.SeepageContext
}

// renderMigrationList provides a unified high-fidelity rendering for migration items
// used by both the Selection and Summary views.
func renderMigrationList(p listRenderParams) (string, string) {
	// windowSize: Subtract 1 for the status line
	windowSize := (p.height - 1) / 2
	if windowSize < 1 {
		windowSize = 1
	}

	start := p.cursor - (windowSize / 2)
	if start < 0 {
		start = 0
	}
	end := start + windowSize
	if end > len(p.items) {
		end = len(p.items)
		start = end - windowSize
		if start < 0 {
			start = 0
		}
	}

	content := ""
	for i := start; i < end; i++ {
		it := p.items[i]
		prefix := "  "
		if i == p.cursor {
			prefix = "\033[7m>\033[0m "
		}

		// Resolve reconstructed hashes
		origBedrockH := it.bedrock
		resBedrockH := origBedrockH
		if p.seepage != nil {
			if rh, ok := p.seepage.OriginalToReconstructed[origBedrockH]; ok && rh != "" {
				resBedrockH = rh
			}
		}

		// Status icon or Hash
		var statusIcon string
		if p.interactive {
			icon := logger.IconSelected
			if _, ok := p.excluded[it.path]; ok {
				icon = logger.IconExcluded
			}
			statusIcon = fmt.Sprintf("[%s] ", icon)
		} else {
			statusIcon = fmt.Sprintf("[%s] ", logger.ColorHash(resBedrockH))
		}

		lithLabel := ""
		if p.dr != nil && len(p.dr.Touched[it.path]) > 1 {
			lithLabel = fmt.Sprintf(" %s%s %s%s", "\033[90m", logger.IconLithify, i18n.T("lithify_label"), "\033[0m")
		}

		line1 := fmt.Sprintf("%s%s%s%s\n", prefix, statusIcon, logger.StylePath.Render(it.path), lithLabel)
		content += line1

		// Indentation for ⮑ (Should be under first char of filename)
		// prefix (2) + statusIcon (len)
		indentLen := 2 + lipgloss.Width(statusIcon)
		indent := strings.Repeat(" ", indentLen)

		// Format details with dynamic hash resolution
		var details []string
		for _, srcH := range it.sources {
			resolvedSrcH := srcH
			if p.seepage != nil {
				if rh, ok := p.seepage.OriginalToReconstructed[srcH]; ok && rh != "" {
					resolvedSrcH = rh
				}
			}

			icon := logger.IconSeep
			dir := "↑" // Rising forward to surface
			if p.seepage != nil {
				srcIdx := models.IndexOf(p.seepage.Strata, srcH)
				bedIdx := models.IndexOf(p.seepage.Strata, it.bedrock)
				if srcIdx > bedIdx {
					icon = logger.IconPercolate
					dir = "↓" // Falling back to foundation
				}
			}
			details = append(details, fmt.Sprintf("%s %s from [%s]", icon, dir, logger.ColorHash(resolvedSrcH)))
		}

		content += fmt.Sprintf("%s⮑  %s\n", indent, i18n.TF("migration_detail", map[string]interface{}{
			"Hash":    "[" + logger.ColorHash(resBedrockH) + "]",
			"Details": strings.Join(details, ", "),
		}))
	}

	var status string
	if p.interactive {
		status = "\n " + i18n.SelectionStatus(p.cursor+1, len(p.items), len(p.items)-len(p.excluded))
	} else {
		status = "\n " + i18n.SummaryStatus(p.cursor+1, len(p.items))
	}

	return content, status
}

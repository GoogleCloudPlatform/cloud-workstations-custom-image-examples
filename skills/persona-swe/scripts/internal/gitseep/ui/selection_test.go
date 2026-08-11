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
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/models"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/testutil"
)

func TestSelectionSubModel_Update(t *testing.T) {
	testutil.InitI18n()
	repo, seepage, _ := testutil.SetupLinearHistory(t)
	seepage.Strata = []string{"h1", "h2"}
	dr := &models.DiscoveryResult{
		Schedule: map[string]map[string][]string{
			"h1": {
				"h1": {"file1.txt"},
				"h2": {"file1.txt"}, // Migration!
			},
		},
	}
	sm := newSelectionSubModel(seepage, dr, repo)

	t.Run("Navigation and selection", func(t *testing.T) {
		if len(sm.items) == 0 {
			t.Fatal("sm.items is empty, expected at least one migration item")
		}
		// Toggle exclusion
		sm.update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(" ")})
		if len(sm.excluded) != 1 {
			t.Errorf("Expected 1 excluded path, got %v", sm.excluded)
		}

		// Toggle back
		sm.update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(" ")})
		if len(sm.excluded) != 0 {
			t.Errorf("Expected 0 excluded paths, got %v", sm.excluded)
		}

		// Move down
		sm.update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
		// Move up
		sm.update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	})

	t.Run("Confirmation", func(t *testing.T) {
		sm.update(tea.KeyMsg{Type: tea.KeyEnter})
		if !sm.confirmed {
			t.Error("Expected confirmed to be true after enter")
		}
	})

	t.Run("GetExcludedPaths", func(t *testing.T) {
		sm.excluded["path/to/file"] = struct{}{}
		paths := sm.getExcludedPaths()
		if len(paths) != 1 || paths[0] != "path/to/file" {
			t.Errorf("Unexpected excluded paths: %v", paths)
		}
	})
}

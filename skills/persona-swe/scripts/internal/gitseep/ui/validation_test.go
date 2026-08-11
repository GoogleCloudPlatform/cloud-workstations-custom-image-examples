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

func TestValidationSubModel_Update(t *testing.T) {
	testutil.InitI18n()
	repo, seepage, _ := testutil.SetupLinearHistory(t)
	vm := newValidationSubModel(repo, seepage)

	t.Run("Event handling", func(t *testing.T) {
		seepage.DateToHash = map[string]string{"2026-01-01": "h1"}
		seepage.Strata = []string{"h1"}
		seepage.EnsureMatcher()

		// Start event
		vm.handleEvent(models.PipelineEvent{
			Phase: models.PhaseValidation,
			Type:  models.EventStart,
			Payload: map[string]interface{}{
				"heads":  map[string]string{"feat": "h1"},
				"linear": []string{"h1"},
			},
		})
		if vm.total == 0 {
			t.Error("Expected total to be updated")
		}

		// Progress event (h1 at index 1 because index 0 is Surface Check)
		vm.handleEvent(models.PipelineEvent{
			Phase:  models.PhaseValidation,
			Type:   models.EventProgress,
			Commit: "h1",
		})

		// Success event
		vm.handleEvent(models.PipelineEvent{
			Phase:  models.PhaseValidation,
			Type:   models.EventSuccess,
			Commit: "h1",
		})
		if vm.completed != 1 {
			t.Errorf("Expected 1 completed (h1), got %d", vm.completed)
		}
	})

	t.Run("Navigation", func(t *testing.T) {
		vm.update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
		vm.update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	})
}

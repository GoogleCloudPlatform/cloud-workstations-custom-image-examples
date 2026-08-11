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
	"strings"
	"testing"

	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/models"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/testutil"
)

func TestUI_Views(t *testing.T) {
	testutil.InitI18n()
	repo, _ := testutil.SetupMemRepo(t)
	seepage := &models.SeepageContext{
		OriginalHead: "h1",
		HashToSubject: map[string]string{
			"h1": "Initial commit",
		},
	}

	m := NewDashboardModel(repo, WithSeepageContext(seepage))
	m.width = 100
	m.height = 40
	m.updateSizes()

	t.Run("Header View", func(t *testing.T) {
		h := HeaderComponent{}
		view := h.View()
		if view == "" {
			t.Error("Header View returned empty string")
		}
	})

	t.Run("Footer View", func(t *testing.T) {
		f := FooterComponent{ActiveTab: TabStratigraphy, HasSummary: false}
		view := f.View()
		// Tab range is 1-4 (Logs, Strat, Sel, Val)
		if !strings.Contains(view, "1-4: tabs") {
			t.Errorf("Expected 1-4 range when HasSummary is false, got: %s", view)
		}

		f.HasSummary = true
		view = f.View()
		// Tab range is 1-5
		if !strings.Contains(view, "1-5: tabs") {
			t.Errorf("Expected 1-5 range when HasSummary is true, got: %s", view)
		}
	})

	t.Run("Tabs View", func(t *testing.T) {
		tabs := TabsComponent{ActiveTab: TabStratigraphy}
		view := tabs.View()
		if view == "" {
			t.Error("Tabs View returned empty string")
		}
	})

	t.Run("Dashboard View", func(t *testing.T) {
		view := m.View()
		if view == "" {
			t.Error("Dashboard View returned empty string")
		}
	})

	t.Run("Help View", func(t *testing.T) {
		m.showHelp = true
		m.help.HasSummary = false
		view := m.help.view()
		// Expected 1, 2, 3, 4 (Logs, Strat, Sel, Val)
		if !strings.Contains(view, "1, 2, 3, 4: Switch directly to tab") {
			t.Errorf("Expected 1, 2, 3, 4 range, got: %s", view)
		}

		m.help.HasSummary = true
		view = m.help.view()
		// Expected 1, 2, 3, 4, 5
		if !strings.Contains(view, "1, 2, 3, 4, 5: Switch directly to tab") {
			t.Errorf("Expected 1, 2, 3, 4, 5 range, got: %s", view)
		}
		m.showHelp = false
	})

	t.Run("Logs View", func(t *testing.T) {
		m.activeTab = TabLogs
		view := m.View()
		if view == "" {
			t.Error("Logs View returned empty string")
		}
	})
}

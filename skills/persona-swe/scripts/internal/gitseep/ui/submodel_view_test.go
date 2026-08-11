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

	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/i18n"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/locales"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/models"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/testutil"
)

func TestUI_SubModelViews(t *testing.T) {
	_ = i18n.Init("en", "en", locales.Content)
	repo, seepage, _ := testutil.SetupLinearHistory(t)

	t.Run("Stratigraphy View", func(t *testing.T) {
		sm := newStratigraphySubModel(seepage, repo)
		if v := sm.view(10); v == "" {
			t.Error("Empty stratigraphy view")
		}
	})

	t.Run("Selection View", func(t *testing.T) {
		dr := &models.DiscoveryResult{}
		sm := newSelectionSubModel(seepage, dr, repo)
		if v := sm.view(10); v == "" {
			t.Error("Empty selection view")
		}
	})

	t.Run("Validation View", func(t *testing.T) {
		sm := newValidationSubModel(repo, seepage)
		if v := sm.view(10); v == "" {
			t.Error("Empty validation view")
		}
	})

	t.Run("Summary View", func(t *testing.T) {
		dr := &models.DiscoveryResult{}
		sm := newSummarySubModel(seepage, dr, repo, nil)
		if v := sm.view(10); v == "" {
			t.Error("Empty summary view")
		}
	})
}

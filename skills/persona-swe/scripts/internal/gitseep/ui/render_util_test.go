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

	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/models"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/testutil"
)

func TestUI_RenderUtil_Coverage(t *testing.T) {
	testutil.InitI18n()
	seepage := testutil.SetupMinimalSeepageContext()
	seepage.Strata = []string{"h1", "h2"}

	t.Run("FormatMigrationItem", func(t *testing.T) {
		mi := models.MigrationItem{
			Path:    "file.txt",
			Bedrock: "h1",
			Sources: []string{"h1", "h2"},
		}
		item := FormatMigrationItem(seepage, mi)
		if item.path != "file.txt" {
			t.Errorf("Expected file.txt, got %s", item.path)
		}
		if item.bedrock != "h1" {
			t.Errorf("Expected h1, got %s", item.bedrock)
		}
	})

	t.Run("renderMigrationList", func(t *testing.T) {
		items := []migrationItem{
			{path: "f1", bedrock: "h1", sources: []string{"h1"}},
			{path: "f2", bedrock: "h2", sources: []string{"h2"}},
		}
		content, status := renderMigrationList(listRenderParams{
			items:       items,
			cursor:      0,
			height:      10,
			interactive: true,
			dr:          &models.DiscoveryResult{},
			seepage:     seepage,
		})
		if content == "" {
			t.Error("Empty migration list content")
		}
		if status == "" {
			t.Error("Empty migration list status")
		}
	})
}

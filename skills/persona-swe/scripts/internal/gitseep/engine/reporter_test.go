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

package engine

import (
	"testing"

	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/models"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/testutil"
)

func TestHeadlessReporter_Coverage(t *testing.T) {
	testutil.InitI18n()
	r := &HeadlessReporter{}

	t.Run("ReportLog", func(t *testing.T) {
		r.ReportLog("test log")
	})

	t.Run("ReportProgress", func(t *testing.T) {
		r.ReportProgress(models.PhaseDiscovery, "test progress")
	})

	t.Run("ReportFailure", func(t *testing.T) {
		r.ReportFailure(models.PipelineEvent{
			Phase:        models.PhaseSurface,
			Type:         models.EventFailure,
			Troubleshoot: &models.TroubleshootMetadata{WorktreeDir: "."},
		})
	})
}

func TestTUIReporter_Coverage(t *testing.T) {
	events := make(chan models.PipelineEvent, 10)
	r := &TUIReporter{Events: events}
	r.ReportLog("log")
	r.ReportProgress(models.PhaseDiscovery, "progress")
	r.ReportSummary(nil, nil)
	r.ReportFailure(models.PipelineEvent{})
}

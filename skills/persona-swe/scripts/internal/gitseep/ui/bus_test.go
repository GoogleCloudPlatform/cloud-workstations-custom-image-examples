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
)

func TestEventBus(t *testing.T) {
	bus := NewEventBus()
	received := false

	bus.Subscribe(func(ev models.PipelineEvent) tea.Cmd {
		if ev.Type == models.EventStart {
			received = true
		}
		return nil
	})

	bus.Publish(models.PipelineEvent{Type: models.EventStart})

	if !received {
		t.Errorf("Event was not received by subscriber")
	}
}

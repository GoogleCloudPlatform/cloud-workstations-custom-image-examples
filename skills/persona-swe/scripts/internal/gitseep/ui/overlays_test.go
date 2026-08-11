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

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/models"
)

func TestLogsSubModel_AddEvent(t *testing.T) {
	l := newLogsSubModel()
	l.viewport = viewport.New(80, 20)

	// 1. General Log
	l.handleEvent(models.PipelineEvent{Type: models.EventLog, Message: "hello"})
	if len(l.items) != 1 {
		t.Errorf("Expected 1 item, got %d", len(l.items))
	}
	if l.items[0].title != "hello" {
		t.Errorf("Expected title 'hello', got %q", l.items[0].title)
	}

	// 2. Validation Start
	l.handleEvent(models.PipelineEvent{
		Phase:   models.PhaseValidation,
		Type:    models.EventStart,
		Commit:  "h1",
		Message: "Validating commit",
	})
	if len(l.items) != 2 {
		t.Errorf("Expected 2 items, got %d", len(l.items))
	}
	if l.items[1].commit != "h1" {
		t.Errorf("Expected commit h1, got %q", l.items[1].commit)
	}

	// 3. Diagnostic Line
	l.handleEvent(models.PipelineEvent{
		Type:    models.EventDiagnosticLine,
		Commit:  "h1",
		Message: "line 1",
	})
	if len(l.items[1].lines) != 1 {
		t.Errorf("Expected 1 line in group, got %d", len(l.items[1].lines))
	}
	if l.items[1].lines[0] != "line 1" {
		t.Errorf("Expected 'line 1', got %q", l.items[1].lines[0])
	}
}

func TestLogsSubModel_Interactive(t *testing.T) {
	l := newLogsSubModel()
	l.viewport = viewport.New(80, 10)
	l.items = []*logGroup{
		{title: "Group 1", lines: []string{"l1", "l2"}, expanded: false},
		{title: "Group 2", lines: []string{"l3"}, expanded: true},
	}
	l.refresh() // Flatten initial state

	// Initial flattened state:
	// 0: Header G1
	// 1: Header G2
	// 2: Detail l3
	if len(l.flattened) != 3 {
		t.Fatalf("Expected 3 flattened lines, got %d", len(l.flattened))
	}

	// 1. Navigation Down
	l.update(tea.KeyMsg{Type: tea.KeyDown})
	if l.cursor != 1 {
		t.Errorf("Expected cursor at 1, got %d", l.cursor)
	}

	// 2. Expand Group 1
	l.cursor = 0
	l.update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(" ")})
	if !l.items[0].expanded {
		t.Error("Expected Group 1 to be expanded")
	}
	if len(l.flattened) != 5 {
		// 0: Header G1
		// 1: Detail l1
		// 2: Detail l2
		// 3: Header G2
		// 4: Detail l3
		t.Errorf("Expected 5 flattened lines after expansion, got %d", len(l.flattened))
	}

	// 3. Toggle All
	l.update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	// G1 and G2 were both expanded, so "a" should collapse all
	if l.items[0].expanded || l.items[1].expanded {
		t.Error("Expected all groups to be collapsed after 'a'")
	}
}

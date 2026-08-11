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

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/models"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/testutil"
)

func TestDashboardModel_Navigation(t *testing.T) {
	testutil.InitI18n()
	repo, _ := testutil.SetupMemRepo(t)
	// Provide a closed events channel to prevent Init() from hanging
	events := make(chan models.PipelineEvent)
	close(events)

	m := NewDashboardModel(repo, WithPipelineEvents(events))

	// Manually initialize selection and summary for full navigation testing
	m.selection = &selectionSubModel{confirmed: true}
	m.summary = &summarySubModel{}

	// Default should be Stratigraphy (index 1)
	if m.activeTab != TabStratigraphy {
		t.Errorf("Default tab should be Stratigraphy (1), got %d", m.activeTab)
	}

	// Tab Right -> Selection (index 2)
	newM, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	dm := newM.(*DashboardModel)
	if dm.activeTab != TabSelection {
		t.Errorf("Expected tab Selection (2), got %d", dm.activeTab)
	}

	// Tab Right -> Validation (index 3)
	newM2, _ := dm.Update(tea.KeyMsg{Type: tea.KeyTab})
	dm2 := newM2.(*DashboardModel)
	if dm2.activeTab != TabValidation {
		t.Errorf("Expected tab Validation (3), got %d", dm2.activeTab)
	}

	// Tab Right -> Summary (index 4)
	newM3, _ := dm2.Update(tea.KeyMsg{Type: tea.KeyTab})
	dm3 := newM3.(*DashboardModel)
	if dm3.activeTab != TabSummary {
		t.Errorf("Expected tab Summary (4), got %d", dm3.activeTab)
	}

	// Tab Right -> Logs (index 0)
	newM4, _ := dm3.Update(tea.KeyMsg{Type: tea.KeyTab})
	dm4 := newM4.(*DashboardModel)
	if dm4.activeTab != TabLogs {
		t.Errorf("Expected tab Logs (0), got %d", dm4.activeTab)
	}

	// Tab Right -> Stratigraphy (index 1)
	newM5, _ := dm4.Update(tea.KeyMsg{Type: tea.KeyTab})
	dm5 := newM5.(*DashboardModel)
	if dm5.activeTab != TabStratigraphy {
		t.Errorf("Expected tab Stratigraphy (1), got %d", dm5.activeTab)
	}
}

func TestDashboardModel_LogBroadcasting(t *testing.T) {
	testutil.InitI18n()
	repo, _ := testutil.SetupMemRepo(t)
	events := make(chan models.PipelineEvent, 10)
	m := NewDashboardModel(repo,
		WithSeepageContext(testutil.SetupMinimalSeepageContext()),
		WithPipelineEvents(events),
	)

	m.Update(models.PipelineEvent{
		Type:    models.EventLog,
		Message: "test log",
	})

	if len(m.logs.items) != 1 {
		t.Errorf("Expected 1 log item, got %d", len(m.logs.items))
	}
	if m.logs.items[0].title != "test log" {
		t.Errorf("Expected 'test log', got '%s'", m.logs.items[0].title)
	}
}

func TestSelectionSubModel_Interactive(t *testing.T) {
	testutil.InitI18n()
	repo, _ := testutil.SetupMemRepo(t)
	seepageCtx := testutil.SetupMinimalSeepageContext()
	// migrations are items where srcH != bedrockH
	dr := &models.DiscoveryResult{
		Schedule: map[string]map[string][]string{
			"h1": {"src1": {"file1.txt"}},
			"hf": {"src2": {"file2.txt", "file3.txt"}},
		},
	}
	sm := newSelectionSubModel(seepageCtx, dr, repo)

	// 1. Initial State (3 files = 3 migration items)
	if len(sm.items) != 3 {
		t.Errorf("Expected 3 items, got %d", len(sm.items))
	}

	// 2. Navigation
	sm.update(tea.KeyMsg{Type: tea.KeyDown})
	if sm.cursor != 1 {
		t.Errorf("Expected cursor at 1, got %d", sm.cursor)
	}

	// 3. Toggle
	sm.update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(" ")})
	if len(sm.excluded) != 1 {
		t.Errorf("Expected 1 excluded item, got %d", len(sm.excluded))
	}

	// 4. View
	view := sm.view(40)

	if !strings.Contains(view, "file1.txt") {
		t.Errorf("Expected file1.txt in summary view")
	}
}

func TestValidationSubModel_Events(t *testing.T) {
	testutil.InitI18n()
	repo, _ := testutil.SetupMemRepo(t)
	vm := newValidationSubModel(repo, nil)

	// 1. Initial Surface Check
	if vm.total != 1 {
		t.Errorf("Expected 1 item initially, got %d", vm.total)
	}

	// 2. Population Event
	seepageCtx := &models.SeepageContext{
		DateToHash:     map[string]string{"d1": "h1", "df": "hf"},
		HashToDate:     map[string]string{"h1": "d1", "hf": "df"},
		DateToBranch:   map[string]string{"d1": "feat", "df": "foundation"},
		BranchToParent: map[string]string{"feat": "foundation"},
		BaseHash:       "BASE",
		Strata:         []string{"hf", "h1"},
	}
	seepageCtx.EnsureMatcher()
	vm.populateHistory([]string{"h1", "hf"}, map[string]string{"feat": "h1", "foundation": "hf"}, seepageCtx)

	// total should be 4: Surface + h1 (newest) + hf + BASE (anchor)
	if vm.total != 4 {
		t.Errorf("Expected 4 items after population, got %d", vm.total)
	}

	// 3. Progress Event (h1 at index 1 because it is the newest commit)
	vm.handleEvent(models.PipelineEvent{
		Phase:  models.PhaseValidation,
		Type:   models.EventProgress,
		Commit: "h1",
	})
	if vm.statuses[1].status != models.EventProgress {
		t.Errorf("Expected h1 (index 1) to be in progress, got %v", vm.statuses[1].status)
	}

	// 4. Surface Event (index 0)
	vm.handleEvent(models.PipelineEvent{
		Phase: models.PhaseSurface,
		Type:  models.EventSuccess,
	})
	if vm.statuses[0].status != models.EventSuccess {
		t.Errorf("Expected surface check to be success, got %v", vm.statuses[0].status)
	}
	if vm.completed != 1 {
		t.Errorf("Expected 1 completed item, got %d", vm.completed)
	}

	// 5. Failure Event (h1 at index 1)
	vm.handleEvent(models.PipelineEvent{
		Phase:   models.PhaseValidation,
		Type:    models.EventFailure,
		Commit:  "h1",
		Message: "failed",
	})
	if vm.statuses[1].status != models.EventFailure {
		t.Errorf("Expected h1 to be failed, got %v", vm.statuses[1].status)
	}
}

func TestDashboardModel_GlobalNavigation(t *testing.T) {
	testutil.InitI18n()
	repo, _ := testutil.SetupMemRepo(t)
	m := NewDashboardModel(repo)
	m.width = 100
	m.height = 40
	m.updateSizes()

	// 1. Enter jumps to Summary if ready
	m.activeTab = TabValidation
	m.summary = &summarySubModel{} // Simulate summary being ready

	newM, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	dm := newM.(*DashboardModel)

	if dm.activeTab != TabSummary {
		t.Errorf("Expected Enter to jump to Summary, but active tab is %d", dm.activeTab)
	}

	// 2. Enter in Summary exits
	newM2, cmd := dm.Update(tea.KeyMsg{Type: tea.KeyEnter})
	dm2 := newM2.(*DashboardModel)

	if !dm2.quitting {
		t.Error("Expected Enter in Summary to set quitting=true")
	}
	if cmd == nil {
		t.Error("Expected tea.Quit command")
	}

	// 3. Enter without Summary does NOT jump
	m.activeTab = TabValidation
	m.summary = nil
	newM3, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	dm3 := newM3.(*DashboardModel)
	if dm3.activeTab != TabValidation {
		t.Error("Enter should not jump to Summary if it is nil")
	}
}

func TestDashboardModel_PipelineNavigation(t *testing.T) {
	testutil.InitI18n()
	repo, _ := testutil.SetupMemRepo(t)
	m := NewDashboardModel(repo)

	// Test auto-switch to validation on execution start
	newM, _ := m.Update(models.PipelineEvent{
		Phase: models.PhaseExecution,
		Type:  models.EventStart,
	})
	dm := newM.(*DashboardModel)
	if dm.activeTab != TabValidation {
		t.Errorf("Expected switch to Validation tab on execution start, got %d", dm.activeTab)
	}

	// Test auto-switch to summary on finish
	seepage := &models.SeepageContext{}
	dr := &models.DiscoveryResult{}
	m.seepage = seepage
	m.dr = dr

	newM2, _ := m.Update(models.PipelineEvent{
		Type: models.EventFinished,
	})
	dm2 := newM2.(*DashboardModel)
	if dm2.activeTab != TabSummary {
		t.Errorf("Expected switch to Summary tab on finished, got %d", dm2.activeTab)
	}
	if dm2.summary == nil {
		t.Error("Expected summary model to be initialized")
	}
}

func TestDashboardModel_OverlayPriority(t *testing.T) {
	testutil.InitI18n()
	repo, _ := testutil.SetupMemRepo(t)
	m := NewDashboardModel(repo)
	m.width = 100
	m.height = 40
	m.updateSizes()

	// 1. Logs Tab Navigation
	m.activeTab = TabLogs
	// Pressing Enter should NOT close logs since it is a tab now, but let's check global Enter logic.
	// Currently global Enter jumps to Summary/Selection/Validation.
	// If nothing to jump to, it stays.
	newM, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	dm := newM.(*DashboardModel)
	if dm.activeTab != TabValidation {
		t.Errorf("Expected Enter while on logs tab to jump to Validation, but got %d", dm.activeTab)
	}

	// 2. Help Overlay Priority
	m.showHelp = true
	newM2, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	dm2 := newM2.(*DashboardModel)
	if dm2.showHelp {
		t.Error("Expected Enter to close help dialogue")
	}
}

func TestDashboardModel_FullEventPump(t *testing.T) {
	testutil.InitI18n()
	repo, _ := testutil.SetupMemRepo(t)
	m := NewDashboardModel(repo)

	events := []models.PipelineEvent{
		{Type: models.EventContextResolved, Payload: &models.SeepageContext{}},
		{Phase: models.PhaseDiscovery, Type: models.EventStart},
		{Phase: models.PhaseDiscovery, Type: models.EventProgress, Message: "Scanning..."},
		{Phase: models.PhaseDiscovery, Type: models.EventSuccess, Payload: &models.DiscoveryResult{}},
		{Phase: models.PhaseStratigraphy, Type: models.EventStart},
		{Phase: models.PhaseStratigraphy, Type: models.EventSuccess},
		{Phase: models.PhaseReview, Type: models.EventStart},
		{Phase: models.PhaseReview, Type: models.EventSuccess},
		{Phase: models.PhaseExecution, Type: models.EventStart},
		{Phase: models.PhaseExecution, Type: models.EventSuccess},
		{Phase: models.PhaseSedimentation, Type: models.EventStart},
		{Phase: models.PhaseSedimentation, Type: models.EventSuccess},
		{Phase: models.PhaseValidation, Type: models.EventStart, Payload: map[string]interface{}{"linear": []string{}, "heads": map[string]string{}}},
		{Phase: models.PhaseValidation, Type: models.EventProgress, Commit: "h1"},
		{Phase: models.PhaseValidation, Type: models.EventDiagnosticLine, Message: "running test..."},
		{Phase: models.PhaseValidation, Type: models.EventSuccess, Commit: "h1"},
		{Phase: models.PhaseValidation, Type: models.EventFinished},
		{Phase: models.PhaseFinalization, Type: models.EventStart},
		{Phase: models.PhaseFinalization, Type: models.EventSuccess},
		{Type: models.EventFinished},
	}

	for _, ev := range events {
		m.Update(ev)
	}

	if m.activeTab != TabSummary {
		t.Errorf("Expected transition to Summary tab after full pump, got %v", m.activeTab)
	}

	// Test View
	view := m.View()
	if view == "" {
		t.Error("Expected non-empty dashboard view")
	}

	// Test Window resize
	m.Update(tea.WindowSizeMsg{Width: 100, Height: 50})

	// Test Global keys
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")}) // Toggle logs
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")}) // Toggle help
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("v")}) // Toggle validation
}

func TestStratigraphySubModel_ViewCoverage(t *testing.T) {
	testutil.InitI18n()
	repo, _ := testutil.SetupMemRepo(t)
	seepage := &models.SeepageContext{
		Strata: []string{"h1", "h2"},
		HashToSubject: map[string]string{
			"h1": "Initial commit",
			"h2": "Feature update",
		},
		ResolvedRules: map[string][]string{
			"h1": {"src/main.go"},
			"h2": {"src/util.go"},
		},
	}
	sm := newStratigraphySubModel(seepage, repo)
	sm.nodes = []string{"h1", "h2"}

	// Test rendering
	view := sm.view(40)
	if view == "" {
		t.Error("Expected non-empty view")
	}

	// Interactive coverage
	sm.update(tea.KeyMsg{Type: tea.KeyDown})
	sm.update(tea.KeyMsg{Type: tea.KeyUp})
	sm.update(tea.KeyMsg{Type: tea.KeyPgDown})
	sm.update(tea.KeyMsg{Type: tea.KeyPgUp})
	sm.update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(" ")}) // Expand
}

func TestValidationSubModel_InteractiveCoverage(t *testing.T) {
	testutil.InitI18n()
	repo, _ := testutil.SetupMemRepo(t)
	vm := newValidationSubModel(repo, testutil.SetupMinimalSeepageContext())

	vm.update(tea.KeyMsg{Type: tea.KeyDown})
	vm.update(tea.KeyMsg{Type: tea.KeyUp})
	vm.update(tea.KeyMsg{Type: tea.KeyPgDown})
	vm.update(tea.KeyMsg{Type: tea.KeyPgUp})
	vm.update(tea.KeyMsg{Type: tea.KeyHome})
	vm.update(tea.KeyMsg{Type: tea.KeyEnd})

	// View rendering
	view := vm.view(40)
	if view == "" {
		t.Error("Expected non-empty view")
	}
}

func TestSummarySubModel_InteractiveCoverage(t *testing.T) {
	testutil.InitI18n()
	repo, _ := testutil.SetupMemRepo(t)
	sm := newSummarySubModel(&models.SeepageContext{}, &models.DiscoveryResult{}, repo, nil)

	sm.update(tea.KeyMsg{Type: tea.KeyDown})
	sm.update(tea.KeyMsg{Type: tea.KeyUp})

	view := sm.view(40)
	if view == "" {
		t.Error("Expected non-empty view")
	}
}

func TestSelectionSubModel_InteractiveCoverage(t *testing.T) {
	testutil.InitI18n()
	repo, _ := testutil.SetupMemRepo(t)
	dr := &models.DiscoveryResult{
		Schedule: map[string]map[string][]string{
			"h1": {"h1": {"f1.txt"}},
		},
	}
	sm := newSelectionSubModel(&models.SeepageContext{}, dr, repo)

	sm.update(tea.KeyMsg{Type: tea.KeyDown})
	sm.update(tea.KeyMsg{Type: tea.KeyUp})
	sm.update(tea.KeyMsg{Type: tea.KeyPgDown})
	sm.update(tea.KeyMsg{Type: tea.KeyPgUp})

	view := sm.view(40)
	if view == "" {
		t.Error("Expected non-empty view")
	}
}

func TestValidationSubModel_EventCoverage(t *testing.T) {
	testutil.InitI18n()
	repo, _ := testutil.SetupMemRepo(t)
	vm := newValidationSubModel(repo, testutil.SetupMinimalSeepageContext())

	vm.handleEvent(models.PipelineEvent{
		Phase:   models.PhaseValidation,
		Type:    models.EventStart,
		Payload: map[string]interface{}{"linear": []string{"h1"}, "heads": map[string]string{}},
	})
	vm.handleEvent(models.PipelineEvent{
		Phase:  models.PhaseValidation,
		Type:   models.EventProgress,
		Commit: "h1",
	})
	vm.handleEvent(models.PipelineEvent{
		Phase:  models.PhaseValidation,
		Type:   models.EventSuccess,
		Commit: "h1",
	})
	vm.handleEvent(models.PipelineEvent{
		Phase:   models.PhaseValidation,
		Type:    models.EventFailure,
		Commit:  "h1",
		Message: "failed",
	})
}

func TestUI_KeyboardInteractions(t *testing.T) {
	testutil.InitI18n()
	repo, seepage, _ := testutil.SetupLinearHistory(t)
	m := NewDashboardModel(repo, WithSeepageContext(seepage))
	m.width = 100
	m.height = 40
	m.updateSizes()

	// Manually initialize selection sub-model for testing tab switching
	dr := &models.DiscoveryResult{}
	m.dr = dr
	m.selection = newSelectionSubModel(seepage, dr, repo)

	t.Run("Tab switching", func(t *testing.T) {
		// 3 = Selection (0=Logs, 1=Stratigraphy, 2=Selection, 3=Validation, 4=Summary)
		// Keys: 1=Logs, 2=Stratigraphy, 3=Selection, 4=Validation, 5=Summary

		// 3 = Selection
		m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("3")})
		dm := m2.(*DashboardModel)
		if dm.activeTab != TabSelection {
			t.Errorf("Expected TabSelection, got %v", dm.activeTab)
		}

		// 2 = Stratigraphy
		m3, _ := dm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("2")})
		dm = m3.(*DashboardModel)
		if dm.activeTab != TabStratigraphy {
			t.Errorf("Expected TabStratigraphy, got %v", dm.activeTab)
		}

		// 1 = Logs
		m4, _ := dm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("1")})
		dm = m4.(*DashboardModel)
		if dm.activeTab != TabLogs {
			t.Errorf("Expected TabLogs, got %v", dm.activeTab)
		}
	})

	t.Run("Help toggle", func(t *testing.T) {
		m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")})
		dm := m2.(*DashboardModel)
		if !dm.showHelp {
			t.Error("Help should be visible")
		}
		m3, _ := dm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")})
		dm = m3.(*DashboardModel)
		if dm.showHelp {
			t.Error("Help should be hidden")
		}
	})

	t.Run("Logs switch", func(t *testing.T) {
		m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")})
		dm := m2.(*DashboardModel)
		if dm.activeTab != TabLogs {
			t.Errorf("Expected logs tab on 'l', got %v", dm.activeTab)
		}
	})

	t.Run("Quit", func(t *testing.T) {
		_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
		if cmd == nil {
			t.Error("Expected quit command")
		}
	})
}

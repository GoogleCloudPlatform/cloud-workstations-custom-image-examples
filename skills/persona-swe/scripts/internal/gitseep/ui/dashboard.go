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

// Package ui provides the Terminal User Interface for GitSeep, including the unified dashboard.
package ui

import (
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-git/go-git/v5"
	gitseepGit "github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/git"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/i18n"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/models"
)

// DashboardTab represents the different views available in the TUI.
type DashboardTab int

const (
	TabLogs         DashboardTab = iota // Tab 1: Application and pipeline logs.
	TabStratigraphy                     // Tab 2: Geological tree visualization.
	TabSelection                        // Tab 3: File migration selection.
	TabValidation                       // Tab 4: Real-time strata validation.
	TabSummary                          // Tab 5: Final geological report.
)

const MaxLogLines = 2000

// DashboardModel is the primary Bubble Tea model for the GitSeep TUI.
// It coordinates navigation between sub-models and handles global overlays.
type DashboardModel struct {
	repo      *git.Repository
	git       gitseepGit.GitService
	seepage   *models.SeepageContext
	dr        *models.DiscoveryResult
	activeTab DashboardTab
	quitting  bool
	err       error

	// Overlays
	help    helpOverlay
	failure failureOverlay

	showHelp          bool
	showFailureDialog bool

	bus *EventBus

	// Shared State
	currentPhase models.PipelinePhase
	events       <-chan models.PipelineEvent
	confirmChan  chan<- []string // Sends excluded files to unblock engine

	// Components
	spinner      spinner.Model
	progress     progress.Model
	stratigraphy *stratigraphySubModel
	selection    *selectionSubModel
	validation   *validationSubModel
	summary      *summarySubModel
	logs         *logsSubModel

	width         int
	height        int
	contentHeight int

	// Animation
	tabAnimationFrame int
}

// DashboardOption defines a functional option for configuring the DashboardModel.
type DashboardOption func(*DashboardModel)

// WithGitService sets the Git service provider.
func WithGitService(git gitseepGit.GitService) DashboardOption {
	return func(m *DashboardModel) {
		m.git = git
	}
}

// WithSeepageContext sets the initial seepage context.
func WithSeepageContext(seepage *models.SeepageContext) DashboardOption {
	return func(m *DashboardModel) {
		m.seepage = seepage
	}
}

// WithPipelineEvents sets the pipeline events source.
func WithPipelineEvents(events <-chan models.PipelineEvent) DashboardOption {
	return func(m *DashboardModel) {
		m.events = events
	}
}

// WithConfirmationChan sets the engine unblocking channel.
func WithConfirmationChan(confirm chan<- []string) DashboardOption {
	return func(m *DashboardModel) {
		m.confirmChan = confirm
	}
}

func NewDashboardModel(repo *git.Repository, options ...DashboardOption) *DashboardModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("33"))

	m := &DashboardModel{
		repo:      repo,
		activeTab: TabStratigraphy,
		spinner:   s,
		progress:  progress.New(progress.WithDefaultGradient(), progress.WithoutPercentage()),
		logs:      newLogsSubModel(),
		bus:       NewEventBus(),
	}

	for _, opt := range options {
		opt(m)
	}

	if m.seepage != nil {
		m.stratigraphy = newStratigraphySubModel(m.seepage, m.repo)
		m.bus.Subscribe(m.stratigraphy.handleEvent)
	}
	m.validation = newValidationSubModel(m.repo, m.seepage)
	m.bus.Subscribe(m.validation.handleEvent)
	m.bus.Subscribe(m.logs.handleEvent)

	return m
}

func (m *DashboardModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.validation.spinner.Tick,
		m.listenEvents(),
	)
}

func (m *DashboardModel) listenEvents() tea.Cmd {
	return func() tea.Msg {
		if m.events == nil {
			return nil
		}
		ev, ok := <-m.events
		if !ok {
			return nil // Channel closed
		}
		return ev
	}
}

func (m *DashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateSizes()

	case tea.KeyMsg:
		k := msg.String()
		if k == "ctrl+c" || k == "q" {
			m.quitting = true
			return m, tea.Quit
		}

		// Priority 1: Overlays (These must capture keys before global navigation)
		if m.showFailureDialog {
			if k == "enter" {
				m.quitting = true
				return m, tea.Quit
			}
			return m, nil
		}
		if m.showHelp {
			m.showHelp = false
			return m, nil
		}

		// Priority 2: Global "Enter" Navigation Logic
		if k == "enter" {
			if m.activeTab == TabSummary {
				m.quitting = true
				return m, tea.Quit
			}

			// Priority 1: If selection is required and we are NOT on that tab, jump there.
			if m.selection != nil && !m.selection.confirmed {
				if m.activeTab != TabSelection {
					m.activeTab = TabSelection
					return m, nil
				}
				// If we ARE on Selection, let sub-model delegation handle the 'enter' to confirm.
			} else if m.summary != nil {
				// Priority 2: If everything is done, jump to Summary.
				if m.activeTab != TabSummary {
					m.activeTab = TabSummary
					return m, nil
				}
			} else if m.activeTab != TabValidation {
				// Priority 3: Jump to Validation (Execution).
				m.activeTab = TabValidation
				return m, nil
			}
		}

		switch k {
		case "?":
			m.showHelp = true
		case "1":
			m.activeTab = TabLogs
		case "2":
			m.activeTab = TabStratigraphy
		case "3":
			m.activeTab = TabSelection
		case "4":
			m.activeTab = TabValidation
		case "5":
			if m.summary != nil {
				m.activeTab = TabSummary
			}
		case "tab", "right":
			m.activeTab = (m.activeTab + 1) % 5
			if m.activeTab == TabSummary && m.summary == nil {
				m.activeTab = TabLogs
			}
		case "shift+tab", "left":
			m.activeTab = (m.activeTab - 1 + 5) % 5
			if m.activeTab == TabSummary && m.summary == nil {
				// Jump back to Validation if Summary is hidden
				m.activeTab = TabValidation
			}
		}

		// Delegate keyboard to sub-models
		var cmd tea.Cmd
		switch m.activeTab {
		case TabLogs:
			if m.logs != nil {
				cmd = m.logs.update(msg)
			}
		case TabStratigraphy:
			if m.stratigraphy != nil {
				cmd = m.stratigraphy.update(msg)
			}
		case TabSelection:
			if m.selection != nil {
				cmd = m.selection.update(msg)
				if m.selection.confirmed && !m.selection.engineUnblocked {
					m.selection.engineUnblocked = true
					if m.confirmChan != nil {
						m.confirmChan <- m.selection.getExcludedPaths()
					}
				}
			}
		case TabValidation:
			if m.validation != nil {
				cmd = m.validation.update(msg)
			}
		case TabSummary:
			if m.summary != nil {
				cmd = m.summary.update(msg)
			}
		}
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
		m.validation.spinner, cmd = m.validation.spinner.Update(msg)
		cmds = append(cmds, cmd)
		m.tabAnimationFrame++
		if m.tabAnimationFrame > 100 {
			m.tabAnimationFrame = 0
		}

	case models.PipelineEvent:
		m.currentPhase = msg.Phase

		if msg.Type == models.EventContextResolved {
			m.seepage = msg.Payload.(*models.SeepageContext)
			if m.stratigraphy == nil {
				m.stratigraphy = newStratigraphySubModel(m.seepage, m.repo)
				m.bus.Subscribe(m.stratigraphy.handleEvent)
			}
			m.updateSizes()
		}

		if msg.Phase == models.PhaseDiscovery && msg.Type == models.EventSuccess {
			m.dr = msg.Payload.(*models.DiscoveryResult)
			if m.selection == nil {
				m.selection = newSelectionSubModel(m.seepage, m.dr, m.repo)
				m.bus.Subscribe(m.selection.handleEvent)
			}
			m.updateSizes()
			if m.seepage != nil && m.seepage.Options.AutoApprove {
				m.activeTab = TabValidation
				m.confirmChan <- nil
			}
		}

		if msg.Phase == models.PhaseExecution && msg.Type == models.EventStart {
			// Reconciliation started, auto-switch to validation tab
			m.activeTab = TabValidation
		}

		if msg.Type == models.EventFailure {
			m.showFailureDialog = true
			m.failure = failureOverlay{
				width:   m.width,
				failure: &msg,
			}
			// Fallback troubleshoot if missing
			if m.failure.failure.Troubleshoot == nil {
				repoRoot := "."
				if m.seepage != nil {
					repoRoot = m.seepage.RepoRoot
				}
				m.failure.failure.Troubleshoot = &models.TroubleshootMetadata{
					Category:    models.FailureUnknown,
					WorktreeDir: repoRoot,
				}
			}
		}

		if msg.Type == models.EventFinished {
			if m.seepage != nil && m.dr != nil {
				excluded := make(map[string]struct{})
				if m.selection != nil {
					excluded = m.selection.excluded
				}
				m.summary = newSummarySubModel(m.seepage, m.dr, m.repo, excluded)
				m.updateSizes()
				m.activeTab = TabSummary
			} else {
				// Fallback if no context resolved
				m.quitting = true
				return m, tea.Quit
			}
		}

		// Broadcast event to all sub-models via Bus
		cmds = append(cmds, m.bus.Publish(msg))

		cmds = append(cmds, m.listenEvents())
	}

	return m, tea.Batch(cmds...)
}

func (m *DashboardModel) updateSizes() {
	// Chrome calculation: Header (3) + Tabs (2) + Footer (1) + Padding (2)
	m.contentHeight = m.height - 8
	if m.contentHeight < 1 {
		m.contentHeight = 1
	}

	if m.stratigraphy != nil {
		m.stratigraphy.width = m.width
		m.stratigraphy.height = m.contentHeight
	}
	if m.selection != nil {
		m.selection.width = m.width
		m.selection.height = m.contentHeight
	}
	if m.summary != nil {
		m.summary.width = m.width
		m.summary.height = m.contentHeight
	}
	m.validation.width = m.width
	m.validation.height = m.contentHeight

	if m.logs != nil {
		m.logs.width = m.width
		m.logs.height = m.contentHeight
	}

	if m.showFailureDialog {
		m.failure.width = m.width
	}
}

func (m *DashboardModel) View() string {
	if m.quitting {
		return ""
	}

	if m.showFailureDialog {
		return m.failure.view()
	}

	if m.showHelp {
		m.help.HasSummary = m.summary != nil
		return m.help.view()
	}

	header := HeaderComponent{}
	tabs := TabsComponent{
		ActiveTab:    m.activeTab,
		HasSelection: m.selection != nil,
		Confirmed:    m.selection != nil && m.selection.confirmed,
		SkipPre:      m.seepage != nil && m.seepage.Options.SkipPreCommit,
		ValidationH:  m.validation != nil && (m.validation.failure != nil || (m.currentPhase == models.PhaseValidation && m.validation.completed < m.validation.total)),
		HasSummary:   m.summary != nil,
		PulseFrame:   m.tabAnimationFrame,
		CurrentPhase: m.currentPhase,
	}

	var content string
	switch m.activeTab {
	case TabLogs:
		content = m.logs.view(m.contentHeight)
	case TabStratigraphy:
		if m.stratigraphy != nil {
			content = m.stratigraphy.view(m.contentHeight)
		} else {
			content = "  " + m.spinner.View() + " " + i18n.T("resolving_context")
		}
	case TabSelection:
		if m.selection != nil {
			content = m.selection.view(m.contentHeight)
		} else {
			content = "  " + m.spinner.View() + " " + i18n.T("analyzing_history")
		}
	case TabValidation:
		content = m.validation.view(m.contentHeight)
	case TabSummary:
		if m.summary != nil {
			content = m.summary.view(m.contentHeight)
		} else {
			content = "  " + m.spinner.View() + " " + i18n.T("finalizing_reconstruction")
		}
	}

	footer := FooterComponent{
		ActiveTab:  m.activeTab,
		Confirmed:  m.selection != nil && m.selection.confirmed,
		HasSummary: m.summary != nil,
		CanProceed: m.canProceed(),
	}

	contentStyle := lipgloss.NewStyle().
		Height(m.contentHeight).
		MaxHeight(m.contentHeight)

	return lipgloss.JoinVertical(lipgloss.Left,
		header.View(),
		tabs.View(),
		contentStyle.Render(content),
		footer.View(),
	)
}

func (m *DashboardModel) canProceed() bool {
	// Enter always does something useful if these conditions are met:
	if m.activeTab == TabSummary {
		return true // Exit
	}
	if m.selection != nil && !m.selection.confirmed && m.activeTab != TabSelection {
		return true // Jump to selection
	}
	if m.summary != nil && m.activeTab != TabSummary {
		return true // Jump to summary
	}
	if m.activeTab != TabValidation && (m.selection == nil || m.selection.confirmed) && m.summary == nil {
		return true // Jump to validation
	}
	return false
}

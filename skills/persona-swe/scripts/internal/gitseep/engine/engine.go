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

// Package engine orchestrates the core GitSeep logic including discovery,
// reconstruction, and sedimentation phases.
package engine

import (
	"fmt"
	"strings"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/config"
	gitseepGit "github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/git"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/i18n"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/locales"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/logger"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/models"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/ui"
)

// Run executes the full GitSeep pipeline. It defaults to an interactive TUI
// but switches to a synchronous Headless Mode if specified in options.
func Run(cfg *config.GitSeepConfig, opts models.Options) error {
	if err := i18n.Init("", "en", locales.Content); err != nil {
		return fmt.Errorf("failed to initialize i18n: %w", err)
	}

	gitSvc, err := gitseepGit.NewGitService(".", opts.ExperimentalGoGit)
	if err != nil {
		return err
	}

	repo := gitSvc.GetRepo()

	if err := VerifyWorktreeIsClean(gitSvc, opts); err != nil {
		return err
	}

	// Mandatory Staged Amendment
	if opts.Amend && !opts.CheckMode {
		if opts.DryRun {
			virtualHash, err := gitSvc.VirtualAmend()
			if err != nil {
				return fmt.Errorf("failed to simulate amendment: %w", err)
			}
			opts.OriginalHead = virtualHash.String()
			logger.Info("Simulating amendment of staged changes (virtual hash: %s)", opts.OriginalHead[:7])

			// REFRESH STATE: The virtual commit is in the object database, but go-git
			// might need a refresh to "see" it if it's heavily caching.
			gitSvc, err = gitseepGit.NewGitService(".", opts.ExperimentalGoGit)
			if err != nil {
				return fmt.Errorf("failed to refresh git state for virtual commit: %w", err)
			}
			repo = gitSvc.GetRepo()
		} else {
			if err := gitSvc.Amend(); err != nil {
				return fmt.Errorf("failed to amend staged changes: %w", err)
			}
			// REFRESH STATE: Re-open repository to sync HEAD after amendment
			gitSvc, err = gitseepGit.NewGitService(".", opts.ExperimentalGoGit)
			if err != nil {
				return fmt.Errorf("failed to refresh git state: %w", err)
			}
			repo = gitSvc.GetRepo()
		}
	}

	events := make(chan models.PipelineEvent, 100)
	confirmChan := make(chan []string, 1)

	var reporter Reporter
	if opts.Headless {
		reporter = &HeadlessReporter{Quiet: opts.Quiet}
		confirmChan <- nil // Auto-approve

		// Background drain for headless (progress only if not quiet)
		go func() {
			for ev := range events {
				reporter.ReportProgress(ev.Phase, ev.Message)
			}
		}()

		pipeline := NewHistoryPipeline(cfg, opts, repo,
			WithGitService(gitSvc),
			WithEvents(events),
			WithConfirmation(confirmChan),
			WithReporter(reporter),
		)
		pipeline.Execute()
		return nil
	}

	// TUI MODE
	reporter = &TUIReporter{Events: events}
	oldQuiet := logger.QuietMode
	logger.QuietMode = true
	logger.LogHook = func(msg string) {
		reporter.ReportLog(msg)
	}
	defer func() {
		logger.LogHook = nil
		logger.QuietMode = oldQuiet
	}()

	m := ui.NewDashboardModel(repo,
		ui.WithGitService(gitSvc),
		ui.WithPipelineEvents(events),
		ui.WithConfirmationChan(confirmChan),
	)
	p := tea.NewProgram(m, tea.WithAltScreen())

	pipeline := NewHistoryPipeline(cfg, opts, repo,
		WithGitService(gitSvc),
		WithEvents(events),
		WithConfirmation(confirmChan),
		WithReporter(reporter),
	)
	go pipeline.Execute()

	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("dashboard failed: %w", err)
	}

	// Handle Exit States (Diagnostic Exit)
	if m, ok := finalModel.(*ui.DashboardModel); ok {
		if dir := m.GetTroubleshootDir(); dir != "" {
			printTroubleshootingInfo(m, dir)
			return nil
		}
		if m.GetError() != nil {
			return m.GetError()
		}
	}

	return nil
}

type troubleshootProvider interface {
	GetTroubleshootLog() string
	GetTroubleshootManualCmd() string
	GetTroubleshootCleanupCmd() string
}

func printTroubleshootingInfo(m troubleshootProvider, dir string) {
	troubleshootLog := m.GetTroubleshootLog()
	if troubleshootLog != "" {
		fmt.Printf("\n%s\n", troubleshootLog)
	}

	fmt.Printf("\n%s\n", logger.StyleBold.Render(i18n.T("shell_escape_activated")))
	fmt.Printf("%s\n\n", i18n.T("shell_escape_guidance"))

	fmt.Printf("%s\n", i18n.T("shell_exit_info"))
	fmt.Printf("%s\n", i18n.T("manual_return_info"))

	// Reproduction block
	fmt.Printf("\n%s\n", i18n.T("reproduce_label"))
	fmt.Printf("  cd %s\n", dir)
	if manualCmd := m.GetTroubleshootManualCmd(); manualCmd != "" {
		fmt.Printf("  %s\n", models.EnsureSkipEnv(manualCmd))
	}

	// Cleanup block
	if cleanup := m.GetTroubleshootCleanupCmd(); cleanup != "" {
		fmt.Printf("\n%s\n", i18n.T("cleanup_instructions_header"))
		fmt.Printf("  %s\n", cleanup)
	}
	fmt.Println()
}

// Check performs a non-interactive validation of the current workspace against seepage rules.
func Check(cfg *config.GitSeepConfig, opts models.Options) error {
	gitSvc, err := gitseepGit.NewGitService(".", opts.ExperimentalGoGit)
	if err != nil {
		return err
	}

	repo := gitSvc.GetRepo()

	opts.DryRun = true
	opts.CheckMode = true

	seepage, err := NewContext(gitSvc, cfg, opts)
	if err != nil {
		return err
	}

	var mu sync.Mutex

	dr, err := Discover(DiscoveryParams{
		Seepage: seepage,
		Repo:    repo,
		Events:  nil,
		RepoMu:  &mu,
	})
	if err != nil {
		return err
	}

	orderedBranches, err := ValidateStratigraphyAndPredictConflicts(seepage, dr)
	if err != nil {
		return err
	}

	// Verify stratigraphy: every managed file should only be touched in its bedrock commit.
	for bedrockH, sourcesMap := range dr.Schedule {
		for srcH := range sourcesMap {
			if srcH != bedrockH {
				return fmt.Errorf(
					"%s", i18n.TF("error_stratigraphy_modification", map[string]interface{}{
						"Hash": logger.ColorHash(srcH),
					}),
				)
			}
		}
	}

	linearCommits, _, err := ReconstructHistory(ReconstructionParams{
		Seepage: seepage,
		Result:  dr,
		Repo:    repo,
		Events:  nil,
	})
	if err != nil {
		return err
	}

	branchHeads, err := PerformSedimentation(SedimentationParams{
		Seepage:         seepage,
		Result:          dr,
		Repo:            repo,
		OrderedBranches: orderedBranches,
		LinearCommits:   linearCommits,
		Events:          nil,
	})
	if err != nil {
		return err
	}

	var linearHashes []string
	for _, stratum := range seepage.Strata {
		if newHash, ok := linearCommits[stratum]; ok {
			linearHashes = append(linearHashes, newHash)
		}
	}

	if err := ValidateHistory(ValidationParams{
		Seepage:      seepage,
		RepoRoot:     seepage.RepoRoot,
		Repo:         repo,
		OriginalHead: seepage.OriginalHead,
		LinearHashes: linearHashes,
		BranchHeads:  branchHeads,
		Skip:         false,
		SkipHash:     seepage.BaseHash,
		Events:       nil,
		RepoMu:       &mu,
	}); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	l := logger.WithPhase(logger.IconLab, "Check")
	l.Success("%s", i18n.T("history_sedimented_success"))
	return nil
}

// NewContext resolves bedrock commits and maps the linear history into a SeepageContext.
func NewContext(gitSvc gitseepGit.GitService, cfg *config.GitSeepConfig, opts models.Options) (*models.SeepageContext, error) {
	repo := gitSvc.GetRepo()
	seepage := &models.SeepageContext{
		BaseRef:        cfg.Global.BaseRef,
		ResolvedRules:  make(map[string][]string),
		PathToBedrock:  make(map[string]string),
		HashToSubject:  make(map[string]string),
		DateToBranch:   make(map[string]string),
		BranchToParent: make(map[string]string),
		DateToHash:     make(map[string]string),
		HashToDate:     make(map[string]string),
		Options:        opts,
		RepoRoot:       gitSvc.RepoRoot(),
	}

	var targetDates []string
	for _, rule := range cfg.Rules {
		targetDates = append(targetDates, rule.Date)
	}

	result, err := gitseepGit.GetSeepageHistory(gitSvc, targetDates, opts.BaseCommit)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate linear history: %w", err)
	}

	seepage.DateToHash = result.ResolvedDates
	seepage.Strata = result.Strata
	seepage.ParentOfStrata = result.ParentOfStrata
	seepage.BaseHash = result.BaseHash
	seepage.OriginalHead = result.OriginalHead
	if opts.OriginalHead != "" {
		seepage.OriginalHead = opts.OriginalHead
	}
	seepage.CurrentBranch = result.CurrentBranch
	seepage.OriginalBranchHeads = make(map[string]string)

	if seepage.TargetBranch == "" {
		seepage.TargetBranch = result.CurrentBranch
	}

	captureMetadata(seepage, repo, result)
	mapRules(seepage, repo, cfg, result)

	seepage.EnsureMatcher()
	return seepage, nil
}

func captureMetadata(seepage *models.SeepageContext, repo *git.Repository, result *gitseepGit.HistoryResult) {
	// Capture original branch heads
	iter, err := repo.Branches()
	if err == nil {
		_ = iter.ForEach(func(ref *plumbing.Reference) error {
			seepage.OriginalBranchHeads[ref.Name().Short()] = ref.Hash().String()
			return nil
		})
	}

	// Reverse mapping for dates
	for d, h := range result.ResolvedDates {
		seepage.HashToDate[h] = d
	}

	// Capture subjects
	subjects := []string{result.BaseHash, result.ParentOfStrata}
	subjects = append(subjects, result.Strata...)

	for _, hash := range subjects {
		if hash == "" {
			continue
		}
		if _, ok := seepage.HashToSubject[hash]; ok {
			continue
		}
		if c, err := gitseepGit.GetCommit(repo, hash); err == nil {
			seepage.HashToSubject[hash] = strings.Split(c.Message, "\n")[0]
		}
	}
}

func mapRules(seepage *models.SeepageContext, repo *git.Repository, cfg *config.GitSeepConfig, result *gitseepGit.HistoryResult) {
	for _, rule := range cfg.Rules {
		hash := result.ResolvedDates[rule.Date]
		branchName := rule.Branch
		if branchName == "" {
			branchName = "sediment/" + rule.Date
		}

		seepage.DateToBranch[rule.Date] = branchName
		seepage.BranchToParent[branchName] = rule.Parent
		seepage.ResolvedRules[hash] = rule.Paths

		for _, p := range rule.Paths {
			seepage.PathToBedrock[p] = hash
		}

		if _, ok := seepage.HashToSubject[hash]; !ok {
			if c, err := gitseepGit.GetCommit(repo, hash); err == nil {
				seepage.HashToSubject[hash] = strings.Split(c.Message, "\n")[0]
			}
		}
	}
}

// FinalizeReferences updates the physical git references (branches) to point to the reconstructed history.
func FinalizeReferences(seepage *models.SeepageContext, gitSvc gitseepGit.GitService, linearHead string, branchHeads map[string]string) (string, error) {
	return gitSvc.SafeFinalize(seepage, linearHead, branchHeads)
}

// VerifyWorktreeIsClean ensures there are no uncommitted changes in the repository.
func VerifyWorktreeIsClean(gitSvc gitseepGit.GitService, opts models.Options) error {
	if opts.DryRun || opts.CheckMode {
		return nil
	}

	isClean, err := gitSvc.IsClean()
	if err != nil {
		return fmt.Errorf("failed to check worktree status: %w", err)
	}

	if !isClean && !opts.Amend {
		return fmt.Errorf("%s", i18n.T("error_worktree_dirty"))
	}

	return nil
}

func ApplyExclusions(dr *models.DiscoveryResult, excluded []string) {
	exMap := make(map[string]bool)
	for _, f := range excluded {
		exMap[f] = true
	}

	newSchedule := make(map[string]map[string][]string)
	for bedrockH, sourcesMap := range dr.Schedule {
		for srcH, files := range sourcesMap {
			var filtered []string
			for _, f := range files {
				// FIX: If the file is excluded, we ONLY keep it in its own bedrock stratum.
				// This ensures it stays anchored to its owner but doesn't receive migrations.
				if !exMap[f] || srcH == bedrockH {
					filtered = append(filtered, f)
				}
			}
			if len(filtered) > 0 {
				if newSchedule[bedrockH] == nil {
					newSchedule[bedrockH] = make(map[string][]string)
				}
				newSchedule[bedrockH][srcH] = filtered
			}
		}
	}
	dr.Schedule = newSchedule
}

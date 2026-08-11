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

// Package cmd defines the command-line interface and subcommands for GitSeep.
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/config"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/engine"
	gitseepGit "github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/git"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/i18n"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/locales"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/logger"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/models"
)

var (
	// keep-sorted start
	autoApprove       bool
	base              string
	branch            string
	dryRun            bool
	experimentalGoGit bool
	noLithify         bool
	quiet             bool
	rulesFile         string
	skipPreCommit     bool
	stageOnly         bool
	// keep-sorted end
)

var rootCmd = &cobra.Command{
	Use:   "gitseep",
	Short: "GitSeep: Geological Source Code History Percolation",
	Long: `GitSeep automates Synthetic History Reconstruction. It reconstructs your
developer branch stratum-by-stratum to perfectly map code changes to source code
Bedrock layers based on declarative .gitseep.yaml rules (autodiscovered at repo root).`,
	SilenceUsage:  true, // Don't print help on every error
	SilenceErrors: true, // We handle our own logging
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := i18n.Init("", "en", locales.Content); err != nil {
			return fmt.Errorf("failed to initialize i18n: %w", err)
		}
		logger.Init(logger.LoggerOptions{Quiet: quiet})

		if experimentalGoGit && !autoApprove && !quiet {
			logger.Warn("%s", i18n.T("experimental_warning"))
			logger.Warn("%s", i18n.T("experimental_danger"))
			fmt.Print(i18n.T("proceed_prompt"))
			reader := bufio.NewReader(os.Stdin)
			response, _ := reader.ReadString('\n')
			response = strings.TrimSpace(strings.ToUpper(response))
			if response != "YES" {
				logger.Info("%s", i18n.T("aborted"))
				return nil
			}
		}

		// Parse configuration
		cfg, err := config.Load(rulesFile)
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}

		baseRef := base
		if baseRef == "" {
			baseRef = cfg.Global.BaseRef
		}

		// HEADLESS MODE LOGIC
		// -q (Quiet) or -y (AutoApprove) triggers Headless execution.
		// Amend is now the mandatory, non-configurable behavior.
		isHeadless := quiet || autoApprove
		isAutoApprove := quiet || autoApprove

		opts := models.Options{
			TargetBranch:      branch,
			BaseCommit:        baseRef,
			AutoApprove:       isAutoApprove,
			Headless:          isHeadless,
			NoLithify:         noLithify,
			StageOnly:         stageOnly,
			DryRun:            dryRun,
			Amend:             true,
			Quiet:             quiet,
			ExperimentalGoGit: experimentalGoGit,
			SkipPreCommit:     skipPreCommit,
		}

		return engine.Run(cfg, opts)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Validate GitSeep stratigraphy and workspace cleanliness (Pre-commit hook)",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := i18n.Init("", "en", locales.Content); err != nil {
			return fmt.Errorf("failed to initialize i18n: %w", err)
		}
		logger.Init(logger.LoggerOptions{Quiet: quiet})

		gitSvc, err := gitseepGit.NewGitService(".", experimentalGoGit)
		if err != nil {
			return err
		}

		cfg, err := config.Load(rulesFile)
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}

		baseRef := base
		if baseRef == "" {
			baseRef = cfg.Global.BaseRef
		}

		opts := models.Options{
			TargetBranch:      branch,
			BaseCommit:        baseRef,
			Amend:             true,
			Quiet:             quiet,
			SkipPreCommit:     skipPreCommit,
			CheckMode:         true,
			ExperimentalGoGit: experimentalGoGit,
		}

		// Pre-flight safety check
		if err := engine.VerifyWorktreeIsClean(gitSvc, opts); err != nil {
			return err
		}

		return engine.Check(cfg, opts)
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)

	// keep-sorted start
	rootCmd.PersistentFlags().BoolVar(&experimentalGoGit, "experimental-go-git", false, "Use buggy go-git for status/checkout instead of system git")
	rootCmd.PersistentFlags().BoolVar(&noLithify, "no-lithify", false, "Prevent squashing multiple historical versions of a file")
	rootCmd.PersistentFlags().BoolVar(&noLithify, "no-squash", false, "Alias for --no-lithify")
	rootCmd.PersistentFlags().BoolVar(&skipPreCommit, "skip-pre-commit", false, "Skip isolated pre-commit validation phase")
	rootCmd.PersistentFlags().BoolVar(&stageOnly, "stage-only", false, "Leave result on temporary branch for inspection")
	rootCmd.PersistentFlags().BoolVarP(&autoApprove, "auto-approve", "y", false, "Skip interactive confirmations")
	rootCmd.PersistentFlags().BoolVarP(&dryRun, "dry-run", "n", false, "Simulate the refactor without modifying history")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "Suppress non-error output (useful with automation)")
	rootCmd.PersistentFlags().StringVar(&base, "base", "", "Base commit to start refactor from (defaults to oldest bedrock)")
	rootCmd.PersistentFlags().StringVarP(&branch, "branch", "b", "", "Target branch name (defaults to current)")
	rootCmd.PersistentFlags().StringVarP(&rulesFile, "config", "c", "", "Path to YAML seepage rules file (autodiscovered at repo root if omitted)")
	// keep-sorted end

	if err := rootCmd.PersistentFlags().MarkHidden("experimental-go-git"); err != nil {
		logger.Warn("%s", i18n.TF("error_hide_experimental_flag", map[string]interface{}{"Error": err}))
	}
}

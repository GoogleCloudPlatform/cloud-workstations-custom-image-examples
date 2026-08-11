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
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/gitutil"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/i18n"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/models"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/sys"
)

// ValidationRunner defines the interface for executing geological validation tasks.
type ValidationRunner interface {
	Setup() error
	Run(hash string, branch string, phase models.PipelinePhase) (logOutput string, changedFiles []string, skipped bool, troubleshoot *models.TroubleshootMetadata, err error)
	Cleanup() error
	WorktreeDir() string
}

// PreCommitRunner implements ValidationRunner using the 'pre-commit' tool and isolated worktrees.
type PreCommitRunner struct {
	repoRoot     string
	repo         *git.Repository
	originalHead string
	events       chan<- models.PipelineEvent
	tempDir      string
	initialized  bool
	shell        sys.ShellService
}

var _ ValidationRunner = (*PreCommitRunner)(nil)

// PreCommitParams defines the configuration and dependencies for the pre-commit runner.
type PreCommitParams struct {
	RepoRoot     string                      // Absolute path to the repository root.
	Repo         *git.Repository             // Shared go-git repository instance.
	OriginalHead string                      // The original HEAD hash before reconstruction.
	Events       chan<- models.PipelineEvent // Channel for broadcasting pipeline events.
	Shell        sys.ShellService            // Dependency-injected shell execution service.
}

// NewPreCommitRunner creates a new validator instance with the provided shell service.
func NewPreCommitRunner(p PreCommitParams) *PreCommitRunner {
	if p.Shell == nil {
		p.Shell = &sys.RealShellService{}
	}
	return &PreCommitRunner{
		repoRoot:     p.RepoRoot,
		repo:         p.Repo,
		originalHead: p.OriginalHead,
		events:       p.Events,
		shell:        p.Shell,
	}
}

func (r *PreCommitRunner) WorktreeDir() string {
	return r.tempDir
}

func (r *PreCommitRunner) Setup() error {
	if r.initialized {
		return nil
	}

	r.tempDir = filepath.Join(os.TempDir(), fmt.Sprintf("gitseep-validation-%d", time.Now().UnixNano()))

	cmdAdd := r.shell.Command("git", "worktree", "add", "--detach", r.tempDir, r.originalHead)
	cmdAdd.SetDir(r.repoRoot)

	// RETRY LOGIC: Worktree operations can fail transiently if the repo was just mutated (e.g. Amend)
	var lastOut []byte
	var err error
	for i := 0; i < 3; i++ {
		lastOut, err = cmdAdd.CombinedOutput()
		if err == nil {
			break
		}
		if i < 2 {
			time.Sleep(100 * time.Millisecond)
		}
	}

	if err != nil {
		return fmt.Errorf("failed to create isolated worktree: %w\nOutput: %s", err, string(lastOut))
	}

	r.initialized = true
	return nil
}

func (r *PreCommitRunner) Cleanup() error {
	if !r.initialized || r.tempDir == "" {
		return nil
	}

	cmdRm := r.shell.Command("git", "worktree", "remove", "--force", r.tempDir)
	cmdRm.SetDir(r.repoRoot)
	err := cmdRm.Run()
	if err == nil {
		r.initialized = false
		r.tempDir = ""
	}
	return err
}

func (r *PreCommitRunner) Run(hashStr string, branch string, phase models.PipelinePhase) (logOutput string, changedFiles []string, skipped bool, troubleshoot *models.TroubleshootMetadata, err error) {
	if !r.initialized {
		if err := r.Setup(); err != nil {
			return "", nil, false, nil, err
		}
	}

	dir := r.tempDir

	cmdCheckout := r.shell.Command("git", "checkout", "-f", "-q", hashStr)
	cmdCheckout.SetDir(dir)

	// RETRY LOGIC: Checkout operations can fail transiently if the repo was just mutated
	for i := 0; i < 3; i++ {
		err = cmdCheckout.Run()
		if err == nil {
			break
		}
		if i < 2 {
			time.Sleep(100 * time.Millisecond)
		}
	}

	if err != nil {
		troubleshoot = &models.TroubleshootMetadata{
			Category:        models.FailureCheckout,
			WorktreeDir:     dir,
			ReproductionCmd: fmt.Sprintf("git checkout %s", hashStr),
		}
		return "", nil, false, troubleshoot, fmt.Errorf("%s", i18n.ErrorCheckoutFailed(hashStr, err.Error()))
	}

	configPath := filepath.Join(dir, ".pre-commit-config.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return "", nil, true, nil, nil
	}

	sourceConfig := filepath.Join(r.repoRoot, ".pre-commit-config.yaml")
	input, err := os.ReadFile(sourceConfig)
	if err == nil {
		_ = os.WriteFile(configPath, input, 0644)
	}

	toolDirs := []string{"skills"}
	for _, td := range toolDirs {
		src := filepath.Join(r.repoRoot, td)
		if _, err := os.Stat(src); err == nil {
			_ = r.shell.Command("cp", "-r", src, dir).Run()
		}
	}

	syncPaths := []string{
		"examples/preflight-web/node_modules",
		"examples/preflight-web/dist",
	}
	for _, p := range syncPaths {
		src := filepath.Join(r.repoRoot, p)
		dst := filepath.Join(dir, p)
		if _, err := os.Stat(src); err == nil {
			_ = os.MkdirAll(filepath.Dir(dst), 0755)
			_ = os.Symlink(src, dst)
		}
	}

	commit, err := r.repo.CommitObject(plumbing.NewHash(hashStr))
	if err != nil {
		return "", nil, false, nil, err
	}
	commitDate := commit.Committer.When.Format("2006-01-02 15:04")
	commitMsg := strings.Split(commit.Message, "\n")[0]

	var parentHash string
	if commit.NumParents() > 0 {
		parentHash = commit.ParentHashes[0].String()
	}

	if parentHash != "" {
		changedFiles, err = gitutil.GetChangedFiles(r.repo, parentHash, hashStr)
		if err != nil {
			return "", nil, false, nil, err
		}
	} else {
		tree, _ := commit.Tree()
		filesMap, _ := gitutil.GetAllEntries(r.repo, tree)
		for f := range filesMap {
			changedFiles = append(changedFiles, f)
		}
	}

	if len(changedFiles) == 0 {
		return "", nil, false, nil, nil
	}

	args := []string{"run", "--files"}
	args = append(args, changedFiles...)
	args = append(args, "-c", configPath)

	cmdTest := r.shell.Command("pre-commit", args...)
	cmdTest.SetDir(dir)
	cmdTest.SetEnv(append(os.Environ(), "SKIP=gitseep-check"))

	pr, pw := io.Pipe()
	cmdTest.SetStdout(pw)
	cmdTest.SetStderr(pw)

	var fullOutput strings.Builder
	done := make(chan struct{})

	go func() {
		defer close(done)
		scanner := bufio.NewScanner(pr)
		for scanner.Scan() {
			line := scanner.Text()
			fullOutput.WriteString(line + "\n")
			if r.events != nil {
				r.events <- models.PipelineEvent{
					Phase:      phase,
					Type:       models.EventDiagnosticLine,
					Commit:     hashStr,
					CommitDate: commitDate,
					CommitMsg:  commitMsg,
					Branch:     branch,
					Message:    line,
				}
			}
		}
	}()

	err = cmdTest.Start()
	if err != nil {
		pw.Close()
		<-done
		return "", nil, false, nil, err
	}

	cmdErr := cmdTest.Wait()
	pw.Close()
	<-done

	outStr := fullOutput.String()
	if cmdErr != nil {
		cwd, _ := os.Getwd()
		troubleshoot = &models.TroubleshootMetadata{
			Category:        models.FailurePreCommit,
			LogOutput:       outStr,
			WorktreeDir:     dir,
			ReproductionCmd: fmt.Sprintf("git diff-tree --no-commit-id --name-only -r HEAD | %s xargs pre-commit run --files", models.GitSeepSkipEnv),
		}
		if dir != r.repoRoot {
			troubleshoot.CleanupCmd = fmt.Sprintf("git worktree remove --force %s && cd %s", dir, cwd)
		}

		return outStr, changedFiles, false, troubleshoot, fmt.Errorf("%s", i18n.T("error_precommit_failed"))
	}
	return outStr, changedFiles, false, nil, nil
}

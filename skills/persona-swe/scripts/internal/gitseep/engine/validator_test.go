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
	"errors"
	"io"
	"testing"

	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/models"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/sys"
)

type failingShellService struct {
	attempts int
	maxFail  int
}

func (m *failingShellService) Command(name string, arg ...string) sys.ShellCommand {
	return &failingShellCommand{svc: m}
}

type failingShellCommand struct {
	svc *failingShellService
}

func (c *failingShellCommand) SetDir(dir string)     {}
func (c *failingShellCommand) SetEnv(env []string)   {}
func (c *failingShellCommand) SetStdout(w io.Writer) {}
func (c *failingShellCommand) SetStderr(w io.Writer) {}
func (c *failingShellCommand) CombinedOutput() ([]byte, error) {
	c.svc.attempts++
	if c.svc.attempts <= c.svc.maxFail {
		return []byte("error"), errors.New("transient error")
	}
	return []byte("ok"), nil
}
func (c *failingShellCommand) Run() error {
	c.svc.attempts++
	if c.svc.attempts <= c.svc.maxFail {
		return errors.New("transient error")
	}
	return nil
}
func (c *failingShellCommand) Start() error                       { return nil }
func (c *failingShellCommand) Wait() error                        { return nil }
func (c *failingShellCommand) StdoutPipe() (io.ReadCloser, error) { return nil, nil }

func TestPreCommitRunner_RetryLogic(t *testing.T) {
	// 1. Test Setup (git worktree add) retries
	mock := &failingShellService{maxFail: 2} // Should succeed on 3rd attempt
	runner := NewPreCommitRunner(PreCommitParams{
		RepoRoot:     ".",
		OriginalHead: "HEAD",
		Shell:        mock,
	})

	err := runner.Setup()
	if err != nil {
		t.Fatalf("Setup should have succeeded after retries: %v", err)
	}
	if mock.attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", mock.attempts)
	}

	// 2. Test Run (git checkout) retries
	mock.attempts = 0
	mock.maxFail = 1          // Succeed on 2nd attempt
	runner.initialized = true // Skip Setup

	_, _, skipped, _, err := runner.Run("hash", "branch", models.PhaseValidation)
	if err != nil {
		t.Errorf("Run failed: %v", err)
	}
	if !skipped {
		t.Error("Run should have been skipped due to missing pre-commit config")
	}
	if mock.attempts != 2 {
		t.Errorf("Expected 2 attempts for checkout, got %d", mock.attempts)
	}
}

func TestPreCommitRunner_ShellDI(t *testing.T) {
	mock := &mockShellService{}

	runner := NewPreCommitRunner(PreCommitParams{
		RepoRoot:     ".",
		OriginalHead: "HEAD",
		Shell:        mock,
	})

	err := runner.Setup()
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	if mock.lastCmd != "git" {
		t.Errorf("Expected last command to be git, got %s", mock.lastCmd)
	}
}

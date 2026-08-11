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

package cmd

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestExecute(t *testing.T) {
	// Simple smoke test for help
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"gitseep", "--help"}

	err := Execute()
	if err != nil {
		t.Fatalf("Expected Execute to return nil for --help, got %v", err)
	}
}

func setupTestRepo(t *testing.T) (string, func()) {
	t.Helper()
	dir, err := os.MkdirTemp("", "gitseep-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	oldWd, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Failed to change directory to %s: %v", dir, err)
	}

	_ = exec.Command("git", "init", "-b", "main").Run()
	_ = exec.Command("git", "config", "user.email", "test@example.com").Run()
	_ = exec.Command("git", "config", "user.name", "Test User").Run()

	cleanup := func() {
		if err := os.Chdir(oldWd); err != nil {
			t.Logf("Warning: failed to change directory back to %s: %v", oldWd, err)
		}
		os.RemoveAll(dir)
	}

	return dir, cleanup
}

func TestCheckCommand(t *testing.T) {
	t.Run("Clean Check Success", func(t *testing.T) {
		_, cleanup := setupTestRepo(t)
		defer cleanup()

		_ = os.WriteFile("a.txt", []byte("a"), 0644)
		_ = exec.Command("git", "add", "a.txt").Run()
		_ = exec.Command("git", "commit", "-m", "msgA").Run()
		dateABytes, _ := exec.Command("git", "show", "-s", "--format=%ai", "HEAD").Output()

		rules := "config:\n  base_ref: main\n" +
			"\"" + string(dateABytes)[:10] + "\":\n  - a.txt\n"
		_ = os.WriteFile(".gitseep.yaml", []byte(rules), 0644)
		_ = exec.Command("git", "add", ".gitseep.yaml").Run()
		_ = exec.Command("git", "commit", "-m", "add rules").Run()

		rootCmd.SetArgs([]string{"check"})
		err := rootCmd.Execute()
		if err != nil {
			t.Errorf("Expected check to pass, got error: %v", err)
		}
	})

	t.Run("Dirty Worktree Failure", func(t *testing.T) {
		_, cleanup := setupTestRepo(t)
		defer cleanup()

		_ = os.WriteFile("a.txt", []byte("a"), 0644)
		_ = exec.Command("git", "add", "a.txt").Run()
		_ = exec.Command("git", "commit", "-m", "msg").Run()

		// Make it dirty
		_ = os.WriteFile("a.txt", []byte("dirty"), 0644)

		rootCmd.SetArgs([]string{"check"})
		err := rootCmd.Execute()
		if err == nil {
			t.Errorf("Expected check to fail with dirty worktree")
		}
	})

	t.Run("Unsedimented History Failure", func(t *testing.T) {
		_, cleanup := setupTestRepo(t)
		defer cleanup()

		// Commit A: Bedrock (Yesterday)
		_ = os.WriteFile("a.txt", []byte("a"), 0644)
		_ = exec.Command("git", "add", "a.txt").Run()
		cmd := exec.Command("git", "commit", "-m", "bedrock")
		cmd.Env = append(os.Environ(), "GIT_AUTHOR_DATE=2026-04-20T12:00:00", "GIT_COMMITTER_DATE=2026-04-20T12:00:00")
		_ = cmd.Run()
		dateA := "2026-04-20"

		// Capture Commit A's hash to use as base_ref
		hashABytes, err := exec.Command("git", "rev-parse", "HEAD").Output()
		if err != nil {
			t.Fatalf("Failed to get hash of Commit A: %v", err)
		}
		hashA := strings.TrimSpace(string(hashABytes))

		// Commit B: Rules (Today)
		rules := "config:\n  base_ref: " + hashA + "\n" +
			"\"" + dateA + "\":\n  - a.txt\n"
		_ = os.WriteFile(".gitseep.yaml", []byte(rules), 0644)
		_ = exec.Command("git", "add", ".gitseep.yaml").Run()
		cmd = exec.Command("git", "commit", "-m", "add rules")
		cmd.Env = append(os.Environ(), "GIT_AUTHOR_DATE=2026-04-21T12:00:00", "GIT_COMMITTER_DATE=2026-04-21T12:00:00")
		_ = cmd.Run()

		// Commit C: Alluvium (Violates rules)
		_ = os.WriteFile("a.txt", []byte("alluvial-change"), 0644)
		_ = exec.Command("git", "add", "a.txt").Run()
		cmd = exec.Command("git", "commit", "-m", "this is alluvium")
		cmd.Env = append(os.Environ(), "GIT_AUTHOR_DATE=2026-04-22T12:00:00", "GIT_COMMITTER_DATE=2026-04-22T12:00:00")
		_ = cmd.Run()

		rootCmd.SetArgs([]string{"check"})
		err = rootCmd.Execute()
		if err == nil {
			t.Errorf("Expected check to fail with unsedimented history")
		}
	})
}

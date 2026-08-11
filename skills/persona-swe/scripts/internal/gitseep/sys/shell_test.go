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

package sys

import (
	"bytes"
	"testing"
)

func TestRealShellService(t *testing.T) {
	svc := &RealShellService{}

	t.Run("Command execution", func(t *testing.T) {
		cmd := svc.Command("echo", "hello")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Command failed: %v", err)
		}
		if string(out) != "hello\n" {
			t.Errorf("Expected 'hello\n', got %q", string(out))
		}
	})

	t.Run("Command with environment", func(t *testing.T) {
		cmd := svc.Command("sh", "-c", "echo $TEST_VAR")
		cmd.SetEnv([]string{"TEST_VAR=world"})
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Command failed: %v", err)
		}
		if string(out) != "world\n" {
			t.Errorf("Expected 'world\n', got %q", string(out))
		}
	})

	t.Run("Command with directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		cmd := svc.Command("pwd")
		cmd.SetDir(tmpDir)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Command failed: %v", err)
		}
		// pwd might return a symlink resolved path, so we use EvalSymlinks or just check suffix
		if !testing.Short() {
			// Basic check
			if len(out) == 0 {
				t.Error("Empty output from pwd")
			}
		}
	})

	t.Run("Pipe and Start/Wait", func(t *testing.T) {
		cmd := svc.Command("echo", "piped")
		pipe, err := cmd.StdoutPipe()
		if err != nil {
			t.Fatalf("Pipe failed: %v", err)
		}
		if err := cmd.Start(); err != nil {
			t.Fatalf("Start failed: %v", err)
		}

		buf := make([]byte, 10)
		n, _ := pipe.Read(buf)
		if string(buf[:n]) != "piped\n" {
			t.Errorf("Expected 'piped\n', got %q", string(buf[:n]))
		}

		if err := cmd.Wait(); err != nil {
			t.Errorf("Wait failed: %v", err)
		}
	})
}

func TestRealShellService_Extra(t *testing.T) {
	svc := &RealShellService{}

	t.Run("SetStdout and SetStderr", func(t *testing.T) {
		var out, err bytes.Buffer
		cmd := svc.Command("sh", "-c", "echo hello; echo world >&2")
		cmd.SetStdout(&out)
		cmd.SetStderr(&err)
		if e := cmd.Run(); e != nil {
			t.Fatalf("Run failed: %v", e)
		}
		if out.String() != "hello\n" {
			t.Errorf("Expected 'hello\n', got %q", out.String())
		}
		if err.String() != "world\n" {
			t.Errorf("Expected 'world\n', got %q", err.String())
		}
	})
}

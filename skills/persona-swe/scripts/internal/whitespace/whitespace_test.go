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

package whitespace

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCleanFile(t *testing.T) {
	tmp, err := os.MkdirTemp("", "whitespace-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmp)

	path := filepath.Join(tmp, "dirty.txt")
	content := "line1  \nline2\t \n\nline3\n"
	expected := "line1\nline2\n\nline3\n"

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	modified, err := CleanFile(path)
	if err != nil {
		t.Fatalf("CleanFile failed: %v", err)
	}
	if !modified {
		t.Errorf("Expected file to be modified")
	}

	cleaned, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read cleaned file: %v", err)
	}

	if string(cleaned) != expected {
		t.Errorf("Expected cleaned content %q, got %q", expected, string(cleaned))
	}
}

func TestCleanFile_EdgeCases(t *testing.T) {
	tmp := t.TempDir()

	tests := []struct {
		name     string
		content  string
		expected string
		modified bool
	}{
		{
			name:     "Empty file",
			content:  "",
			expected: "",
			modified: false,
		},
		{
			name:     "Only whitespace",
			content:  "   \n  \t ",
			expected: "",
			modified: true,
		},
		{
			name:     "Already clean",
			content:  "line1\nline2\n",
			expected: "line1\nline2\n",
			modified: false,
		},
		{
			name:     "No trailing newline",
			content:  "line1\nline2",
			expected: "line1\nline2\n",
			modified: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(tmp, tt.name)
			_ = os.WriteFile(path, []byte(tt.content), 0644)

			modified, err := CleanFile(path)
			if err != nil {
				t.Fatalf("CleanFile failed: %v", err)
			}
			if modified != tt.modified {
				t.Errorf("Expected modified=%v, got %v", tt.modified, modified)
			}

			data, _ := os.ReadFile(path)
			if string(data) != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, string(data))
			}
		})
	}
}

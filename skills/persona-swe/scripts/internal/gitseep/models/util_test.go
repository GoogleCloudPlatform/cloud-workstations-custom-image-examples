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

package models

import "testing"

func TestEnsureSkipEnv(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Already skipped",
			input:    "SKIP=gitseep-check pre-commit run",
			expected: "SKIP=gitseep-check pre-commit run",
		},
		{
			name:     "No pre-commit",
			input:    "git status",
			expected: "git status",
		},
		{
			name:     "Pre-commit without skip",
			input:    "pre-commit run --files a.txt",
			expected: "SKIP=gitseep-check pre-commit run --files a.txt",
		},
		{
			name:     "Piped pre-commit",
			input:    "git diff | xargs pre-commit run",
			expected: "SKIP=gitseep-check git diff | xargs pre-commit run",
		},
		{
			name:     "Redundant skip",
			input:    "SKIP=gitseep-check SKIP=other pre-commit",
			expected: "SKIP=gitseep-check SKIP=other pre-commit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EnsureSkipEnv(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

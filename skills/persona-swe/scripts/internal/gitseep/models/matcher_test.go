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

// Package models_test provides integration tests for the GitSeep data models.
package models_test

import (
	"testing"

	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/models"
)

func TestGeologicalMatcher(t *testing.T) {
	pathToBedrock := map[string]string{
		"src/":      "hash-src",
		"src/core/": "hash-core",
		"README.md": "hash-readme",
	}

	matcher := models.NewGeologicalMatcher(pathToBedrock)

	tests := []struct {
		name        string
		filePath    string
		wantBedrock string
		wantRule    string
	}{
		{
			name:        "Exact match",
			filePath:    "README.md",
			wantBedrock: "hash-readme",
			wantRule:    "README.md",
		},
		{
			name:        "Prefix match",
			filePath:    "src/main.go",
			wantBedrock: "hash-src",
			wantRule:    "src/",
		},
		{
			name:        "Longest match wins",
			filePath:    "src/core/api.go",
			wantBedrock: "hash-core",
			wantRule:    "src/core/",
		},
		{
			name:        "No match",
			filePath:    "docs/index.md",
			wantBedrock: "",
			wantRule:    "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotBedrock, gotRule := matcher.ResolveTarget(tc.filePath)
			if gotBedrock != tc.wantBedrock {
				t.Errorf("ResolveTarget(%q) bedrock = %q, want %q", tc.filePath, gotBedrock, tc.wantBedrock)
			}
			if gotRule != tc.wantRule {
				t.Errorf("ResolveTarget(%q) rule = %q, want %q", tc.filePath, gotRule, tc.wantRule)
			}
		})
	}
}

func TestBelongsToPath(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		rulePath string
		want     bool
	}{
		{"Exact file match", "src/main.go", "src/main.go", true},
		{"Directory prefix match", "src/main.go", "src/", true},
		{"Directory without trailing slash match", "src/main.go", "src", true},
		{"Exact directory without trailing slash match", "src", "src", true},
		{"Prefix mismatch", "src2/main.go", "src/", false},
		{"Prefix mismatch no trailing slash", "src2", "src", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := models.BelongsToPath(tc.filePath, tc.rulePath)
			if got != tc.want {
				t.Errorf("BelongsToPath(%q, %q) = %v, want %v", tc.filePath, tc.rulePath, got, tc.want)
			}
		})
	}
}

func TestIndexOf(t *testing.T) {
	slice := []string{"a", "b", "c"}

	if i := models.IndexOf(slice, "b"); i != 1 {
		t.Errorf("Expected index 1, got %d", i)
	}

	if i := models.IndexOf(slice, "z"); i != -1 {
		t.Errorf("Expected index -1, got %d", i)
	}
}

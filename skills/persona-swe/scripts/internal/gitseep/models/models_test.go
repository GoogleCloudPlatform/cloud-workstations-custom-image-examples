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

import (
	"testing"
)

func TestDiscoveryResult_HasMigrations(t *testing.T) {
	tests := []struct {
		name     string
		schedule map[string]map[string][]string
		expected bool
	}{
		{
			name:     "Empty schedule",
			schedule: nil,
			expected: false,
		},
		{
			name: "Only bedrock sources",
			schedule: map[string]map[string][]string{
				"h1": {"h1": {"f1.txt"}},
			},
			expected: false,
		},
		{
			name: "Different source hash (Migration)",
			schedule: map[string]map[string][]string{
				"h1": {"h2": {"f1.txt"}},
			},
			expected: true,
		},
		{
			name: "Multiple sources for one bedrock",
			schedule: map[string]map[string][]string{
				"h1": {
					"h1": {"f1.txt"},
					"h2": {"f1.txt"},
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dr := &DiscoveryResult{Schedule: tt.schedule}
			if got := dr.HasMigrations(); got != tt.expected {
				t.Errorf("HasMigrations() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestSeepageContext_GetReconstructedStrata(t *testing.T) {
	c := &SeepageContext{
		Strata: []string{"h1", "h2", "h3"},
		OriginalToReconstructed: map[string]string{
			"h1": "new1",
			"h2": "new2",
		},
	}
	// h3 is not mapped, so it should be skipped in reconstructed list
	res := c.GetReconstructedStrata()
	if len(res) != 2 {
		t.Errorf("Expected 2 reconstructed strata, got %d", len(res))
	}
	if res[0] != "new1" || res[1] != "new2" {
		t.Errorf("Unexpected reconstructed strata order or values: %v", res)
	}
}

func TestModels_Coverage(t *testing.T) {
	t.Run("ResolveBedrockForPath", func(t *testing.T) {
		seepage := &SeepageContext{
			Matcher: NewGeologicalMatcher(map[string]string{
				"src": "hash1",
			}),
		}
		if h := seepage.ResolveBedrockForPath("src/main.go"); h != "hash1" {
			t.Errorf("Expected hash1, got %q", h)
		}
		if h := seepage.ResolveBedrockForPath("README.md"); h != "" {
			t.Errorf("Expected empty hash for unmanaged path, got %q", h)
		}
	})

	t.Run("ResolveSubject", func(t *testing.T) {
		seepage := &SeepageContext{
			HashToSubject: map[string]string{
				"hash1": "feat: init",
			},
		}
		if s := seepage.ResolveSubject("hash1"); s != "feat: init" {
			t.Errorf("Expected 'feat: init', got %q", s)
		}
		if s := seepage.ResolveSubject("unknown"); s != "" {
			t.Errorf("Expected empty string for unknown subject, got %q", s)
		}
	})

	t.Run("GetMigrationItems", func(t *testing.T) {
		dr := &DiscoveryResult{
			Schedule: map[string]map[string][]string{
				"bedrock1": {
					"source1": {"file1.txt", "file2.txt"},
					"source2": {"file1.txt"},
				},
			},
		}
		items := dr.GetMigrationItems()
		if len(items) != 2 {
			t.Errorf("Expected 2 migration items, got %d", len(items))
		}

		foundFile1 := false
		for _, item := range items {
			if item.Path == "file1.txt" {
				foundFile1 = true
				if len(item.Sources) != 2 {
					t.Errorf("Expected 2 sources for file1.txt, got %d", len(item.Sources))
				}
			}
		}
		if !foundFile1 {
			t.Error("file1.txt not found in migration items")
		}
	})

	t.Run("ParseRuleDate", func(t *testing.T) {
		d1 := ParseRuleDate("2026-04-29")
		if d1.Year() != 2026 || d1.Month() != 4 || d1.Day() != 29 {
			t.Errorf("Failed to parse simple date: %v", d1)
		}

		d2 := ParseRuleDate("2026-04-29 15:04:05 +0000")
		if d2.Hour() != 15 || d2.Minute() != 4 {
			t.Errorf("Failed to parse full date: %v", d2)
		}

		d3 := ParseRuleDate("invalid")
		if !d3.IsZero() {
			t.Errorf("Expected zero time for invalid date, got %v", d3)
		}
	})

	t.Run("Errors", func(t *testing.T) {
		if !IsAborted(ErrPipelineAborted) {
			t.Error("IsAborted(ErrPipelineAborted) should be true")
		}
		if IsAborted(nil) {
			t.Error("IsAborted(nil) should be false")
		}
	})
}

func TestDiscoveryResult_GetSummary(t *testing.T) {
	seepage := &SeepageContext{
		Strata: []string{"h1", "h2", "h3"}, // h1 (oldest) -> h3 (newest)
	}

	dr := &DiscoveryResult{
		Schedule: map[string]map[string][]string{
			"h1": {
				"h2": {"file1"}, // Seep (h2 -> h1)
			},
			"h3": {
				"h1": {"file2"}, // Percolate (h1 -> h3)
			},
		},
		Sources: map[string]map[string]struct{}{
			"file3": {
				"h1": {},
				"h2": {}, // Lithification
			},
		},
	}

	// Add file3 to schedule so it's processed
	dr.Schedule["h1"]["h1"] = append(dr.Schedule["h1"]["h1"], "file3")
	dr.Schedule["h1"]["h2"] = append(dr.Schedule["h1"]["h2"], "file3")

	summary := dr.GetSummary(seepage)

	if summary.SeepFiles != 1 {
		t.Errorf("Expected 1 seep, got %d", summary.SeepFiles)
	}
	if summary.PercolateFiles != 2 {
		t.Errorf("Expected 2 percolates, got %d", summary.PercolateFiles)
	}
	if len(summary.LithifiedFiles) != 1 {
		t.Errorf("Expected 1 lithified file, got %d", len(summary.LithifiedFiles))
	}
	if _, ok := summary.LithifiedFiles["file3"]; !ok {
		t.Error("Expected file3 to be lithified")
	}
}

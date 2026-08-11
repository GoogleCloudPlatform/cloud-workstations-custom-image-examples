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

package i18n

import (
	"embed"
	"testing"
	"testing/fstest"

	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/locales"
)

//go:embed test_locales/*.json
var testLocales embed.FS

func TestI18n(t *testing.T) {
	// We use the real locales if possible, or just mock it.
	// For testing the package itself, we can try to Init with empty or minimal FS.
	_ = Init("en", "en", testLocales)

	t.Run("Translate missing key", func(t *testing.T) {
		msg := T("non_existent_key")
		if msg != "non_existent_key" {
			t.Errorf("Expected key itself for missing translation, got %q", msg)
		}
	})

	t.Run("Helper functions", func(t *testing.T) {
		if s := SelectionStatus(1, 10, 5); s == "" {
			t.Error("SelectionStatus returned empty string")
		}
		if s := SummaryStatus(5, 5); s == "" {
			t.Error("SummaryStatus returned empty string")
		}
		if s := ValidationStatus("bar", 50, 1, 2); s == "" {
			t.Error("ValidationStatus returned empty string")
		}
		if s := HeadlessSummaryStatus(10); s == "" {
			t.Error("HeadlessSummaryStatus returned empty string")
		}
		if s := ScanningHistory(1, 1, "hash", "subj"); s == "" {
			t.Error("ScanningHistory returned empty string")
		}
		if s := ReconstructingStrata(1, 1, "hash", "subj"); s == "" {
			t.Error("ReconstructingStrata returned empty string")
		}
		if s := SedimentingBranch(1, 1, "feat"); s == "" {
			t.Error("SedimentingBranch returned empty string")
		}
		if s := ValidatingFeatureBranch(1, 1, "hash", "feat"); s == "" {
			t.Error("ValidatingFeatureBranch returned empty string")
		}
		if s := ValidatingLinearStratum(1, 1, "hash"); s == "" {
			t.Error("ValidatingLinearStratum returned empty string")
		}
		if s := ErrorPrecommitNotConfigured(); s == "" {
			t.Error("ErrorPrecommitNotConfigured returned empty string")
		}
		if s := ErrorPrecommitMissing("hash"); s == "" {
			t.Error("ErrorPrecommitMissing returned empty string")
		}
		if s := ErrorRulesNotFound("file"); s == "" {
			t.Error("ErrorRulesNotFound returned empty string")
		}
		if s := ErrorDependencyCycle("A->B"); s == "" {
			t.Error("ErrorDependencyCycle returned empty string")
		}
		if s := ErrorCheckoutFailed("hash", "err"); s == "" {
			t.Error("ErrorCheckoutFailed returned empty string")
		}
		if s := ErrorConflictPredicted("B", "H", "S", "F"); s == "" {
			t.Error("ErrorConflictPredicted returned empty string")
		}
	})
}

func TestI18n_InitErrors(t *testing.T) {
	t.Run("Empty FS", func(t *testing.T) {
		fs := fstest.MapFS{}
		err := Init("en", "en", fs)
		if err != nil {
			t.Errorf("Init should not fail on empty FS (just no translations), got %v", err)
		}
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		fs := fstest.MapFS{
			"en.json": &fstest.MapFile{Data: []byte("invalid")},
		}
		err := Init("en", "en", fs)
		if err == nil {
			t.Error("Init should fail on invalid JSON")
		}
	})
}

func TestI18n_RealLocales(t *testing.T) {
	err := Init("en", "en", locales.Content)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	t.Run("Translate real key", func(t *testing.T) {
		msg := T("tui_header")
		if msg == "" || msg == "tui_header" {
			t.Errorf("Expected translation for tui_header, got %q", msg)
		}
	})

	t.Run("Translate real key with format", func(t *testing.T) {
		msg := TF("scanning_history", map[string]interface{}{"Current": 1, "Total": 10})
		if msg == "" || msg == "scanning_history" {
			t.Errorf("Expected formatted translation, got %q", msg)
		}
	})
}

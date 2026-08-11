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
	"fmt"
)

// UI Helpers (Typed wrappers for common translation keys)

func SelectionStatus(current, total, selected int) string {
	return TF("selection_status", map[string]interface{}{
		"Current":  current,
		"Total":    total,
		"Selected": selected,
	})
}

func SummaryStatus(current, total int) string {
	return TF("summary_status", map[string]interface{}{
		"Current": current,
		"Total":   total,
	})
}

func ValidationStatus(bar string, percent int, completed, total int) string {
	return TF("validation_status", map[string]interface{}{
		"Bar":       bar,
		"Percent":   percent,
		"Completed": completed,
		"Total":     total,
	})
}

func HeadlessSummaryStatus(total int) string {
	return TF("headless_summary_status", map[string]interface{}{
		"Total": total,
	})
}

// Pipeline Progress Helpers

func ScanningHistory(current, total int, hash, subject string) string {
	shortHash := hash
	if len(hash) > 7 {
		shortHash = hash[:7]
	}
	return TF("scanning_history", map[string]interface{}{
		"Current": current,
		"Total":   total,
		"Hash":    shortHash,
		"Subject": subject,
	})
}

func ReconstructingStrata(current, total int, hash, subject string) string {
	shortHash := hash
	if len(hash) > 7 {
		shortHash = hash[:7]
	}
	return TF("reconstructing_strata", map[string]interface{}{
		"Current": current,
		"Total":   total,
		"Hash":    shortHash,
		"Subject": subject,
	})
}

func SedimentingBranch(current, total int, branch string) string {
	return TF("sedimenting_branch", map[string]interface{}{
		"Current": current,
		"Total":   total,
		"Branch":  branch,
	})
}

func ValidatingSurface(hash string) string {
	shortHash := hash
	if len(hash) > 7 {
		shortHash = hash[:7]
	}
	return TF("validating_surface", map[string]interface{}{"Hash": shortHash})
}

func ValidatingFeatureBranch(current, total int, hash, branch string) string {
	shortHash := hash
	if len(hash) > 7 {
		shortHash = hash[:7]
	}
	return TF("validating_feature_branch", map[string]interface{}{
		"Current": current,
		"Total":   total,
		"Hash":    shortHash,
		"Branch":  branch,
	})
}

func ValidatingLinearStratum(current, total int, hash string) string {
	shortHash := hash
	if len(hash) > 7 {
		shortHash = hash[:7]
	}
	return TF("validating_linear_stratum", map[string]interface{}{
		"Current": current,
		"Total":   total,
		"Hash":    shortHash,
	})
}

func HistoryStaged(branch string) string {
	return TF("history_staged", map[string]interface{}{"Branch": branch})
}

// GeologicalReportLines returns the formatted lines for the session summary.
func GeologicalReportLines(seeps, percolations, lithifications int) []string {
	return []string{
		fmt.Sprintf("  🫧  %s:          %d", T("report_seeps"), seeps),
		fmt.Sprintf("  💧  %s:   %d", T("report_percolations"), percolations),
		fmt.Sprintf("  🪨  %s: %d", T("report_lithifications"), lithifications),
	}
}

// Error Helpers

func ErrorPrecommitNotConfigured() string {
	return T("error_precommit_not_configured")
}

func ErrorPrecommitMissing(hash string) string {
	shortHash := hash
	if len(hash) > 7 {
		shortHash = hash[:7]
	}
	return TF("error_precommit_missing_config", map[string]interface{}{"Hash": shortHash})
}

func ErrorRulesNotFound(file string) string {
	return TF("error_rules_not_found", map[string]interface{}{"File": file})
}

func ErrorDependencyCycle(cycle string) string {
	return TF("error_dependency_cycle", map[string]interface{}{"Cycle": cycle})
}

func ErrorCheckoutFailed(hash, err string) string {
	return TF("error_checkout_failed", map[string]interface{}{"Hash": hash, "Error": err})
}

func ErrorConflictPredicted(branch, bedrock, stratum, files string) string {
	return TF("error_conflict_predicted", map[string]interface{}{
		"Branch":  branch,
		"Bedrock": bedrock,
		"Stratum": stratum,
		"Files":   files,
	})
}

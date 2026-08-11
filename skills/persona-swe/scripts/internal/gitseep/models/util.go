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

// Package models defines shared data structures and utilities for GitSeep.
package models

import (
	"strings"
	"time"
)

// GitSeepSkipEnv is the environment variable prefix used to skip the gitseep-check
// pre-commit hook during manual reproduction or automated validation.
const GitSeepSkipEnv = "SKIP=gitseep-check"

// EnsureSkipEnv ensures that the pre-commit reproduction command includes the
// necessary environment variable to skip recursive gitseep checks.
func EnsureSkipEnv(cmd string) string {
	if !strings.Contains(cmd, "pre-commit") || strings.Contains(cmd, GitSeepSkipEnv) {
		return cmd
	}
	return GitSeepSkipEnv + " " + cmd
}

// IndexOf returns the index of a string in a slice, or -1 if not found.
func IndexOf(slice []string, val string) int {
	for i, item := range slice {
		if item == val {
			return i
		}
	}
	return -1
}

// ParseRuleDate parses a GitSeep rule date string into a time.Time object.
func ParseRuleDate(dateStr string) time.Time {
	// Try parsing full ISO format
	if len(dateStr) > 10 {
		t, err := time.Parse("2006-01-02 15:04:05 -0700", dateStr)
		if err == nil {
			return t
		}
	}
	// Try YYYY-MM-DD
	t, err := time.Parse("2006-01-02", dateStr)
	if err == nil {
		return t
	}
	return time.Time{}
}

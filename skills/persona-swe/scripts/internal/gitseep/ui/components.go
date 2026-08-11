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

package ui

// This file is now primarily for shared UI logic that doesn't fit into
// specific sub-models or layout components.

func (m *DashboardModel) GetTroubleshootDir() string {
	if m.failure.failure != nil && m.failure.failure.Troubleshoot != nil {
		return m.failure.failure.Troubleshoot.WorktreeDir
	}
	return ""
}

func (m *DashboardModel) GetTroubleshootLog() string {
	if m.failure.failure != nil && m.failure.failure.Troubleshoot != nil {
		return m.failure.failure.Troubleshoot.LogOutput
	}
	return ""
}

func (m *DashboardModel) GetTroubleshootManualCmd() string {
	if m.failure.failure != nil && m.failure.failure.Troubleshoot != nil {
		return m.failure.failure.Troubleshoot.ReproductionCmd
	}
	return ""
}

func (m *DashboardModel) GetTroubleshootCleanupCmd() string {
	if m.failure.failure != nil && m.failure.failure.Troubleshoot != nil {
		return m.failure.failure.Troubleshoot.CleanupCmd
	}
	return ""
}

func (m *DashboardModel) GetRepoRoot() string {
	if m.seepage != nil {
		return m.seepage.RepoRoot
	}
	return "."
}

func (m *DashboardModel) GetError() error {
	return m.err
}

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

import tea "github.com/charmbracelet/bubbletea"

// NavigationController provides a unified way to handle common keyboard navigation.
type NavigationController struct {
	cursor   *int
	maxItems int
	pageSize int
}

// NewNavigationController creates a controller linked to a specific cursor.
func NewNavigationController(cursor *int, maxItems int, pageSize int) *NavigationController {
	return &NavigationController{
		cursor:   cursor,
		maxItems: maxItems,
		pageSize: pageSize,
	}
}

// HandleKey processes common navigation keys and returns true if the key was handled.
func (n *NavigationController) HandleKey(msg tea.KeyMsg) bool {
	switch msg.String() {
	case "up", "k":
		if *n.cursor > 0 {
			*n.cursor--
		}
		return true
	case "down", "j":
		if *n.cursor < n.maxItems-1 {
			*n.cursor++
		}
		return true
	case "pgup":
		*n.cursor -= n.pageSize
		if *n.cursor < 0 {
			*n.cursor = 0
		}
		return true
	case "pgdown":
		*n.cursor += n.pageSize
		if *n.cursor >= n.maxItems {
			*n.cursor = n.maxItems - 1
		}
		if *n.cursor < 0 {
			*n.cursor = 0
		}
		return true
	case "home":
		*n.cursor = 0
		return true
	case "end":
		if n.maxItems > 0 {
			*n.cursor = n.maxItems - 1
		}
		return true
	}
	return false
}

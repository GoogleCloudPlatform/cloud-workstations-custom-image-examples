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

import (
	"strings"
	"testing"

	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/models"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/testutil"
)

func TestTreeContext_Geometry(t *testing.T) {
	// Setup a simple linear history with one feature branch
	// BASE -> H1 (main) -> H2 (feat)
	seepage := &models.SeepageContext{
		BaseHash:     "BASE",
		DateToHash:   map[string]string{"2026-01-01": "H1", "2026-01-02": "H2"},
		DateToBranch: map[string]string{"2026-01-01": "main", "2026-01-02": "feat"},
		BranchToParent: map[string]string{
			"feat": "main",
			"main": "BASE",
		},
	}

	tc := NewTreeContext(seepage)

	// 1. Check ordering (Tips to Base)
	// Now includes BASE
	if len(tc.OrderedHashes) != 3 {
		t.Fatalf("Expected 3 ordered hashes, got %d", len(tc.OrderedHashes))
	}
	if tc.OrderedHashes[0] != "H2" {
		t.Errorf("Expected first hash H2 (tip), got %s", tc.OrderedHashes[0])
	}
	if tc.OrderedHashes[2] != "BASE" {
		t.Errorf("Expected last hash BASE (root), got %s", tc.OrderedHashes[2])
	}

	// 2. Check lane assignment
	if tc.HashToColumn["BASE"] != 0 {
		t.Errorf("Expected BASE in column 0, got %d", tc.HashToColumn["BASE"])
	}
	if tc.HashToColumn["H1"] != 0 {
		t.Errorf("Expected H1 in column 0, got %d", tc.HashToColumn["H1"])
	}
	// H2 is a child of H1, should inherit lane if it's the only child
	if tc.HashToColumn["H2"] != 0 {
		t.Errorf("Expected H2 in column 0, got %d", tc.HashToColumn["H2"])
	}
}

func TestTreeContext_Rendering(t *testing.T) {
	tc := &TreeContext{
		HashToParent: map[string]string{"H1": "BASE", "H2": "H1"},
	}

	activeLanes := map[int]string{0: "H1"}

	// 1. Straight lane
	lanes := tc.RenderLanes(0, activeLanes)
	if lanes != "● " {
		t.Errorf("Expected '● ', got %q", lanes)
	}

	// 2. Anchor lanes
	anchor := tc.RenderAnchorLanes(0, activeLanes)
	if anchor != "● " {
		t.Errorf("Expected '● ', got %q", anchor)
	}

	// 3. Vertical connector
	activeLanes[0] = "H2" // Lane 0 is busy with H2
	conn := tc.RenderConnectors("H1", 0, activeLanes)
	if conn != "│ " {
		t.Errorf("Expected '│ ', got %q", conn)
	}

	// 4. Metadata lanes
	meta := tc.RenderMetaLanes(0, activeLanes)
	if meta != "│ " {
		t.Errorf("Expected '│ ', got %q", meta)
	}
}

func TestTreeContext_FullRendering(t *testing.T) {
	testutil.InitI18n()
	_, seepage, _ := testutil.SetupLinearHistory(t)
	// Add another branch to make it more complex
	seepage.DateToHash["2026-01-03"] = "h3"
	seepage.DateToBranch["2026-01-03"] = "feat2"
	seepage.BranchToParent["feat2"] = "main"
	seepage.BaseHash = "h1"

	tc := NewTreeContext(seepage)

	t.Run("RenderConnectors", func(t *testing.T) {
		activeLanes := map[int]string{1: "h3", 2: "h2"}
		// Render connectors for h1 (parent of h2 and h3)
		res := tc.RenderConnectors("h1", 0, activeLanes)
		if res == "" {
			t.Error("Empty connectors rendered")
		}
	})

	t.Run("RenderLanes", func(t *testing.T) {
		activeLanes := map[int]string{1: "h3"}
		res := tc.RenderLanes(0, activeLanes)
		if res == "" {
			t.Error("Empty lanes rendered")
		}
	})

	t.Run("RenderMetaLanes", func(t *testing.T) {
		activeLanes := map[int]string{1: "h3"}
		res := tc.RenderMetaLanes(0, activeLanes)
		if res == "" {
			t.Error("Empty meta lanes rendered")
		}
	})
}

func TestTreeContext_LaneRendering(t *testing.T) {
	seepage := &models.SeepageContext{
		DateToBranch: map[string]string{
			"2026-01-01": "main",
			"2026-01-02": "feat1",
			"2026-01-03": "feat2",
		},
		DateToHash: map[string]string{
			"2026-01-01": "h1",
			"2026-01-02": "h2",
			"2026-01-03": "h3",
		},
		BranchToParent: map[string]string{
			"feat1": "main",
			"feat2": "feat1",
		},
		BaseHash: "h1",
	}

	treeCtx := NewTreeContext(seepage)
	if treeCtx == nil {
		t.Fatal("Expected non-nil tree context")
	}

	activeLanes := map[int]string{0: "h1", 1: "h2"}

	// Render lanes for each node
	lanes := treeCtx.RenderLanes(1, activeLanes)
	if lanes == "" {
		t.Errorf("Expected non-empty lanes")
	}

	meta := treeCtx.RenderMetaLanes(1, activeLanes)
	if meta == "" {
		t.Errorf("Expected non-empty meta lanes")
	}

	conn := treeCtx.RenderConnectors("h1", 0, activeLanes)
	if conn == "" {
		t.Errorf("Expected non-empty connectors")
	}
}

func TestTreeContext_AnchorConvergence(t *testing.T) {
	tc := &TreeContext{MaxColumn: 2}
	activeLanes := map[int]string{0: "h1", 2: "h3"}

	// 1. Anchor with multiple lanes
	// Col 0 is the anchor's primary lane. Col 2 is an active branch lane.
	anchor := tc.RenderAnchorLanes(0, activeLanes)
	// Expected: ●─ (col0) + ── (col1 horizontal) + ┴─ (col2 terminating)
	if !strings.Contains(anchor, "●─") || !strings.Contains(anchor, "┴─") {
		t.Errorf("Expected anchor to contain convergence symbols, got %q", anchor)
	}

	// 2. Full-width lanes verification
	lanes := tc.RenderLanes(0, activeLanes)
	// Even though we render for col 0, we expect col 2 (MaxColumn) to be processed
	if len(strings.Split(lanes, " ")) < 3 {
		t.Errorf("Expected lanes to iterate up to MaxColumn, got %q", lanes)
	}
}

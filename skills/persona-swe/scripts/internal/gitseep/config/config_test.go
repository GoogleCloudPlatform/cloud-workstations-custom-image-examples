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

package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/config"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name        string
		yaml        string
		wantBaseRef string
		wantRules   int
		checkRules  func(t *testing.T, cfg *config.GitSeepConfig)
	}{
		{
			name: "Standard Multi-Path Rule",
			yaml: `
"2026-04-15":
  - examples/preflight/
  - examples/images/gnome/
`,
			wantRules: 1,
			checkRules: func(t *testing.T, cfg *config.GitSeepConfig) {
				if len(cfg.Rules[0].Paths) != 2 {
					t.Errorf("expected 2 paths, got %d", len(cfg.Rules[0].Paths))
				}
			},
		},
		{
			name: "Advanced Stacked PR Rule",
			yaml: `
config:
  base_ref: origin/main

"2026-04-16 10:00:00 +0000":
  branch: feat/gnome
  paths:
    - examples/images/gnome/

"2026-04-17 12:00:00 +0000":
  branch: feat/asfp
  parent: feat/gnome
  paths:
    - examples/images/android-studio-for-platform/
`,
			wantBaseRef: "origin/main",
			wantRules:   2,
			checkRules: func(t *testing.T, cfg *config.GitSeepConfig) {
				for _, r := range cfg.Rules {
					if r.Date == "2026-04-17 12:00:00 +0000" {
						if r.Branch != "feat/asfp" || r.Parent != "feat/gnome" {
							t.Errorf("incorrect branch/parent for asfp rule")
						}
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			rulesPath := filepath.Join(tmpDir, ".gitseep.yaml")
			if err := os.WriteFile(rulesPath, []byte(tt.yaml), 0644); err != nil {
				t.Fatalf("failed to write mock config: %v", err)
			}

			cfg, err := config.Load(rulesPath)
			if err != nil {
				t.Fatalf("failed to load config: %v", err)
			}

			if cfg.Global.BaseRef != tt.wantBaseRef {
				t.Errorf("expected base_ref %q, got %q", tt.wantBaseRef, cfg.Global.BaseRef)
			}

			if len(cfg.Rules) != tt.wantRules {
				t.Errorf("expected %d rules, got %d", tt.wantRules, len(cfg.Rules))
			}

			if tt.checkRules != nil {
				tt.checkRules(t, cfg)
			}
		})
	}
}

func TestConfig_Coverage(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("Load missing file", func(t *testing.T) {
		_, err := config.Load(filepath.Join(tmpDir, "nonexistent.yaml"))
		if err == nil {
			t.Error("Load should fail for nonexistent file")
		}
	})

	t.Run("Load invalid YAML", func(t *testing.T) {
		path := filepath.Join(tmpDir, "invalid.yaml")
		_ = os.WriteFile(path, []byte("rules: ["), 0644)
		_, err := config.Load(path)
		if err == nil {
			t.Error("Load should fail for invalid YAML")
		}
	})

	t.Run("Load with map format", func(t *testing.T) {
		path := filepath.Join(tmpDir, "map.yaml")
		content := `
config:
  base_ref: main
2026-01-01:
  branch: feat1
  parent: main
  paths:
    - src/
`
		_ = os.WriteFile(path, []byte(content), 0644)
		cfg, err := config.Load(path)
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}
		if cfg.Global.BaseRef != "main" {
			t.Errorf("Expected base_ref main, got %q", cfg.Global.BaseRef)
		}
		if len(cfg.Rules) != 1 {
			t.Errorf("Expected 1 rule, got %d", len(cfg.Rules))
		}
	})

	t.Run("Load with invalid rule format", func(t *testing.T) {
		path := filepath.Join(tmpDir, "invalid_rule.yaml")
		_ = os.WriteFile(path, []byte("2026-01-01: 123"), 0644)
		_, err := config.Load(path)
		if err == nil {
			t.Error("Load should fail for invalid rule format (scalar)")
		}
	})
}

func TestLoadFromGitRoot(t *testing.T) {
	tmpDir := t.TempDir()
	gitRootDir := filepath.Join(tmpDir, "repo")
	subDir := filepath.Join(gitRootDir, "subdir/deep")

	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("failed to create subdirs: %v", err)
	}

	// Create .git marker
	if err := os.Mkdir(filepath.Join(gitRootDir, ".git"), 0755); err != nil {
		t.Fatalf("failed to create .git marker: %v", err)
	}

	// Create .gitseep.yaml at git root
	content := "2026-01-01: [root/path]"
	if err := os.WriteFile(filepath.Join(gitRootDir, ".gitseep.yaml"), []byte(content), 0644); err != nil {
		t.Fatalf("failed to write config at root: %v", err)
	}

	// Save CWD and change to subDir
	oldCwd, _ := os.Getwd()
	if err := os.Chdir(subDir); err != nil {
		t.Fatalf("failed to change dir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(oldCwd)
	})

	// Load without custom path - should discover at git root
	cfg, err := config.Load("")
	if err != nil {
		t.Fatalf("Load should have discovered config at git root: %v", err)
	}

	if len(cfg.Rules) != 1 || cfg.Rules[0].Paths[0] != "root/path" {
		t.Errorf("loaded incorrect config from git root")
	}
}

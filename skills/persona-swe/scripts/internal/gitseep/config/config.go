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

// Package config handles the parsing and validation of GitSeep YAML rules.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/i18n"
	"gopkg.in/yaml.v3"
)

// Standardized exit codes for CLI integration.
const (
	// keep-sorted start
	ExitCodeConfig      = 78
	ExitCodeDataErr     = 65
	ExitCodeError       = 1
	ExitCodeNoInput     = 66
	ExitCodeOK          = 0
	ExitCodePolicy      = 3
	ExitCodeSigInt      = 130
	ExitCodeUnavailable = 69
	ExitCodeUsage       = 2
	// keep-sorted end
)

// RulesFileName defines the standard configuration file name.
const RulesFileName = ".gitseep.yaml"

// GlobalConfig contains repository-wide settings for GitSeep.
type GlobalConfig struct {
	BaseRef string `yaml:"base_ref"`
}

// Rule represents a single declarative bedrock definition in the config file.
type Rule struct {
	// keep-sorted start
	Branch string
	Date   string
	Parent string
	Paths  []string
	// keep-sorted end
}

// GitSeepConfig holds the complete parsed configuration.
type GitSeepConfig struct {
	Global GlobalConfig
	Rules  []Rule
}

// Load parses the GitSeep rules from the specified path or the default local file.
func Load(customPath string) (*GitSeepConfig, error) {
	rulesPath := customPath
	if rulesPath == "" {
		cwd, _ := os.Getwd()
		localRules := filepath.Join(cwd, RulesFileName)
		if _, err := os.Stat(localRules); err == nil {
			rulesPath = localRules
		} else {
			// Try Git root discovery
			gitRoot := findGitRoot(cwd)
			if gitRoot != "" && gitRoot != cwd {
				rootRules := filepath.Join(gitRoot, RulesFileName)
				if _, err := os.Stat(rootRules); err == nil {
					rulesPath = rootRules
				}
			}
		}
	}

	if rulesPath == "" {
		return nil, fmt.Errorf("%s", i18n.TF("error_rules_not_found", map[string]interface{}{"File": RulesFileName}))
	}

	data, err := os.ReadFile(rulesPath)
	if err != nil {
		return nil, fmt.Errorf("%s", i18n.TF("error_read_rules", map[string]interface{}{"Error": err}))
	}

	var raw map[string]interface{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("%s", i18n.TF("error_parse_yaml", map[string]interface{}{"Error": err}))
	}

	cfg := &GitSeepConfig{}

	if configNode, ok := raw["config"]; ok {
		if configMap, ok := configNode.(map[string]interface{}); ok {
			if baseRef, ok := configMap["base_ref"].(string); ok {
				cfg.Global.BaseRef = baseRef
			}
		}
		delete(raw, "config")
	}

	for key, val := range raw {
		rule := Rule{Date: key}
		switch v := val.(type) {
		case []interface{}:
			for _, p := range v {
				if pathStr, ok := p.(string); ok {
					rule.Paths = append(rule.Paths, pathStr)
				}
			}
		case map[string]interface{}:
			if branch, ok := v["branch"].(string); ok {
				rule.Branch = branch
			}
			if parent, ok := v["parent"].(string); ok {
				rule.Parent = parent
			}
			if paths, ok := v["paths"].([]interface{}); ok {
				for _, p := range paths {
					if pathStr, ok := p.(string); ok {
						rule.Paths = append(rule.Paths, pathStr)
					}
				}
			}
		default:
			return nil, fmt.Errorf("%s", i18n.TF("error_invalid_rule_format", map[string]interface{}{"Key": key}))
		}
		cfg.Rules = append(cfg.Rules, rule)
	}

	return cfg, nil
}

// findGitRoot traverses up the directory tree to find the Git repository root.
func findGitRoot(startDir string) string {
	curr := startDir
	for {
		if _, err := os.Stat(filepath.Join(curr, ".git")); err == nil {
			return curr
		}
		parent := filepath.Dir(curr)
		if parent == curr {
			break
		}
		curr = parent
	}
	return ""
}

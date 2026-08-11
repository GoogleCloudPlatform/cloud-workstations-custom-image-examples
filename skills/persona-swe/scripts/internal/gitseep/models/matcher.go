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

import "strings"

// trieNode represents a single level in the path-prefix tree.
type trieNode struct {
	children map[string]*trieNode
	bedrock  string
	rulePath string
}

func (n *trieNode) insert(components []string, bedrock, rulePath string) {
	curr := n
	for _, comp := range components {
		if curr.children == nil {
			curr.children = make(map[string]*trieNode)
		}
		child, ok := curr.children[comp]
		if !ok {
			child = &trieNode{}
			curr.children[comp] = child
		}
		curr = child
	}
	curr.bedrock = bedrock
	curr.rulePath = rulePath
}

func (n *trieNode) resolve(components []string) (string, string) {
	var bestBedrock, bestRulePath string
	curr := n
	for _, comp := range components {
		if curr.bedrock != "" {
			bestBedrock = curr.bedrock
			bestRulePath = curr.rulePath
		}
		if curr.children == nil {
			return bestBedrock, bestRulePath
		}
		child, ok := curr.children[comp]
		if !ok {
			return bestBedrock, bestRulePath
		}
		curr = child
	}
	if curr.bedrock != "" {
		bestBedrock = curr.bedrock
		bestRulePath = curr.rulePath
	}
	return bestBedrock, bestRulePath
}

// GeologicalMatcher handles the mapping of file paths to architectural bedrock commits using an optimized Path-Trie.
type GeologicalMatcher struct {
	PathToBedrock map[string]string
	OwnedPaths    []string
	root          *trieNode
}

// NewGeologicalMatcher initializes a matcher with the provided path-to-bedrock mapping.
func NewGeologicalMatcher(pathToBedrock map[string]string) *GeologicalMatcher {
	owned := make([]string, 0, len(pathToBedrock))
	root := &trieNode{}
	for p, h := range pathToBedrock {
		owned = append(owned, p)
		components := strings.Split(strings.Trim(p, "/"), "/")
		if p == "." || p == "" {
			components = []string{}
		}
		root.insert(components, h, p)
	}
	return &GeologicalMatcher{
		PathToBedrock: pathToBedrock,
		OwnedPaths:    owned,
		root:          root,
	}
}

// ResolveTarget returns the bedrock hash and the matching rule path for a given file.
// It follows the "longest match wins" principle, optimized via Trie lookup.
func (m *GeologicalMatcher) ResolveTarget(filePath string) (bedrockHash string, rulePath string) {
	components := strings.Split(strings.Trim(filePath, "/"), "/")
	return m.root.resolve(components)
}

// BelongsToPath checks if a file path is governed by a rule path (prefix matching).
func BelongsToPath(filePath, rulePath string) bool {
	if filePath == rulePath {
		return true
	}
	// Handle directory matching: /src matches /src/main.go
	if len(rulePath) > 0 && rulePath[len(rulePath)-1] != '/' {
		return filePath == rulePath || (len(filePath) > len(rulePath) && filePath[len(rulePath)] == '/' && filePath[:len(rulePath)] == rulePath)
	}
	return len(filePath) >= len(rulePath) && filePath[:len(rulePath)] == rulePath
}

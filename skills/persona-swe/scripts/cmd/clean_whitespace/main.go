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

// Package main provides the entry point for the clean_whitespace utility.
package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/i18n"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/locales"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/whitespace"
)

const (
	concurrencyLimit = 20
)

var (
	// supportedExtensions defines the set of file types that are safe to clean.
	// We use a map for O(1) lookups during the file walk.
	supportedExtensions = map[string]bool{
		// keep-sorted start
		".bash": true,
		".bats": true,
		".conf": true,
		".go":   true,
		".js":   true,
		".json": true,
		".list": true,
		".md":   true,
		".sh":   true,
		".ts":   true,
		".yaml": true,
		".yml":  true,
		// keep-sorted end
	}

	// supportedFilenames defines exact filenames that are safe to clean.
	supportedFilenames = map[string]bool{
		// keep-sorted start
		"Dockerfile": true,
		"GEMINI.md":  true,
		// keep-sorted end
	}
)

func isSupported(path string) bool {
	name := filepath.Base(path)
	if supportedFilenames[name] {
		return true
	}
	ext := strings.ToLower(filepath.Ext(path))
	return supportedExtensions[ext]
}

func main() {
	if err := i18n.Init("", "en", locales.Content); err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize i18n: %v\n", err)
		os.Exit(1)
	}

	flag.Parse()
	args := flag.Args()

	if len(args) == 0 {
		fmt.Println(i18n.T("usage_clean_whitespace"))
		os.Exit(0)
	}

	var paths []string
	for _, arg := range args {
		info, err := os.Stat(arg)
		if err != nil {
			fmt.Println(i18n.TF("error_failed_to_stat", map[string]interface{}{"Path": arg, "Error": err}))
			continue
		}

		if !info.IsDir() {
			if isSupported(arg) {
				paths = append(paths, arg)
			}
			continue
		}

		err = filepath.WalkDir(arg, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() {
				if isSupported(path) {
					paths = append(paths, path)
				}
			}
			if d.IsDir() && (d.Name() == ".git" || d.Name() == "node_modules") {
				return filepath.SkipDir
			}
			return nil
		})
		if err != nil {
			fmt.Println(i18n.TF("error_failed_to_walk", map[string]interface{}{"Path": arg, "Error": err}))
		}
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	modifiedCount := 0
	sem := make(chan struct{}, concurrencyLimit)

	for _, p := range paths {
		wg.Add(1)
		go func(path string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			modified, err := whitespace.CleanFile(path)
			if err != nil {
				fmt.Println(i18n.TF("error_failed_to_clean", map[string]interface{}{"Path": path, "Error": err}))
				return
			}
			if modified {
				mu.Lock()
				fmt.Println(i18n.TF("clean_whitespace_fixed", map[string]interface{}{"Path": path}))
				modifiedCount++
				mu.Unlock()
			}
		}(p)
	}

	wg.Wait()

	if modifiedCount > 0 {
		fmt.Println(i18n.TF("clean_whitespace_summary", map[string]interface{}{"Count": modifiedCount}))
		os.Exit(1)
	}
}

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

// Package main provides a CLI utility to verify i18n locale JSON files for sorting and consistency.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/locales"
)

func main() {
	flag.Parse()
	args := flag.Args()

	if len(args) == 0 {
		fmt.Println("usage: verify_locales <path1> <path2> ...")
		os.Exit(1)
	}

	dirs := make(map[string]bool)
	for _, arg := range args {
		info, err := os.Stat(arg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error statting path %s: %v\n", arg, err)
			os.Exit(1)
		}
		if info.IsDir() {
			dirs[arg] = true
		} else {
			dirs[filepath.Dir(arg)] = true
		}
	}

	failed := false
	for dir := range dirs {
		if err := locales.VerifyDir(dir); err != nil {
			fmt.Fprintf(os.Stderr, "error in directory %s: %v\n", dir, err)
			failed = true
		}
	}

	if failed {
		os.Exit(1)
	}
	fmt.Println("All specified locale files are sorted and consistent.")
}

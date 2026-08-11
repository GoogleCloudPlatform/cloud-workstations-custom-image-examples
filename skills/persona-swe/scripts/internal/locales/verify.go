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

// Package locales provides utilities for verifying i18n locale JSON files.
package locales

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// LocaleItem represents a single translation entry in the Array format.
type LocaleItem struct {
	ID          string `json:"id"`
	Translation string `json:"translation"`
}

// GetKeys extracts the IDs from a locale JSON file, supporting both Array and Map formats.
// It uses a Decoder to strictly verify the order of keys as they appear in the file.
func GetKeys(data []byte) ([]string, bool, error) {
	trimmed := strings.TrimSpace(string(data))
	isMap := strings.HasPrefix(trimmed, "{")

	dec := json.NewDecoder(strings.NewReader(trimmed))
	var keys []string

	t, err := dec.Token()
	if err != nil {
		return nil, false, err
	}

	if isMap {
		if t != json.Delim('{') {
			return nil, false, fmt.Errorf("expected { at start of map")
		}
		for dec.More() {
			k, err := dec.Token()
			if err != nil {
				return nil, false, err
			}
			keys = append(keys, k.(string))

			// Skip value
			var v interface{}
			if err := dec.Decode(&v); err != nil {
				return nil, false, err
			}
		}
	} else {
		if t != json.Delim('[') {
			return nil, false, fmt.Errorf("expected [ at start of array")
		}
		for dec.More() {
			var item map[string]interface{}
			if err := dec.Decode(&item); err != nil {
				return nil, false, err
			}
			id, ok := item["id"].(string)
			if !ok {
				return nil, false, fmt.Errorf("missing or invalid id in array item")
			}
			keys = append(keys, id)
		}
	}

	return keys, isMap, nil
}

// VerifyDir ensures that all locale JSON files in a directory are sorted and consistent with en.json.
func VerifyDir(dir string) error {
	primaryPath := filepath.Join(dir, "en.json")
	primaryData, err := os.ReadFile(primaryPath)
	if err != nil {
		return fmt.Errorf("failed to read primary file en.json: %w", err)
	}

	// 0. Verify newline at end of file
	if len(primaryData) == 0 || primaryData[len(primaryData)-1] != '\n' {
		return fmt.Errorf("en.json does not end with a newline")
	}

	primaryKeys, _, err := GetKeys(primaryData)
	if err != nil {
		return fmt.Errorf("failed to parse en.json: %w", err)
	}

	// Check if primary itself is sorted
	if !sort.StringsAreSorted(primaryKeys) {
		return fmt.Errorf("en.json keys are not alphabetically sorted")
	}

	primaryKeySet := make(map[string]bool)
	for _, k := range primaryKeys {
		primaryKeySet[k] = true
	}

	files, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	for _, f := range files {
		if f.IsDir() || !strings.HasSuffix(f.Name(), ".json") || f.Name() == "en.json" {
			continue
		}

		path := filepath.Join(dir, f.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", f.Name(), err)
		}

		// 0. Verify newline at end of file
		if len(data) == 0 || data[len(data)-1] != '\n' {
			return fmt.Errorf("%s does not end with a newline", f.Name())
		}

		keys, _, err := GetKeys(data)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", f.Name(), err)
		}

		// 1. Check Sorting
		if !sort.StringsAreSorted(keys) {
			return fmt.Errorf("%s keys are not alphabetically sorted", f.Name())
		}

		// 2. Check Consistency
		keySet := make(map[string]bool)
		for _, k := range keys {
			keySet[k] = true
		}

		// Check for missing keys
		for _, k := range primaryKeys {
			if !keySet[k] {
				return fmt.Errorf("%s is missing key: %s", f.Name(), k)
			}
		}

		// Check for extra keys
		for _, k := range keys {
			if !primaryKeySet[k] {
				return fmt.Errorf("%s has extra key not in en.json: %s", f.Name(), k)
			}
		}
	}

	return nil
}

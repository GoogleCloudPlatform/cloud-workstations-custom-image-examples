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

// Package whitespace provides utilities for cleaning source code whitespace.
package whitespace

import (
	"bytes"
	"fmt"
	"os"
)

// CleanFile reads a file, applies whitespace cleaning, and writes it back if changed.
func CleanFile(path string) (bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return false, fmt.Errorf("failed to read file: %w", err)
	}

	if len(data) == 0 {
		return false, nil
	}

	// Skip binary files (naive check for null byte).
	if bytes.IndexByte(data, 0) != -1 {
		return false, nil
	}

	newData, modified := CleanData(data)
	if !modified {
		return false, nil
	}

	if err := os.WriteFile(path, newData, 0644); err != nil {
		return false, fmt.Errorf("failed to write file: %w", err)
	}

	return true, nil
}

// CleanData removes trailing whitespace from lines and ensures exactly one newline at EOF.
// It returns the cleaned data and a boolean indicating if any changes were made.
func CleanData(data []byte) ([]byte, bool) {
	// Split by newline.
	lines := bytes.Split(data, []byte("\n"))
	var cleanedLines [][]byte

	// 1. Trim trailing spaces/tabs from each line.
	for _, line := range lines {
		cleanedLines = append(cleanedLines, bytes.TrimRight(line, " \t\r"))
	}

	// 2. Reconstruct with \n and then TrimRight to handle any trailing blank lines.
	newData := bytes.Join(cleanedLines, []byte("\n"))
	newData = bytes.TrimRight(newData, " \n\r\t")

	// 3. Ensure exactly one newline at EOF if not empty.
	if len(newData) > 0 {
		newData = append(newData, '\n')
	}

	return newData, !bytes.Equal(data, newData)
}

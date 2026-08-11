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

package locales

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetKeys(t *testing.T) {
	t.Run("Array format sorted", func(t *testing.T) {
		data := []byte(`[{"id": "a", "translation": "1"}, {"id": "b", "translation": "2"}]`)
		keys, isMap, err := GetKeys(data)
		if err != nil {
			t.Fatalf("GetKeys failed: %v", err)
		}
		if isMap {
			t.Error("Expected array format, got map")
		}
		if len(keys) != 2 || keys[0] != "a" || keys[1] != "b" {
			t.Errorf("Unexpected keys: %v", keys)
		}
	})

	t.Run("Map format sorted", func(t *testing.T) {
		data := []byte(`{"a": "1", "b": "2"}`)
		keys, isMap, err := GetKeys(data)
		if err != nil {
			t.Fatalf("GetKeys failed: %v", err)
		}
		if !isMap {
			t.Error("Expected map format, got array")
		}
		if len(keys) != 2 || keys[0] != "a" || keys[1] != "b" {
			t.Errorf("Unexpected keys: %v", keys)
		}
	})

	t.Run("Invalid format", func(t *testing.T) {
		data := []byte(`invalid`)
		_, _, err := GetKeys(data)
		if err == nil {
			t.Error("Expected error for invalid format")
		}
	})

	t.Run("Missing id in array", func(t *testing.T) {
		data := []byte(`[{"translation": "1"}]`)
		_, _, err := GetKeys(data)
		if err == nil || err.Error() != "missing or invalid id in array item" {
			t.Errorf("Expected id error, got %v", err)
		}
	})
}

func TestVerifyDir(t *testing.T) {
	tmp := t.TempDir()

	t.Run("Consistent and sorted", func(t *testing.T) {
		_ = os.WriteFile(filepath.Join(tmp, "en.json"), []byte(`{"a": "1", "b": "2"}`+"\n"), 0644)
		_ = os.WriteFile(filepath.Join(tmp, "es.json"), []byte(`{"a": "1", "b": "2"}`+"\n"), 0644)

		if err := VerifyDir(tmp); err != nil {
			t.Errorf("Expected success, got %v", err)
		}
	})

	t.Run("Missing primary", func(t *testing.T) {
		emptyDir := t.TempDir()
		if err := VerifyDir(emptyDir); err == nil || !strings.Contains(err.Error(), "failed to read primary file en.json") {
			t.Errorf("Expected missing primary error, got %v", err)
		}
	})

	t.Run("Malformed primary", func(t *testing.T) {
		dir := t.TempDir()
		_ = os.WriteFile(filepath.Join(dir, "en.json"), []byte(`{invalid`+"\n"), 0644)
		if err := VerifyDir(dir); err == nil || !strings.Contains(err.Error(), "failed to parse en.json") {
			t.Errorf("Expected parse error, got %v", err)
		}
	})

	t.Run("Unsorted primary", func(t *testing.T) {
		dir := filepath.Join(tmp, "unsorted_primary")
		_ = os.Mkdir(dir, 0755)
		_ = os.WriteFile(filepath.Join(dir, "en.json"), []byte(`{"b": "2", "a": "1"}`+"\n"), 0644)

		if err := VerifyDir(dir); err == nil || err.Error() != "en.json keys are not alphabetically sorted" {
			t.Errorf("Expected unsorted primary error, got %v", err)
		}
	})

	t.Run("Unsorted locale", func(t *testing.T) {
		dir := filepath.Join(tmp, "unsorted_locale")
		_ = os.Mkdir(dir, 0755)
		_ = os.WriteFile(filepath.Join(dir, "en.json"), []byte(`{"a": "1", "b": "2"}`+"\n"), 0644)
		_ = os.WriteFile(filepath.Join(dir, "es.json"), []byte(`{"b": "2", "a": "1"}`+"\n"), 0644)

		if err := VerifyDir(dir); err == nil || err.Error() != "es.json keys are not alphabetically sorted" {
			t.Errorf("Expected unsorted locale error, got %v", err)
		}
	})

	t.Run("Missing key", func(t *testing.T) {
		dir := filepath.Join(tmp, "missing_key")
		_ = os.Mkdir(dir, 0755)
		_ = os.WriteFile(filepath.Join(dir, "en.json"), []byte(`{"a": "1", "b": "2"}`+"\n"), 0644)
		_ = os.WriteFile(filepath.Join(dir, "es.json"), []byte(`{"a": "1"}`+"\n"), 0644)

		if err := VerifyDir(dir); err == nil || err.Error() != "es.json is missing key: b" {
			t.Errorf("Expected missing key error, got %v", err)
		}
	})

	t.Run("Extra key", func(t *testing.T) {
		dir := filepath.Join(tmp, "extra_key")
		_ = os.Mkdir(dir, 0755)
		_ = os.WriteFile(filepath.Join(dir, "en.json"), []byte(`{"a": "1"}`+"\n"), 0644)
		_ = os.WriteFile(filepath.Join(dir, "es.json"), []byte(`{"a": "1", "b": "2"}`+"\n"), 0644)

		if err := VerifyDir(dir); err == nil || err.Error() != "es.json has extra key not in en.json: b" {
			t.Errorf("Expected extra key error, got %v", err)
		}
	})

	t.Run("Missing newline", func(t *testing.T) {
		dir := t.TempDir()
		_ = os.WriteFile(filepath.Join(dir, "en.json"), []byte(`{"a": "1"}`), 0644) // No newline
		if err := VerifyDir(dir); err == nil || err.Error() != "en.json does not end with a newline" {
			t.Errorf("Expected newline error, got %v", err)
		}
	})
}

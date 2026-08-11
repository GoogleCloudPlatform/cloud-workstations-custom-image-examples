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

// Package gitutil provides low-level generic Git tree operations.
package gitutil

import (
	"fmt"
	"sort"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/utils/merkletrie"
)

// GetChangedFiles identifies exactly which files were modified between two commits.
func GetChangedFiles(repo *git.Repository, fromHash, toHash string) ([]string, error) {
	from, err := repo.CommitObject(plumbing.NewHash(fromHash))
	if err != nil {
		return nil, fmt.Errorf("failed to get from commit %s: %w", fromHash, err)
	}
	to, err := repo.CommitObject(plumbing.NewHash(toHash))
	if err != nil {
		return nil, fmt.Errorf("failed to get to commit %s: %w", toHash, err)
	}

	fromTree, err := from.Tree()
	if err != nil {
		return nil, err
	}
	toTree, err := to.Tree()
	if err != nil {
		return nil, err
	}

	changes, err := fromTree.Diff(toTree)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, change := range changes {
		action, _ := change.Action()
		if action == merkletrie.Delete {
			files = append(files, change.From.Name)
		} else {
			files = append(files, change.To.Name)
		}
	}
	return files, nil
}

// GetAllEntries recursively retrieves all files from a Git tree.
func GetAllEntries(repo *git.Repository, tree *object.Tree) (map[string]object.TreeEntry, error) {
	entries := make(map[string]object.TreeEntry)
	err := GetAllEntriesRecursive(repo, tree, "", entries)
	return entries, err
}

// GetAllEntriesRecursive recursively retrieves all files from a Git tree.
func GetAllEntriesRecursive(repo *git.Repository, tree *object.Tree, prefix string, entries map[string]object.TreeEntry) error {
	for _, e := range tree.Entries {
		fullPath := e.Name
		if prefix != "" {
			fullPath = prefix + "/" + e.Name
		}
		if e.Mode == filemode.Dir {
			subTree, err := repo.TreeObject(e.Hash)
			if err != nil {
				return err
			}
			err = GetAllEntriesRecursive(repo, subTree, fullPath, entries)
			if err != nil {
				return err
			}
		} else {
			e.Name = fullPath
			entries[fullPath] = e
		}
	}
	return nil
}

// BuildTree constructs a Git tree object from a map of file paths to tree entries.
func BuildTree(repo *git.Repository, files map[string]object.TreeEntry) (plumbing.Hash, error) {
	if len(files) == 0 {
		t := &object.Tree{}
		obj := repo.Storer.NewEncodedObject()
		_ = t.Encode(obj)
		return repo.Storer.SetEncodedObject(obj)
	}

	treeEntries := make(map[string]map[string]object.TreeEntry)
	treeEntries[""] = make(map[string]object.TreeEntry)

	for path, entry := range files {
		parts := strings.Split(path, "/")
		dir := ""
		for i := 0; i < len(parts)-1; i++ {
			parentDir := dir
			if dir != "" {
				dir += "/"
			}
			dir += parts[i]

			if _, ok := treeEntries[dir]; !ok {
				treeEntries[dir] = make(map[string]object.TreeEntry)
				treeEntries[parentDir][parts[i]] = object.TreeEntry{
					Name: parts[i],
					Mode: filemode.Dir,
				}
			}
		}

		name := parts[len(parts)-1]
		entry.Name = name
		treeEntries[dir][name] = entry
	}

	var dirs []string
	for d := range treeEntries {
		dirs = append(dirs, d)
	}
	sort.Slice(dirs, func(i, j int) bool {
		c1, c2 := strings.Count(dirs[i], "/"), strings.Count(dirs[j], "/")
		if c1 != c2 {
			return c1 > c2 // Deepest first
		}
		return dirs[i] > dirs[j]
	})

	for _, dir := range dirs {
		entriesMap := treeEntries[dir]
		var entries []object.TreeEntry
		for _, e := range entriesMap {
			entries = append(entries, e)
		}
		sort.Sort(sortableEntries(entries))

		t := &object.Tree{Entries: entries}
		obj := repo.Storer.NewEncodedObject()
		if err := t.Encode(obj); err != nil {
			return plumbing.ZeroHash, fmt.Errorf("failed to encode tree object: %w", err)
		}
		hash, err := repo.Storer.SetEncodedObject(obj)
		if err != nil {
			return plumbing.ZeroHash, fmt.Errorf("failed to set encoded object for tree: %w", err)
		}

		if dir != "" {
			lastSlash := strings.LastIndex(dir, "/")
			parentDir := ""
			name := dir
			if lastSlash != -1 {
				parentDir = dir[:lastSlash]
				name = dir[lastSlash+1:]
			}

			parentEntry := treeEntries[parentDir][name]
			parentEntry.Hash = hash
			treeEntries[parentDir][name] = parentEntry
		} else {
			return hash, nil
		}
	}

	return plumbing.ZeroHash, fmt.Errorf("root not found")
}

type sortableEntries []object.TreeEntry

func (se sortableEntries) sortName(te object.TreeEntry) string {
	if te.Mode == filemode.Dir {
		return te.Name + "/"
	}
	return te.Name
}
func (se sortableEntries) Len() int           { return len(se) }
func (se sortableEntries) Less(i, j int) bool { return se.sortName(se[i]) < se.sortName(se[j]) }
func (se sortableEntries) Swap(i, j int)      { se[i], se[j] = se[j], se[i] }

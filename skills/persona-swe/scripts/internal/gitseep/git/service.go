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

// Package git provides helpers for interacting with the Git repository.
package git

import (
	"fmt"
	"os/exec"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/gitutil"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/models"
)

// GitService defines the abstract operations for interacting with a Git repository.
// It bridges the gap between fast in-memory 'go-git' operations and robust
// 'system-git' worktree mutations.
type GitService interface {
	// GetRepo returns the underlying go-git repository instance for object resolution.
	GetRepo() *git.Repository

	// GetStatus returns the current state of the worktree.
	GetStatus() (hasChanges, hasUntracked bool, err error)

	// IsClean returns true if there are no uncommitted changes (tracked or untracked).
	IsClean() (bool, error)

	// Amend folds already staged changes into the current HEAD commit.
	Amend() error

	// ResetMixed performs a 'git reset --mixed', updating the branch and index
	// while preserving all worktree modifications.
	ResetMixed(target string) error

	// ResolveRevision resolves a symbolic reference (like 'HEAD', 'main', or a partial hash)
	// into a full plumbing.Hash.
	ResolveRevision(rev string) (plumbing.Hash, error)

	// UpdateBranch force-updates a branch reference to point to a new commit.
	UpdateBranch(name string, hash plumbing.Hash) error

	// GetChangedFiles identifies paths modified between two revisions.
	GetChangedFiles(from, to string) ([]string, error)

	// GetCurrentBranch returns the name of the currently checked-out branch.
	GetCurrentBranch() (string, error)

	// RepoRoot returns the absolute path to the repository root.
	RepoRoot() string

	// SafeFinalize performs the multi-step geological finalization, updating
	// all references and syncing the worktree safely.
	SafeFinalize(seepage *models.SeepageContext, linearHead string, branchHeads map[string]string) (string, error)

	// VirtualAmend simulates a 'git commit --amend' by creating a dangling commit
	// object in the Git database without moving any branch references.
	VirtualAmend() (plumbing.Hash, error)
}

// hybridGitService is the primary implementation of GitService, using a combination
// of go-git for read-heavy object resolution and system git for reliable worktree mutations.
type hybridGitService struct {
	repo         *git.Repository
	root         string
	experimental bool
}

// NewGitService initializes a new hybrid Git service for the specified directory.
func NewGitService(path string, experimental bool) (GitService, error) {
	repo, err := git.PlainOpenWithOptions(path, &git.PlainOpenOptions{DetectDotGit: true})
	if err != nil {
		return nil, fmt.Errorf("failed to open git repository: %w", err)
	}

	wt, err := repo.Worktree()
	if err != nil {
		return nil, err
	}

	return &hybridGitService{
		repo:         repo,
		root:         wt.Filesystem.Root(),
		experimental: experimental,
	}, nil
}

func (s *hybridGitService) GetRepo() *git.Repository {
	return s.repo
}

func (s *hybridGitService) GetStatus() (bool, bool, error) {
	return GetStatusSummary(s.repo, StatusOptions{UseGoGit: s.experimental})
}

func (s *hybridGitService) IsClean() (bool, error) {
	hasChanges, hasUntracked, err := s.GetStatus()
	return !hasChanges && !hasUntracked, err
}

func (s *hybridGitService) Amend() error {
	return Amend(s.repo)
}

func (s *hybridGitService) VirtualAmend() (plumbing.Hash, error) {
	return VirtualAmend(s.repo)
}

func (s *hybridGitService) ResetMixed(target string) error {
	if s.experimental {
		wt, err := s.repo.Worktree()
		if err != nil {
			return err
		}
		return wt.Reset(&git.ResetOptions{
			Commit: plumbing.NewHash(target),
			Mode:   git.MixedReset,
		})
	}

	// System git is required for --mixed reset as go-git's Reset implementation
	// is heavily tied to worktree checkout or hard reset.
	cmd := exec.Command("git", "reset", "--mixed", target)
	cmd.Dir = s.root
	return cmd.Run()
}

func (s *hybridGitService) ResolveRevision(rev string) (plumbing.Hash, error) {
	h, err := s.repo.ResolveRevision(plumbing.Revision(rev))
	if err != nil {
		return plumbing.ZeroHash, err
	}
	return *h, nil
}

func (s *hybridGitService) UpdateBranch(name string, hash plumbing.Hash) error {
	if s.experimental {
		refName := plumbing.NewBranchReferenceName(name)
		return s.repo.Storer.SetReference(plumbing.NewHashReference(refName, hash))
	}

	// System git is used for reliable reference updates that are immediately visible
	cmd := exec.Command("git", "update-ref", "refs/heads/"+name, hash.String())
	cmd.Dir = s.root
	err := cmd.Run()
	if err == nil {
		fmt.Printf("Updated branch %s -> %s\n", name, hash.String()[:7])
	}
	return err
}

func (s *hybridGitService) GetChangedFiles(from, to string) ([]string, error) {
	return gitutil.GetChangedFiles(s.repo, from, to)
}

func (s *hybridGitService) GetCurrentBranch() (string, error) {
	head, err := s.repo.Head()
	if err != nil {
		return "", err
	}
	if head.Name().IsBranch() {
		return head.Name().Short(), nil
	}
	return head.Hash().String(), nil
}

func (s *hybridGitService) RepoRoot() string {
	return s.root
}

func (s *hybridGitService) SafeFinalize(seepage *models.SeepageContext, linearHead string, branchHeads map[string]string) (string, error) {
	if seepage.Options.DryRun {
		return "", nil
	}

	// 1. Update Target/Staged Branch
	targetBranch := seepage.TargetBranch
	isStagedOnly := seepage.Options.StageOnly
	if isStagedOnly {
		// Use unique name to avoid clobbering concurrent runs
		targetBranch = fmt.Sprintf("gitseep/staged-%d", time.Now().Unix())
	}

	// DETACHED HEAD PROTECTION: If target branch is a hash, it's not a real branch.
	// We skip the update-ref but will still reset HEAD to the new history.
	isDetached := len(targetBranch) == 40 // simple hash check

	if !isDetached {
		if err := s.UpdateBranch(targetBranch, plumbing.NewHash(linearHead)); err != nil {
			return "", fmt.Errorf("failed to update target branch %s: %w", targetBranch, err)
		}
	} else {
		fmt.Printf("Starting from detached HEAD [%s], skipping branch reference update.\n", targetBranch[:7])
	}

	// 2. Update Feature Branches
	if !isStagedOnly {
		for branchName, hashStr := range branchHeads {
			if err := s.UpdateBranch(branchName, plumbing.NewHash(hashStr)); err != nil {
				return "", fmt.Errorf("failed to update feature branch %s: %w", branchName, err)
			}
		}
	}

	// 3. Sync Worktree
	if !isStagedOnly {
		head, _ := s.repo.Head()
		currentRef := ""
		if head != nil {
			if head.Name().IsBranch() {
				currentRef = head.Name().Short()
			} else {
				currentRef = head.Hash().String()
			}
		}

		if currentRef == seepage.TargetBranch {
			if err := s.ResetMixed(linearHead); err != nil {
				return "", fmt.Errorf("failed to sync worktree: %w", err)
			}
		}
	}

	return targetBranch, nil
}

// NewTestGitService creates a GitService wrapping an existing go-git repository,
// primarily for use in unit tests.
func NewTestGitService(repo *git.Repository) GitService {
	wt, _ := repo.Worktree()
	root := ""
	if wt != nil {
		root = wt.Filesystem.Root()
	}
	return &hybridGitService{
		repo:         repo,
		root:         root,
		experimental: true, // Default to go-git for test stability
	}
}

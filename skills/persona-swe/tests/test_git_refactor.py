#!/usr/bin/env python3

# Copyright 2026 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     https://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

"""Unit tests for git_refactor.py.

This module provides comprehensive test coverage for the interactive
Source Code Versioning History Refactoring tool.
"""

import sys
import os
import subprocess
import unittest
from unittest import mock

# Add the scripts directory to sys.path to allow importing the module under test
sys.path.insert(0, os.path.abspath(os.path.join(os.path.dirname(__file__), '../scripts')))

import git_refactor


class TestGitRefactor(unittest.TestCase):
    """Test suite for git_refactor.py functions."""

    @mock.patch('git_refactor.subprocess.run')
    def test_run_git_success(self, mock_run):
        """Test successful git command execution."""
        mock_run.return_value = subprocess.CompletedProcess(
            args=['git', 'status'], returncode=0, stdout='On branch main', stderr=''
        )
        result = git_refactor.run_git(['status'])
        self.assertEqual(result.stdout, 'On branch main')
        mock_run.assert_called_once()

    @mock.patch('git_refactor.subprocess.run')
    def test_run_git_failure_with_check(self, mock_run):
        """Test git command failure when check is True."""
        mock_run.side_effect = subprocess.CalledProcessError(
            returncode=1, cmd=['git', 'invalid'], stderr='error message'
        )
        with self.assertRaises(git_refactor.GitError):
            git_refactor.run_git(['invalid'], check=True)

    @mock.patch('git_refactor.subprocess.run')
    def test_run_git_failure_without_check(self, mock_run):
        """Test git command failure when check is False."""
        mock_run.side_effect = subprocess.CalledProcessError(
            returncode=1, cmd=['git', 'invalid'], stderr='error message'
        )
        result = git_refactor.run_git(['invalid'], check=False)
        self.assertEqual(result.returncode, 1)
        self.assertEqual(result.stderr, 'error message')

    @mock.patch('git_refactor.run_git')
    def test_get_commit_sequence(self, mock_run_git):
        """Test retrieving commit sequence."""
        mock_run_git.return_value = mock.Mock(stdout='hash3\nhash2\nhash1\n')
        sequence = git_refactor.get_commit_sequence('hash1')
        self.assertEqual(sequence, ['hash1', 'hash2', 'hash3'])
        mock_run_git.assert_called_with(['log', '--format=%H', 'hash1^..HEAD'])

    @mock.patch('git_refactor.run_git')
    def test_resolve_commit_by_date_success(self, mock_run_git):
        """Test resolving commit hash from date."""
        mock_run_git.return_value = mock.Mock(stdout='HASH123\n')
        result = git_refactor.resolve_commit_by_date('2026-04-15')
        self.assertEqual(result, 'HASH123')
        mock_run_git.assert_called_with([
            'log', '--format=%H', '--author-date-order', 
            '--after=2026-04-15', '--before=2026-04-15', '--all'
        ])

    @mock.patch('builtins.input')
    def test_user_confirm_yes(self, mock_input):
        """Test user confirmation with 'y'."""
        mock_input.side_effect = ['y']
        self.assertTrue(git_refactor.user_confirm("Confirm?"))

    @mock.patch('builtins.input')
    def test_user_confirm_no(self, mock_input):
        """Test user confirmation with 'n'."""
        mock_input.side_effect = ['n']
        self.assertFalse(git_refactor.user_confirm("Confirm?"))

    @mock.patch('git_refactor.run_git')
    @mock.patch('git_refactor.resolve_commit_by_date')
    @mock.patch('git_refactor.user_confirm')
    def test_refactor_history_aborted(self, mock_confirm, mock_resolve, mock_run_git):
        """Test refactoring when user aborts at pre-flight."""
        mock_run_git.side_effect = [
            mock.Mock(stdout=''),  # status
            mock.Mock(stdout='HEAD_HASH'),  # rev-parse HEAD
            mock.Mock(stdout='main'),  # rev-parse branch
            mock.Mock(stdout='1000'),  # log timestamp
            mock.Mock(stdout='PARENT_HASH'),  # rev-parse parent
            mock.Mock(stdout='hash1\n'),  # get_commit_sequence
        ]
        mock_resolve.return_value = 'hash1'
        mock_confirm.return_value = False
        
        rules = {'2026-04-15': ['path1']}
        git_refactor.refactor_history(rules, target_branch='main')
        mock_confirm.assert_called_with("\nInitialize Synthetic History Reconstruction?")

    @mock.patch('git_refactor.run_git')
    def test_refactor_history_dirty_workdir(self, mock_run_git):
        """Test refusal to run on dirty working directory."""
        mock_run_git.return_value = mock.Mock(stdout='M file.txt')
        with self.assertRaises(SystemExit) as cm:
            git_refactor.refactor_history({'date': ['path']}, target_branch='main')
        self.assertEqual(cm.exception.code, 1)


if __name__ == '__main__':
    unittest.main()

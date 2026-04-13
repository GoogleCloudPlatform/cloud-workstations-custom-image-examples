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

"""Unit tests for gitseep.py.

This module provides comprehensive test coverage for the interactive
GitSeep: Geological History Percolation Tool.
"""

import subprocess
import unittest
from unittest import mock
import gitseep


class TestGitseep(unittest.TestCase):
    """Test suite for gitseep.py functions."""

    @mock.patch("gitseep.subprocess.run")
    def test_run_git_success(self, mock_run):
        """Test successful git command execution."""
        mock_run.return_value = subprocess.CompletedProcess(
            args=["git", "status"], returncode=0, stdout="On branch main", stderr=""
        )
        result = gitseep.run_git(["status"])
        self.assertEqual(result.stdout, "On branch main")
        mock_run.assert_called_once()

    @mock.patch("gitseep.subprocess.run")
    def test_run_git_failure(self, mock_run):
        """Test git command failure raises GitError."""
        mock_run.side_effect = subprocess.CalledProcessError(
            returncode=1, cmd=["git", "invalid"], stderr="error message"
        )
        with self.assertRaises(gitseep.GitError):
            gitseep.run_git(["invalid"], check=True)

    @mock.patch("gitseep.run_git")
    def test_get_commit_sequence(self, mock_run_git):
        """Test retrieving commit sequence."""
        mock_run_git.return_value = mock.Mock(stdout="hash3\nhash2\nhash1\n")
        sequence = gitseep.get_commit_sequence("hash1")
        self.assertEqual(sequence, ["hash1", "hash2", "hash3"])
        mock_run_git.assert_called_with(["log", "--format=%H", "hash1^..HEAD"])

    @mock.patch("gitseep.run_git")
    def test_resolve_commit_by_date_success(self, mock_run_git):
        """Test resolving commit hash from date."""
        mock_run_git.side_effect = [
            mock.Mock(stdout="HASH123 2026-04-15 10:00:00 +0000\n"),
        ]
        result = gitseep.resolve_commit_by_date("2026-04-15 10:00:00 +0000")
        self.assertEqual(result, "HASH123")

    @mock.patch("gitseep.run_git")
    def test_resolve_commit_by_date_not_found(self, mock_run_git):
        """Test resolving commit hash from date when not found."""
        mock_run_git.return_value = mock.Mock(stdout="\n")
        with self.assertRaises(gitseep.GitError):
            gitseep.resolve_commit_by_date("2026-04-15 10:00:00 +0000")

    @mock.patch("gitseep.get_single_key")
    def test_user_confirm_yes(self, mock_get_key):
        """Test user confirmation with 'y'."""
        mock_get_key.return_value = "y"
        self.assertTrue(gitseep.user_confirm("Confirm?"))

    @mock.patch("gitseep.get_single_key")
    def test_user_confirm_no(self, mock_get_key):
        """Test user confirmation with 'n'."""
        mock_get_key.return_value = "n"
        self.assertFalse(gitseep.user_confirm("Confirm?"))

    @mock.patch("gitseep.get_git_root")
    @mock.patch("gitseep.os.chdir")
    @mock.patch("gitseep.run_git")
    @mock.patch("gitseep.resolve_commit_by_date")
    @mock.patch("gitseep.get_single_key")
    def test_seep_history_initialization(
        self, mock_get_key, mock_resolve, mock_run_git, mock_chdir, mock_root
    ):
        """Test percolation initialization without initial prompt."""
        mock_root.return_value = "/repo"

        def run_git_side_effect(args, **kwargs):
            cmd = " ".join(args)
            if "status" in cmd:
                return mock.Mock(stdout="")
            if "rev-parse HEAD" in cmd:
                return mock.Mock(stdout="HEAD_HASH")
            if "rev-parse --abbrev-ref" in cmd:
                return mock.Mock(stdout="main")
            if "log -1 --format=%at" in cmd:
                return mock.Mock(stdout="1000")
            if "rev-parse" in cmd:
                return mock.Mock(stdout="PARENT_HASH")
            if "log --format=%H" in cmd:
                return mock.Mock(stdout="hash1\n")
            if "log -1 --format=%ai|%s" in cmd:
                return mock.Mock(stdout="date|msg")
            if "log -1 --format=%B" in cmd:
                return mock.Mock(stdout="full message\n")
            if "--format=%an%n%ae%n%ad%n%cn%n%ce%n%cd" in cmd:
                return mock.Mock(stdout="an\nae\nad\ncn\nce\ncd\n")
            if "diff-tree" in cmd:
                return mock.Mock(stdout="A\tfile.txt\n")
            if "ls-tree" in cmd:
                return mock.Mock(stdout="file.txt")
            if "diff" in cmd:
                return mock.Mock(stdout="")
            return mock.Mock(stdout="")

        mock_run_git.side_effect = run_git_side_effect
        mock_resolve.return_value = "hash1"
        # Only the finalization prompt remains
        mock_get_key.return_value = "y"

        rules = {"2026-04-15 10:00:00 +0000": ["path1"]}
        gitseep.seep_history(rules, target_branch="main")

        # Verify it proceeded and reached finalization
        mock_get_key.assert_any_call("\nFinalize: Update branch 'main'? [Y/n]: ")

    @mock.patch("gitseep.get_git_root")
    @mock.patch("gitseep.os.chdir")
    @mock.patch("gitseep.run_git")
    def test_seep_history_dirty_workdir(self, mock_run_git, mock_chdir, mock_root):
        """Test refusal to run on dirty working directory."""
        mock_root.return_value = "/repo"
        mock_run_git.side_effect = [
            mock.Mock(stdout="M file.txt"),  # status
        ]
        with self.assertRaises(SystemExit) as cm:
            gitseep.seep_history({"date": ["path"]}, target_branch="main")
        self.assertEqual(cm.exception.code, 1)

    @mock.patch("gitseep.get_git_root")
    @mock.patch("gitseep.os.chdir")
    @mock.patch("gitseep.os.environ.copy")
    @mock.patch("gitseep.run_git")
    @mock.patch("gitseep.resolve_commit_by_date")
    @mock.patch("gitseep.get_single_key")
    @mock.patch("gitseep.force_path_state")
    def test_seep_history_success(
        self,
        mock_force,
        mock_get_key,
        mock_resolve,
        mock_run_git,
        mock_env,
        mock_chdir,
        mock_root,
    ):
        """Test successful history percolation."""
        mock_root.return_value = "/repo"
        mock_env.return_value = {}

        def run_git_side_effect(args, **kwargs):
            cmd = " ".join(args)
            if "status" in cmd:
                return mock.Mock(stdout="")
            if "rev-parse HEAD" in cmd:
                return mock.Mock(stdout="ORIGINAL_HEAD")
            if "rev-parse --abbrev-ref" in cmd:
                return mock.Mock(stdout="main")
            if "log -1 --format=%at" in cmd:
                return mock.Mock(stdout="1000")
            if "rev-parse hash1^" in cmd:
                return mock.Mock(stdout="PARENT_HASH")
            if "log --format=%H" in cmd:
                return mock.Mock(stdout="hash1\n")
            if "log -1 --format=%ai|%s" in cmd:
                return mock.Mock(stdout="date|msg")
            if "diff-tree" in cmd:
                return mock.Mock(stdout="path1/file1.txt\n")
            if "log -1 --format=%B" in cmd:
                return mock.Mock(stdout="Message\n")
            if "--format=%an%n%ae%n%ad%n%cn%n%ce%n%cd" in cmd:
                return mock.Mock(stdout="an\nae\nad\ncn\nce\ncd\n")
            if "read-tree" in cmd:
                return mock.Mock(stdout="")
            if "commit" in cmd:
                return mock.Mock(stdout="")
            if "diff" in cmd:
                return mock.Mock(stdout="")
            return mock.Mock(stdout="")

        mock_run_git.side_effect = run_git_side_effect
        mock_resolve.return_value = "hash1"
        mock_get_key.return_value = "y"

        rules = {"2026-04-15 10:00:00 +0000": ["path1"]}
        gitseep.seep_history(rules, target_branch="main")

        # Verify key operations
        mock_run_git.assert_any_call(["read-tree", "-u", "--reset", "hash1"])
        mock_force.assert_any_call("path1", "ORIGINAL_HEAD")
        mock_run_git.assert_any_call(
            ["commit", "--allow-empty", "-C", "hash1"], env=mock.ANY
        )

    @mock.patch("gitseep.get_git_root")
    @mock.patch("gitseep.os.chdir")
    @mock.patch("gitseep.run_git")
    @mock.patch("gitseep.resolve_commit_by_date")
    def test_seep_history_compression_error(
        self, mock_resolve, mock_run_git, mock_chdir, mock_root
    ):
        """Test exit code 3 on unpermitted compression with auto_approve."""
        mock_root.return_value = "/repo"
        mock_resolve.return_value = "bedrock_h"

        def run_git_side_effect(args, **kwargs):
            cmd = " ".join(args)
            if "status" in cmd:
                return mock.Mock(stdout="")
            if "rev-parse HEAD" in cmd:
                return mock.Mock(stdout="ORIGINAL_HEAD")
            if "rev-parse --abbrev-ref" in cmd:
                return mock.Mock(stdout="main")
            if "log -1 --format=%at" in cmd:
                return mock.Mock(stdout="1000")
            if "rev-parse" in cmd:
                return mock.Mock(stdout="PARENT_HASH")
            if "log --format=%H" in cmd:
                return mock.Mock(stdout="bedrock_h\nsource_h\n")
            if "log -1 --format=%ai|%s" in cmd:
                return mock.Mock(stdout="date|msg")
            if "diff-tree" in cmd:
                return mock.Mock(stdout="file.txt\n")
            if "log -1 --format=%B" in cmd:
                return mock.Mock(stdout="Msg\n")
            if "--format=%an%n%ae%n%ad%n%cn%n%ce%n%cd" in cmd:
                return mock.Mock(stdout="an\nae\nad\ncn\nce\ncd\n")
            return mock.Mock(stdout="")

        mock_run_git.side_effect = run_git_side_effect

        rules = {"date": ["file.txt"]}
        with self.assertRaises(SystemExit) as cm:
            gitseep.seep_history(rules, "main", auto_approve=True)
        self.assertEqual(cm.exception.code, 3)

    @mock.patch("gitseep.get_git_root")
    @mock.patch("gitseep.os.chdir")
    @mock.patch("gitseep.run_git")
    @mock.patch("gitseep.resolve_commit_by_date")
    @mock.patch("gitseep.get_single_key")
    def test_seep_history_compression_automated_diff(
        self, mock_get_key, mock_resolve, mock_run_git, mock_chdir, mock_root
    ):
        """Test that compression automatically shows diffs."""
        mock_root.return_value = "/repo"
        mock_resolve.return_value = "bedrock_h"

        # Side effects for user confirm: 
        # 1. 'y' for permit compression
        # 2. 'y' for show parity diff
        # 3. 'y' for finalization
        mock_get_key.side_effect = ["y", "y", "y"]

        def run_git_side_effect(args, **kwargs):
            cmd = " ".join(args)
            if "status" in cmd:
                return mock.Mock(stdout="")
            if "rev-parse HEAD" in cmd:
                return mock.Mock(stdout="ORIGINAL_HEAD")
            if "rev-parse --abbrev-ref" in cmd:
                return mock.Mock(stdout="main")
            if "log -1 --format=%at" in cmd:
                return mock.Mock(stdout="1000")
            if "rev-parse" in cmd:
                return mock.Mock(stdout="PARENT_HASH")
            if "log --format=%H" in cmd:
                return mock.Mock(stdout="bedrock_h\nsource_h\n")
            if "log -1 --format=%ai|%s" in cmd:
                return mock.Mock(stdout="date|msg")
            if "diff-tree" in cmd:
                return mock.Mock(stdout="file.txt\n")
            if "log -1 --format=%B" in cmd:
                return mock.Mock(stdout="Msg\n")
            if "log -1 --format=%s" in cmd:
                return mock.Mock(stdout="Subject")
            if "diff" in cmd:
                if "--shortstat" in cmd:
                    return mock.Mock(stdout=" 1 file changed, 1 insertion(+)")
                return mock.Mock(stdout="diff content")
            if "--format=%an%n%ae%n%ad%n%cn%n%ce%n%cd" in cmd:
                return mock.Mock(stdout="an\nae\nad\ncn\nce\ncd\n")
            return mock.Mock(stdout="")

        mock_run_git.side_effect = run_git_side_effect

        rules = {"date": ["file.txt"]}
        gitseep.seep_history(rules, "main")

        # Verify automated diff was called
        diff_called = any(
            "diff" in call[0][0] and "--color" in call[0][0]
            for call in mock_run_git.call_args_list
        )
        self.assertTrue(diff_called)

    @mock.patch("gitseep.run_git")
    def test_calculate_percolation_schedule(self, mock_run_git):
        """Test calculation of percolation schedule."""
        ctx = gitseep.SeepageContext(
            repo_root="/repo",
            original_head="HEAD_HASH",
            current_branch="main",
            target_branch="main",
            parent_of_strata="PARENT_HASH",
            strata=["h1", "h2"],
            resolved_rules={"h1": ["path1"]},
            path_to_bedrock={"path1": "h1"},
            all_files=False,
            auto_approve=False,
            compress=False,
            stage_only=False,
        )

        def run_git_side_effect(args, **kwargs):
            if "diff-tree" in args:
                if "h1" in args:
                    return mock.Mock(stdout="path1/file1.txt\n")
                if "h2" in args:
                    return mock.Mock(stdout="path1/file2.txt\n")
            return mock.Mock(stdout="")

        mock_run_git.side_effect = run_git_side_effect

        schedule, sources = gitseep._calculate_percolation_schedule(ctx)

        self.assertIn("h1", schedule)
        self.assertEqual(schedule["h1"]["h2"], ["path1/file2.txt"])
        self.assertEqual(sources["path1/file1.txt"], {"h1"})


if __name__ == "__main__":
    unittest.main()

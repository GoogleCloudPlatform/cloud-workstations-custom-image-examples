#!/usr/bin/env bats

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

load test_helper.bash

@test "Verify new directory structure: examples/ exists" {
  [ -d "${REPO_ROOT}/examples" ]
}

@test "Verify new directory structure: examples/images/gnome exists" {
  [ -d "${REPO_ROOT}/examples/images/gnome" ]
}

@test "Verify new directory structure: examples/images/antigravity exists" {
  [ -d "${REPO_ROOT}/examples/images/antigravity" ]
}

@test "Verify new directory structure: examples/preflight exists" {
  [ -d "${REPO_ROOT}/examples/preflight" ]
}

@test "Ensure old apps/ directory does not exist" {
  [ ! -d "${REPO_ROOT}/apps" ]
}

@test "Ensure old preflight/ directory does not exist at top level" {
  # We know preflight exists, but it should be under examples/
  # The find command in previous turns showed it moved.
  [ ! -d "${REPO_ROOT}/preflight" ]
}

@test "Check for hardcoded old paths in shell scripts" {
  # Search for apps/ or preflight/ but exclude those preceded by examples/
  # We use grep -v to filter out the correct ones.
  run bash -c "grep -rE '(\./|/|^)(apps/|preflight/)' '${REPO_ROOT}' \
    --exclude-dir='.git' \
    --exclude-dir='.gemini' \
    --exclude='*.bats' \
    --exclude='*.md' \
    --exclude='sbom.json' \
    --exclude='.pre-commit-config.yaml' | grep -v 'examples/'"

  [ "$status" -ne 0 ]
}

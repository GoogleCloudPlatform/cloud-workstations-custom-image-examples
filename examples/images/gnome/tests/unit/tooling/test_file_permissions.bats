#!/usr/bin/env bats

# Copyright 2025-2026 Google LLC
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

load ../test_helper.bash
setup() {
  skip_if_slow
  IMAGE_DIRS=("examples/images/gnome" "examples/images/antigravity" "examples/preflight")
}

@test "All .sh files in blueprint are executable" {
  local fail=0
  local non_executable_scripts=""

  for dir in "${IMAGE_DIRS[@]}"; do
    local full_path="${REPO_ROOT}/${dir}"
...
    if [[ ! -d "${full_path}" ]]; then
        continue
    fi

    # Find all .sh files that are NOT executable, excluding node_modules
    local found
    found=$(find "${full_path}" -name "*.sh" -not -path "*/node_modules/*" -not -executable)

    if [[ -n "${found}" ]]; then
      non_executable_scripts="${non_executable_scripts}${found}\n"
      fail=1
    fi
  done

  if [[ "${fail}" -eq 1 ]]; then
    echo "Found non-executable .sh files:"
    echo -e "${non_executable_scripts}"
    return 1
  fi
}

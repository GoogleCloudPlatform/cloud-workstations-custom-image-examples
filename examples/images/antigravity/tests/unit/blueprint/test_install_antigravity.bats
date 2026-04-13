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

load ../test_helper.bash

setup() {
  # Mock apt-get
  apt-get() {
    echo "MOCK_APT: $*"
  }
  export -f apt-get

  # Mock sed (use a wrapper to avoid breaking other things if possible, but BATS run is isolated)
  # Actually, we can just mock the script execution by copying it and sed-ing it
}

@test "10_install_antigravity.sh configures repo and installs package" {
  local temp_dir="${BATS_FILE_TMPDIR}/mock_etc"
  mkdir -p "${temp_dir}/apt/sources.list.d"
  local mock_list="${temp_dir}/apt/sources.list.d/antigravity.list"
  echo "deb http://example.com/antigravity us-central1 main" > "${mock_list}"

  # Copy script and patch it to use our mock paths
  local test_script="${BATS_FILE_TMPDIR}/test_install.sh"
  cp "${HOOKS_DIR}/10_install_antigravity.sh" "${test_script}"
  sed -i "s|/etc/apt/sources.list.d/antigravity.list|${mock_list}|g" "${test_script}"

  export GCP_REGION="europe-west1"
  run bash "${test_script}"

  [ "$status" -eq 0 ]
  [[ "$output" =~ "Configuring Antigravity repository for region: europe-west1" ]]
  [[ "$output" =~ "Installing Antigravity package" ]]
  [[ "$output" =~ "MOCK_APT: update" ]]
  [[ "$output" =~ "MOCK_APT: install -y --no-install-recommends antigravity" ]]

  # Verify sed worked on the mock list
  grep -q "europe-west1" "${mock_list}"
}

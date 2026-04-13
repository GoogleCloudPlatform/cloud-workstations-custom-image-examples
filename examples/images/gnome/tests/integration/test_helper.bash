#!/usr/bin/env bash

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

# Helper functions for BATS integration tests

# We use the helper's own location as a stable anchor.
HELPER_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Find the repository root by looking for the skills directory
find_repo_root() {
  local dir="${1}"
  while [[ "${dir}" != "/" ]]; do
    if [[ -f "${dir}/skills/common.sh" ]]; then
      echo "${dir}"
      return 0
    fi
    dir="$(dirname "${dir}")"
  done
  return 1
}

PROJECT_ROOT="$(find_repo_root "${HELPER_DIR}")"

# Source common.sh from the project root
# shellcheck disable=SC1091
source "${PROJECT_ROOT}/skills/common.sh"

setup_env() {
  export PROJECT="${PROJECT:-YOUR_PROJECT_ID}"
  export CLUSTER="${CLUSTER:-YOUR_CLUSTER_NAME}"
  export CONFIG="${CONFIG:-YOUR_CONFIG_NAME}"
  export REGION="${REGION:-YOUR_REGION}"
  export WORKSTATION="${WORKSTATION:-YOUR_WORKSTATION_NAME}"
  export EXTRA_BINARIES="${EXTRA_BINARIES:-}"

  if [[ "${PROJECT}" == "YOUR_PROJECT_ID" ]]; then
    skip "Missing integration test configuration. Please set variables or provide a .env file."
  fi
}

run_ssh() {
  local cmd="$1"
  local output

  output=$(gcloud workstations ssh \
    --project="${PROJECT}" \
    --cluster="${CLUSTER}" \
    --config="${CONFIG}" \
    --region="${REGION}" \
    "${WORKSTATION}" \
    --command="${cmd} && echo _CWS_CMD_SUCCESS_" 2>&1)

  echo "$output"
  if [[ "$output" == *_CWS_CMD_SUCCESS_* ]]; then
    return 0
  else
    return 1
  fi
}

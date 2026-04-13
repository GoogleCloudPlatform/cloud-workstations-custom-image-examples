#!/bin/bash

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

set -euo pipefail

# Sourced root logic
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "${SCRIPT_DIR}/../../common.sh"

main() {
  log "======================================"
  log " 🚀 Running All Local Blueprint Tests (via pre-commit)"
  log "======================================"

  cd "${REPO_ROOT}"
  if command -v pre-commit > /dev/null 2>&1; then
    pre-commit run --all-files
  elif [[ -x "$HOME/.local/bin/pre-commit" ]]; then
    "$HOME/.local/bin/pre-commit" run --all-files
  else
    error "pre-commit not found. Please install it via 'apt-get install pre-commit' or 'pipx install pre-commit'."
  fi

  log "======================================"
  log " 🎉 All local blueprint checks passed!"
  log "======================================"
}

main "$@"

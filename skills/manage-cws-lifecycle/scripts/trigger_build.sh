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

# This script triggers a manual Cloud Build for the gnome project using an
# ephemeral configuration. It submits the build from the repository root
# context to ensure all dependencies (like preflight) are available.

main() {
  # Validate required variables
  local missing_vars=()
  [[ -z "${PROJECT:-}" ]] && missing_vars+=("PROJECT")
  [[ -z "${BUILD_REGION:-}" ]] && missing_vars+=("BUILD_REGION")

  if [[ ${#missing_vars[@]} -gt 0 ]]; then
    error "Missing required variables in .env: ${missing_vars[*]}"
  fi

  log "======================================"
  log " 🚀 Triggering Ephemeral Cloud Build"
  log "======================================"
  log "Project:   ${PROJECT}"
  log "Build Reg: ${BUILD_REGION}"
  log "Artif Reg: ${ARTIFACT_REGION}"
  log "Context:   ${REPO_ROOT}"
  log "======================================"

  # Create an ephemeral ignore file in /run/user for better transience
  local ignore_file
  ignore_file=$(mktemp --tmpdir="/run/user/$(id -u)")
  cat > "${ignore_file}" <<EOF
.git/
.gitignore
.gcloudignore
**/node_modules/
**/dist/
**/.gemini/
**/*~
**/*.log
.vscode/
.idea/
EOF

  # Submit the build using the repository context and an ephemeral inline config
  # Use the ephemeral ignore file and ensure it is removed afterwards
  gcloud builds submit "${REPO_ROOT}" \
    --project="${PROJECT}" \
    --region="${BUILD_REGION}" \
    --ignore-file="${ignore_file}" \
    --substitutions=_SKAFFOLD_DEFAULT_REPO="${ARTIFACT_REGION}-docker.pkg.dev/${PROJECT}/cicd-foundation" \
    --config=<(cat <<EOF
steps:
- name: 'gcr.io/google.com/cloudsdktool/cloud-sdk:latest'
  entrypoint: 'bash'
  args:
  - '-c'
  - |
    # Install skaffold
    curl -Lo skaffold https://storage.googleapis.com/skaffold/releases/latest/skaffold-linux-amd64
    chmod +x skaffold
    mv skaffold /usr/local/bin/

    # Run skaffold build from the app context
    cd examples/images/gnome
    skaffold build --default-repo=\$_SKAFFOLD_DEFAULT_REPO
EOF
)

  rm -f "${ignore_file}"
}

main "$@"

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

# This script triggers a manual Cloud Build for a specific workstation image
# (defaulting to gnome) using an ephemeral configuration. It submits the build
# from the repository root context to ensure all dependencies (like preflight)
# are available.

usage() {
  echo "Usage: $0 [IMAGE_PATH] [--service-account SA_EMAIL]"
  echo "Example: $0 examples/images/android-studio-for-platform"
  echo "Defaults:"
  echo "  IMAGE_PATH: examples/images/gnome"
  echo "  SA_EMAIL:   cloudbuild@\${PROJECT}.iam.gserviceaccount.com"
  exit 1
}

main() {
  local image_path=""
  local service_account="${SERVICE_ACCOUNT:-}"

  # Parse arguments
  while [[ $# -gt 0 ]]; do
    case "$1" in
      --service-account)
        service_account="$2"
        shift 2
        ;;
      -h|--help)
        usage
        ;;
      *)
        if [[ -z "$image_path" ]]; then
          image_path="$1"
        else
          error "Unknown argument: $1"
        fi
        shift
        ;;
    esac
  done

  # Defaults
  image_path="${image_path:-examples/images/gnome}"

  # Validate required variables
  local missing_vars=()
  [[ -z "${PROJECT:-}" ]] && missing_vars+=("PROJECT")
  [[ -z "${BUILD_REGION:-}" ]] && missing_vars+=("BUILD_REGION")

  if [[ ${#missing_vars[@]} -gt 0 ]]; then
    error "Missing required variables in .env: ${missing_vars[*]}"
  fi

  # Final SA default (just the email/ID part)
  service_account="${service_account:-cloudbuild@${PROJECT}.iam.gserviceaccount.com}"

  log "======================================"
  log " 🚀 Triggering Ephemeral Cloud Build"
  log "======================================"
  log "Project:   ${PROJECT}"
  log "Build Reg: ${BUILD_REGION}"
  log "Artif Reg: ${ARTIFACT_REGION}"
  log "Image Path: ${image_path}"
  log "Service Account: ${service_account}"
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
    --service-account="projects/${PROJECT}/serviceAccounts/${service_account}" \
    --ignore-file="${ignore_file}" \
    --substitutions=_SKAFFOLD_DEFAULT_REPO="${ARTIFACT_REGION}-docker.pkg.dev/${PROJECT}/cicd-foundation" \
    --config=<(cat <<EOF
options:
  logging: CLOUD_LOGGING_ONLY
  machineType: E2_HIGHCPU_32
timeout: 7200s
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
    cd ${image_path}
    skaffold build \
      --default-repo=\$_SKAFFOLD_DEFAULT_REPO \
      --cache-artifacts=true
EOF
)

  rm -f "${ignore_file}"
}

main "$@"

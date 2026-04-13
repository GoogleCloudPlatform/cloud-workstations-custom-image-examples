#!/bin/bash

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

set -euo pipefail

# This hook installs the Antigravity application.
# It is executed during the Docker build process of the layered image.

main() {
  local gcp_region="${GCP_REGION:-us-central1}"
  local apt_list="/etc/apt/sources.list.d/antigravity.list"

  echo "Configuring Antigravity repository for region: ${gcp_region}..."
  if [[ -f "${apt_list}" ]]; then
    # Replace the default us-central1 with the actual build-time region.
    sed -i "s/us-central1/${gcp_region}/g" "${apt_list}"
  else
    echo "Error: ${apt_list} not found. Ensure assets are correctly copied."
    exit 1
  fi

  echo "Installing Antigravity package..."
  apt-get update
  apt-get install -y --no-install-recommends antigravity

  echo "Antigravity installation complete."
}

main "$@"

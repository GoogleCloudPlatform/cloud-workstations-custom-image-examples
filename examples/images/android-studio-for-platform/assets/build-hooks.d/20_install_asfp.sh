#!/bin/bash

# Copyright 2024-2026 Google LLC
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

# Sourced logic (from preflight or local)
# shellcheck disable=SC1091
source "/google/scripts/common.sh"

install_asfp() {
  local asfp_version="${ASFP_VERSION:-canary}"
  local curl_opts="${CURL_OPTS:--fsSL --retry 3}"
  local package

  echo "Installing Android Studio for Platform (version: ${asfp_version})..."

  if [[ "${asfp_version}" == "canary" ]]; then
    package="asfp-canary-current-linux.deb"
  else
    package="asfp-current-linux.deb"
  fi

  # shellcheck disable=SC2086
  curl ${curl_opts} -O https://dl.google.com/android/asfp/${package}
  dpkg -i "${package}"
  rm "${package}"

  # Update desktop entry to canary if applicable
  if [[ "${asfp_version}" == "canary" ]] && [[ -f /usr/share/applications/jetbrains-studio.desktop ]]; then
    sed -i 's/android-studio-for-platform/android-studio-for-platform-canary/g' /usr/share/applications/jetbrains-studio.desktop
  fi
}

main() {
  install_asfp
}

if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
  main "$@"
fi
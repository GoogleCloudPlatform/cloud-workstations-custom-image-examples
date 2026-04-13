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

# This hook dynamically scans for desktop entries and enables autostart
# for those containing the 'X-GNOME-Autostart-enabled=true' marker.

main() {
  local apps_dir="/usr/share/applications"
  local autostart_dir="/etc/xdg/autostart"

  if [[ ! -d "${apps_dir}" ]]; then
    return 0
  fi

  mkdir -p "${autostart_dir}"

  echo "Scanning for applications to autostart..."

  # Find all .desktop files that have the autostart marker enabled
  # We use grep to identify candidates and then create symlinks.
  grep -l "X-GNOME-Autostart-enabled=true" "${apps_dir}"/*.desktop 2>/dev/null | while read -r desktop_file; do
    local filename
    filename=$(basename "${desktop_file}")
    
    echo "  -> Enabling autostart for: ${filename}"
    ln -sf "${desktop_file}" "${autostart_dir}/${filename}"
  done
}

main "$@"

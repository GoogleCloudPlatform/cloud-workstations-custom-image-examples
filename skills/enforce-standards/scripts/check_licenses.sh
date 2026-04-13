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

# This script checks all tracked files for the Apache 2.0 license header.
# It ensures exactly one full header exists and removes redundant short headers.

main() {
  local repo_root
  repo_root=$(git rev-parse --show-toplevel)
  cd "${repo_root}"

  echo "🔍 Checking and applying Apache 2.0 License Headers..."

  local files_to_check
  files_to_check=$(git ls-files | grep -iE '\.(ts|js|sh|css|html|md|bash|bats)$' || true)

  local current_year
  current_year=$(date +%Y)

  local license_body
  read -r -d '' license_body << BODY_EOF || true
Copyright ${current_year} Google LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
BODY_EOF

  local modified_files=()
  local file
  for file in ${files_to_check}; do
    # Skip vendor/build directories
    if [[ "${file}" == *"node_modules"* || "${file}" == *"dist/"* || "${file}" == *"licenses/"* ]]; then
      continue
    fi

    if [[ -L "${file}" || ! -f "${file}" ]]; then
      continue
    fi

    # Skip the script itself
    if [[ "${file}" == *"check_licenses.sh" ]]; then
      continue
    fi

    local ext="${file##*.}"
    local tmp
    tmp=$(mktemp)
    cp "${file}" "${tmp}"

    # 1. Remove all short headers
    # TS/JS/CSS style
    perl -0777 -i -pe 's/\/\*\*\s*\n\s*\* Copyright.*?\n\s*\*\/(\s*\n)?//gs' "${tmp}"
    # HTML/MD style
    perl -0777 -i -pe 's/<!--\s*\n\s*Copyright.*?\n\s*-->(\s*\n)?//gs' "${tmp}"

    # 2. Check for full license
    if ! grep -qi "Licensed under the Apache License, Version 2.0" "${tmp}"; then
      echo "  - Applying full license to ${file}"
      local final_tmp
      final_tmp=$(mktemp)

      if [[ "${ext}" == "sh" || "${ext}" == "bash" || "${ext}" == "bats" ]]; then
        local first_line
        first_line=$(head -n 1 "${tmp}" || true)
        if [[ "${first_line}" == "#!"* ]]; then
          echo "${first_line}" > "${final_tmp}"
          echo "" >> "${final_tmp}"
          echo "${license_body}" | sed 's/^/# /' >> "${final_tmp}"
          echo "" >> "${final_tmp}"
          tail -n +2 "${tmp}" >> "${final_tmp}"
        else
          echo "${license_body}" | sed 's/^/# /' > "${final_tmp}"
          echo "" >> "${final_tmp}"
          cat "${tmp}" >> "${final_tmp}"
        fi
      elif [[ "${ext}" == "html" || "${ext}" == "md" ]]; then
        echo "<!--" > "${final_tmp}"
        echo "${license_body}" >> "${final_tmp}"
        echo "-->" >> "${final_tmp}"
        echo "" >> "${final_tmp}"
        cat "${tmp}" >> "${final_tmp}"
      elif [[ "${ext}" == "css" || "${ext}" == "ts" || "${ext}" == "js" ]]; then
        echo "/**" > "${final_tmp}"
        echo "${license_body}" | sed 's/^/ * /' | sed 's/ \* $/\ \*/' >> "${final_tmp}"
        echo " */" >> "${final_tmp}"
        echo "" >> "${final_tmp}"
        cat "${tmp}" >> "${final_tmp}"
      fi
      mv "${final_tmp}" "${tmp}"
    fi

    # Remove extra leading blank lines
    perl -0777 -i -pe 's/^(\n)+//' "${tmp}"

    if ! diff -q "${file}" "${tmp}" > /dev/null; then
      mv "${tmp}" "${file}"
      modified_files+=("${file}")
    else
      rm "${tmp}"
    fi
  done

  if [[ ${#modified_files[@]} -gt 0 ]]; then
    echo "✅ Applied license headers or fixes to ${#modified_files[@]} files."
  else
    echo "✅ All tracked files already have valid license headers and formatting."
  fi
}

main "$@"

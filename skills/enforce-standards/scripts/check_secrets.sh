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

# This script checks for potential secrets in the codebase.
# It uses gitleaks if available, otherwise it falls back to grep.

main() {
  echo "🔍 Checking for potential secrets..."

  if command -v gitleaks &> /dev/null; then
    echo "  - Running gitleaks..."
    gitleaks detect --source . -v --redact --no-banner
  else
    echo "  ⚠️  gitleaks not found. Falling back to basic pattern search..."

    # Define common secret patterns
    # These are basic and may produce false positives, but it's better than nothing.
    local patterns=(
      "AIza[0-9A-Za-z-_]{35}" # GCP API Key
      "AKIA[0-9A-Z]{16}"      # AWS Access Key
      "(\"|')[0-9a-zA-Z]{40}(\"|')" # Potential Secret Key
      "-----BEGIN (RSA|EC|DSA|OPENSSH|PGP) PRIVATE KEY-----"
    )

    local found_secrets=0
    local pattern
    for pattern in "${patterns[@]}"; do
      # Search all tracked files that exist
      if git ls-files | xargs -r ls -d 2>/dev/null | xargs -r grep -E -e "${pattern}" --color=always; then
        found_secrets=1
      fi
    done

    if [[ "${found_secrets}" -eq 1 ]]; then
      echo "❌ Potential secrets found! Please review and remove them before committing."
      exit 1
    fi
  fi

  echo "✅ No secrets detected."
}

main "$@"

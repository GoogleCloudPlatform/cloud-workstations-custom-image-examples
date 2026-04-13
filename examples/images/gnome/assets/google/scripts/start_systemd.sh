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

# This script handles final dynamic service enablement and starts systemd.

set -euo pipefail

# Source common utilities
# shellcheck source=/dev/null
source /google/scripts/common.sh

# Propagates required environment variables to the systemd manager.
# This ensures ConditionEnvironment= checks work correctly for PID 1.
propagate_env() {
  local env_conf="/etc/systemd/system.conf.d/10-env.conf"
  mkdir -p "$(dirname "${env_conf}")"

  log "Propagating environment variables to systemd manager..."
  {
    echo "[Manager]"
    echo -n "DefaultEnvironment="
    # Capture ENABLE_ and DEFAULT_ variables from the current environment
    env | grep -E '^(ENABLE_|DEFAULT_|GCP_REGION)' | xargs || true
  } > "${env_conf}"
}

# Executes any runtime customization scripts provided in /etc/workstation-startup.d/
run_runtime_hooks() {
  local hook_dir="/etc/workstation-startup.d"
  if [[ -d "${hook_dir}" ]]; then
    log "Executing runtime hooks from ${hook_dir}..."
    local hook
    for hook in "${hook_dir}"/*.sh; do
      [[ -e "${hook}" ]] || continue

      # SECURITY VALIDATION:
      # 1. Must be owned by root.
      # 2. Must not be world-writable (666 or 777).
      local stat_info
      stat_info=$(stat -c "%u %a" "${hook}")
      local owner_uid=${stat_info% *}
      local permissions=${stat_info#* }

      if [[ "${owner_uid}" != "0" ]]; then
        log "error: hook ${hook} is not owned by root (owner UID: ${owner_uid}). skipping."
        continue
      fi

      # Check if world-writable bit (octal 2) is set in the last digit of permissions.
      if (( (8#${permissions} & 0002) != 0 )); then
        log "error: hook ${hook} is world-writable (permissions: ${permissions}). skipping."
        continue
      fi

      if [[ -x "${hook}" ]]; then
        log "Running hook: ${hook}"
        "${hook}" || log "warning: hook ${hook} failed with exit code $?"
      elif [[ -f "${hook}" ]]; then
        log "Running hook (via bash): ${hook}"
        /bin/bash "${hook}" || log "warning: hook ${hook} failed with exit code $?"
      fi
    done
  fi
}

main() {
  run_runtime_hooks
  propagate_env

  # Start systemd with the explicitly defined machine id (inherited from entrypoint)
  log "starting systemd"
  exec /sbin/init --system --unit=multi-user.target --machine-id "${MACHINE_ID:-}"
}

main "$@"

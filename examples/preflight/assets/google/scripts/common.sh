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

# Common utility functions and centralized configuration registry for
# Google Cloud Workstations scripts.
#
# Source of Truth Hierarchy:
# 1. Dockerfile (ARG/ENV): Sets build-time and early-runtime defaults.
# 2. common.sh (Library): Aggregates ENVs, provides calculated defaults,
#    and serves as the shared registry for all layer scripts.
# 3. Layer Scripts: Source this file to ensure consistent behavior.

# User Identity Defaults
export WORKSTATION_USER="${WORKSTATION_USER:-user}"
export WORKSTATION_UID="${WORKSTATION_UID:-1000}"

# Centralized paths for ephemeral credentials
export EPHEMERAL_ENV_DIR="/tmp/workstation"
export EPHEMERAL_ENV_FILENAME="ephemeral.env"
export EPHEMERAL_ENV_PATH="${EPHEMERAL_ENV_DIR}/${EPHEMERAL_ENV_FILENAME}"

# RDP Connection Registry
export RDP_HOST="127.0.0.1"
export RDP_PORT="3389"

# SSH Connection Registry
export SSH_HOST="127.0.0.1"
export SSH_PORT="22"

# Protocol and UI Defaults (Mirroring Dockerfile ENVs)
export ENABLE_TIGERVNC="${ENABLE_TIGERVNC:-false}"
export ENABLE_CRD="${ENABLE_CRD:-false}"
export DEFAULT_ENABLE_AUDIO_INPUT="${DEFAULT_ENABLE_AUDIO_INPUT:-false}"
export DEFAULT_PROTOCOL="${DEFAULT_PROTOCOL:-RDP}"
export DEFAULT_TIMEOUT_MS="${DEFAULT_TIMEOUT_MS:-200000}"

# Logs a message to stdout with a timestamp and hostname.
# Arguments:
#   $1: The message to log.
log() {
  echo "$(date -u +'%b %d %H:%M:%S') $(hostname) $(basename "$0")[$$]: $1"
}

# Helper to verify if a file exists and is readable.
# Arguments:
#   $1: File path.
check_file() {
  if [[ ! -r "$1" ]]; then
    log "error: file not found or not readable: $1"
    return 1
  fi
  return 0
}

# Logs a warning message to stdout
warn() {
  echo -e "[\033[0;33mWARN\033[0m] $*"
}

# Logs an error message to stderr and exits
error() {
  echo -e "[\033[0;31mERROR\033[0m] $*" >&2
  exit 1
}

# Global Retry Configuration (can be overridden by ENVs or Args)
export RETRY_ATTEMPTS="${RETRY_ATTEMPTS:-3}"
export RETRY_DELAY="${RETRY_DELAY:-5}"

# Retries a command multiple times with exponential backoff.
# Usage: retry [ATTEMPTS] [INITIAL_DELAY] "command to run"
retry() {
  local attempts
  local delay
  local command

  # Support both legacy "retry 3 5 'cmd'" and "retry 'cmd'" usages.
  if [[ $# -eq 1 ]]; then
    attempts="${RETRY_ATTEMPTS}"
    delay="${RETRY_DELAY}"
    command="${1}"
  else
    attempts="${1}"
    delay="${2}"
    command="${3}"
  fi

  local current_delay="${delay}"

  for i in $(seq 1 "${attempts}"); do
    if eval "${command}"; then
      return 0
    fi
    if [[ "${i}" -lt "${attempts}" ]]; then
      warn "Command failed (Attempt ${i}/${attempts}). Retrying in ${current_delay}s: ${command}"
      sleep "${current_delay}"
      current_delay=$((current_delay * 2))
    else
      error "Command failed after ${attempts} attempts: ${command}"
    fi
  done
}

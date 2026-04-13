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

# This script handles one-time user and session setup tasks.

set -euo pipefail

# Max seconds to wait for gnome-shell to recognize extensions
EXTENSION_READY_TIMEOUT=15

# Source common utilities
# shellcheck source=/dev/null
source /google/scripts/common.sh

# Waits for a GNOME extension to become available in the shell.
wait_for_extension() {
  local ext_id="${1}"
  local start_time
  start_time=$(date +%s)

  log "Waiting for extension ${ext_id} to become available..."
  while true; do
    if runuser -u user -- bash -c "export XDG_RUNTIME_DIR=/run/user/1000; export DBUS_SESSION_BUS_ADDRESS=unix:path=/run/user/1000/bus; gnome-extensions list" | grep -q "${ext_id}"; then
      log "Extension ${ext_id} is now available."
      return 0
    fi

    local current_time
    current_time=$(date +%s)
    if (( current_time - start_time > EXTENSION_READY_TIMEOUT )); then
      log "Warning: Timeout waiting for extension ${ext_id}."
      return 1
    fi
    sleep 0.5
  done
}

# Sets up the user-specific runtime directory (XDG_RUNTIME_DIR).
setup_runtime_dir() {
  local xdg_runtime_dir="/run/user/1000"
  log "Setting up XDG_RUNTIME_DIR at ${xdg_runtime_dir}..."
  mkdir -p "${xdg_runtime_dir}"
  chown 1000:1000 "${xdg_runtime_dir}"
  chmod 700 "${xdg_runtime_dir}"
}

# Ensures correct ownership of the user home directory.
setup_home_dir() {
  log "Fixing ownership for /home/user..."

  # Ensure user is in the audio group for sound support
  usermod -aG audio user || true

  # Clean up legacy persistent configurations that break the updated headless shell
  rm -f /home/user/.config/systemd/user/org.gnome.Shell@wayland.service.d/override.conf
  rm -f /home/user/.local/share/applications/org.gnome.Shell.desktop

  # Use -h to avoid dereferencing symlinks (prevents failure on broken pulse/machine-id links)
  chown -R -h 1000:1000 /home/user
}

# Generates RDP TLS certificates if necessary.
setup_rdp_certs() {
  log "Checking RDP TLS certificates..."
  local grd_cert_dir="/home/user/.local/share/gnome-remote-desktop"
  if command -v winpr-makecert3 >/dev/null 2>&1 || command -v winpr-makecert >/dev/null 2>&1; then
    if [[ ! -f "${grd_cert_dir}/rdp-tls.crt" ]]; then
      log "Generating new RDP TLS certificate..."
      mkdir -p "${grd_cert_dir}"
      local make_cert
      make_cert=$(command -v winpr-makecert3 || command -v winpr3-makecert || command -v winpr-makecert)
      runuser -u user -- "${make_cert}" -silent -rdp -path "${grd_cert_dir}" rdp-tls
      chown -R 1000:1000 "${grd_cert_dir}"
    fi
  fi
}

# Initializes the GNOME keyring with a blank password.
setup_keyring() {
  if command -v gnome-keyring-daemon >/dev/null 2>&1; then
    log "Initializing gnome-keyring..."
    local keyring_dir="/home/user/.local/share/keyrings"

    # The home directory is persistent, but our password is ephemeral.
    # We MUST nuke the old keyrings on boot, otherwise the old password
    # remains and the daemon cannot be unlocked, breaking RDP sync.
    rm -rf "${keyring_dir}"
    mkdir -p "${keyring_dir}"

    # Ensure a default keyring exists and is unlocked
    # We use a blank password for the headless session
    runuser -u user -- bash -c "
      export HOME=/home/user
      export XDG_RUNTIME_DIR=/run/user/1000
      export DBUS_SESSION_BUS_ADDRESS=unix:path=/run/user/1000/bus
      echo 'Creating and unlocking default keyring...'
      printf '\n' | gnome-keyring-daemon --unlock
    " || true

    chown -R 1000:1000 "${keyring_dir}"
  fi
}

# Sets Google Chrome as the default browser.
setup_browser() {
  if [[ ! -f /home/user/.config/mimeapps.list ]]; then
    log "Setting google-chrome as default browser..."
    mkdir -p /home/user/.config
    runuser -u user -- bash -c '
      export HOME=/home/user
      xdg-mime default google-chrome.desktop text/html
      xdg-mime default google-chrome.desktop x-scheme-handler/http
      xdg-mime default google-chrome.desktop x-scheme-handler/https
      xdg-mime default google-chrome.desktop x-scheme-handler/about
      xdg-settings set default-web-browser google-chrome.desktop
    '
    chown -R 1000:1000 /home/user/.config
  fi
}

# Configures GNOME Shell extensions and appearance.
setup_gnome_settings() {
  log "Configuring GNOME settings..."

  local ext="just-perfection-desktop@just-perfection"
  wait_for_extension "${ext}" || true

  runuser -u user -- bash <<EOF
    export XDG_RUNTIME_DIR=/run/user/1000
    export DBUS_SESSION_BUS_ADDRESS="unix:path=/run/user/1000/bus"

    # Ensure extensions are allowed
    gsettings set org.gnome.shell disable-user-extensions false

    # Enable extensions using the more robust gnome-extensions command
    if gnome-extensions list | grep -q "${ext}"; then
      echo "Enabling ${ext}..."
      gnome-extensions enable "${ext}" || true
    fi

    # Configure Just Perfection specific settings
    gsettings set org.gnome.shell.extensions.just-perfection screen-sharing-indicator false
    gsettings set org.gnome.shell.extensions.just-perfection screen-recording-indicator false
    gsettings set org.gnome.shell.extensions.just-perfection startup-status 0
    gsettings set org.gnome.shell.extensions.just-perfection support-notifier-showed-version 999

    # Aggressive privacy overrides to prevent indicator triggers
    gsettings set org.gnome.SessionManager auto-save-session true
EOF
}
main() {
  if [[ -f /run/user-setup-done ]]; then
    log "User setup already done, skipping."
    exit 0
  fi

  log "Starting user setup..."
  setup_runtime_dir
  setup_home_dir
  setup_rdp_certs
  setup_keyring
  setup_browser
  setup_gnome_settings

  log "User setup complete."
  touch /run/user-setup-done
}

if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
  main "$@"
fi

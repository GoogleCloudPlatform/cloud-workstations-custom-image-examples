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
    if runuser -u "${WORKSTATION_USER}" -- bash -c "
      export XDG_RUNTIME_DIR=/run/user/${WORKSTATION_UID}
      export DBUS_SESSION_BUS_ADDRESS=unix:path=/run/user/${WORKSTATION_UID}/bus
      gnome-extensions list" | grep -q "${ext_id}"; then
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
  local xdg_runtime_dir="/run/user/${WORKSTATION_UID}"
  log "Setting up XDG_RUNTIME_DIR at ${xdg_runtime_dir}..."
  mkdir -p "${xdg_runtime_dir}"
  chown "${WORKSTATION_UID}:${WORKSTATION_UID}" "${xdg_runtime_dir}"
  chmod 700 "${xdg_runtime_dir}"
}

# Ensures correct ownership of the user home directory.
setup_home_dir() {
  log "Fixing ownership for /home/${WORKSTATION_USER}..."

  # Ensure user is in the audio group for sound support
  usermod -aG audio "${WORKSTATION_USER}" || true

  # Clean up legacy persistent configurations that break the updated headless shell
  rm -f "/home/${WORKSTATION_USER}/.config/systemd/user/org.gnome.Shell@wayland.service.d/override.conf"
  rm -f "/home/${WORKSTATION_USER}/.local/share/applications/org.gnome.Shell.desktop"

  # Use -h to avoid dereferencing symlinks (prevents failure on broken pulse/machine-id links)
  chown -R -h "${WORKSTATION_UID}:${WORKSTATION_UID}" "/home/${WORKSTATION_USER}"
}

# Generates RDP TLS certificates if necessary.
setup_rdp_certs() {
  log "Checking RDP TLS certificates..."
  local grd_cert_dir="/home/${WORKSTATION_USER}/.local/share/gnome-remote-desktop"
  if command -v winpr-makecert3 >/dev/null 2>&1 || command -v winpr-makecert >/dev/null 2>&1; then
    if [[ ! -f "${grd_cert_dir}/rdp-tls.crt" ]]; then
      log "Generating new RDP TLS certificate..."
      mkdir -p "${grd_cert_dir}"
      chown -R "${WORKSTATION_UID}:${WORKSTATION_UID}" "/home/${WORKSTATION_USER}/.local"
      local make_cert
      make_cert=$(command -v winpr-makecert3 || command -v winpr3-makecert || command -v winpr-makecert)
      runuser -u "${WORKSTATION_USER}" -- "${make_cert}" -silent -rdp -path "${grd_cert_dir}" rdp-tls

      # Strip trailing null bytes that break some PEM parsers (e.g. FreeRDP/GRD)
      for f in "${grd_cert_dir}"/rdp-tls.*; do
        if [[ -f "$f" ]]; then
          tr -d '\0' < "$f" > "$f.tmp" && mv "$f.tmp" "$f"
        fi
      done

      chown -R "${WORKSTATION_UID}:${WORKSTATION_UID}" "${grd_cert_dir}"
    fi
  fi
}

# Initializes the GNOME keyring with a blank password.
setup_keyring() {
  if command -v gnome-keyring-daemon >/dev/null 2>&1; then
    log "Initializing gnome-keyring..."
    local keyring_dir="/home/${WORKSTATION_USER}/.local/share/keyrings"

    # The home directory is persistent, but our password is ephemeral.
    # We MUST nuke the old keyrings on boot, otherwise the old password
    # remains and the daemon cannot be unlocked, breaking RDP sync.
    rm -rf "${keyring_dir}"
    mkdir -p "${keyring_dir}"
    chown -R "${WORKSTATION_UID}:${WORKSTATION_UID}" "/home/${WORKSTATION_USER}/.local"

    # Ensure a default keyring exists and is unlocked
    # We use a blank password for the headless session
    runuser -u "${WORKSTATION_USER}" -- bash -c "
      export HOME=/home/${WORKSTATION_USER}
      export XDG_RUNTIME_DIR=/run/user/${WORKSTATION_UID}
      export DBUS_SESSION_BUS_ADDRESS=unix:path=/run/user/${WORKSTATION_UID}/bus
      echo 'Creating and unlocking default keyring...'
      printf '\n' | gnome-keyring-daemon --unlock
    " || true

    chown -R "${WORKSTATION_UID}:${WORKSTATION_UID}" "${keyring_dir}"
  fi
}

# Sets Google Chrome as the default browser.
setup_browser() {
  if [[ ! -f /home/${WORKSTATION_USER}/.config/mimeapps.list ]]; then
    log "Setting google-chrome as default browser..."
    mkdir -p "/home/${WORKSTATION_USER}/.config"
    chown -R "${WORKSTATION_UID}:${WORKSTATION_UID}" "/home/${WORKSTATION_USER}/.config"
    runuser -u "${WORKSTATION_USER}" -- bash -c "
      export HOME=/home/${WORKSTATION_USER}
      xdg-mime default google-chrome.desktop text/html
      xdg-mime default google-chrome.desktop x-scheme-handler/http
      xdg-mime default google-chrome.desktop x-scheme-handler/https
      xdg-mime default google-chrome.desktop x-scheme-handler/about
      xdg-settings set default-web-browser google-chrome.desktop
    "
  fi
}

# Configures GNOME Shell extensions and appearance.
setup_gnome_settings() {
  log "Configuring GNOME settings..."

  local ext="just-perfection-desktop@just-perfection"
  wait_for_extension "${ext}" || true

  # We pass the extension ID and user UID as positional parameters to the inner bash shell
  # to avoid repetition while maintaining a quoted HEREDOC (which prevents
  # unwanted expansion of other variables like ${favorites[@]}).
  runuser -u "${WORKSTATION_USER}" -- bash -s "${ext}" "${WORKSTATION_UID}" <<'EOF'
    ext_id="${1}"
    uid="${2}"

    export XDG_RUNTIME_DIR="/run/user/${uid}"
    export DBUS_SESSION_BUS_ADDRESS="unix:path=/run/user/${uid}/bus"

    # Ensure extensions are allowed
    gsettings set org.gnome.shell disable-user-extensions false

    # Enable extensions using the more robust gnome-extensions command
    if gnome-extensions list | grep -q "${ext_id}"; then
      echo "Enabling ${ext_id}..."
      gnome-extensions enable "${ext_id}" || true
    fi

    # Configure Just Perfection specific settings
    gsettings set org.gnome.shell.extensions.just-perfection screen-sharing-indicator false
    gsettings set org.gnome.shell.extensions.just-perfection screen-recording-indicator false
    gsettings set org.gnome.shell.extensions.just-perfection startup-status 0
    gsettings set org.gnome.shell.extensions.just-perfection support-notifier-showed-version 999

    # Aggressive privacy overrides to prevent indicator triggers
    gsettings set org.gnome.SessionManager auto-save-session true

    # Disable screen locking and idle timeouts to avoid "Screen Locking disabled" notifications
    gsettings set org.gnome.desktop.screensaver lock-enabled false
    gsettings set org.gnome.desktop.session idle-delay 0
    gsettings set org.gnome.desktop.lockdown disable-lock-screen true

    # Set favorite apps (dynamically discovered from /usr/share/applications)
    local apps_dir="/usr/share/applications"
    local ignored_patterns="mimeinfo.cache|gnome-session|system-config-|org.gnome.Terminal|org.gnome.Nautilus|google-chrome"
    
    # 1. Start with core defaults
    local favorites=("google-chrome.desktop" "org.gnome.Nautilus.desktop" "org.gnome.Terminal.desktop")
    
    # 2. Append custom layer-supplied apps
    if [[ -d "${apps_dir}" ]]; then
      while read -r app; do
        favorites+=("${app}")
      done < <(ls "${apps_dir}"/*.desktop 2>/dev/null | xargs -n1 basename | grep -vE "${ignored_patterns}")
    fi

    # 3. Format as GNOME array and apply
    local apps_array=$(printf "'%s', " "${favorites[@]}" | sed 's/, $//')
    gsettings set org.gnome.shell favorite-apps "[${apps_array}]"

    # Run additional modular user-level hooks
    if [ -d "/google/scripts/user-setup.d/" ]; then
      for hook in /google/scripts/user-setup.d/*.sh; do
        [ -e "$hook" ] || continue
        echo "Running session hook: $hook"
        # Source the hook inside the runuser bash session to inherit the environment
        source "$hook"
      done
    fi
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

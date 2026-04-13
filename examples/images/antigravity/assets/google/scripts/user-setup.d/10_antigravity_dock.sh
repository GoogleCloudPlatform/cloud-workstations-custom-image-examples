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

# Additive dock pinning for Antigravity.
# This hook is executed by the base GNOME layer's user_setup.sh script.

APP_ID="antigravity.desktop"

# Read current favorite apps
current_apps=$(gsettings get org.gnome.shell favorite-apps)

# Prepend if not already present
if [[ "$current_apps" != *"$APP_ID"* ]]; then
  echo "Adding ${APP_ID} to GNOME dock..."
  if [[ "$current_apps" == "[]" || "$current_apps" == "@as []" ]]; then
    gsettings set org.gnome.shell favorite-apps "['${APP_ID}']"
  else
    # Prepend to the list
    new_apps=$(echo "$current_apps" | sed "s/\[/['${APP_ID}', /")
    gsettings set org.gnome.shell favorite-apps "$new_apps"
  fi
fi

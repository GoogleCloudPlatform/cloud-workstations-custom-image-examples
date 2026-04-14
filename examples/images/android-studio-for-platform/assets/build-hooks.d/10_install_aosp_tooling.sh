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

install_aosp_tooling() {
  local apt_opts="${APT_OPTS:--y --no-install-recommends}"
  local aosp_pkgs="${AOSP_PKGS:-bison build-essential curl flex fontconfig git-core gnome-terminal gnupg lib32z1-dev libc6-dev-i386 libgl1-mesa-dev libncurses6 libx11-dev libxml2-utils repo rsync unzip x11proto-core-dev xsltproc zip zlib1g-dev}"

  echo "Updating apt repositories..."
  retry "apt-get update"

  echo "Installing Cuttlefish from /cuttlefish..."
  pushd /cuttlefish > /dev/null
  # shellcheck disable=SC2086
  apt-get install ${apt_opts} ./cuttlefish-base_*.deb ./cuttlefish-user_*.deb
  popd > /dev/null

  echo "Installing AOSP tooling..."
  # shellcheck disable=SC2086
  apt-get install ${apt_opts} ${aosp_pkgs}

  ln -sf /usr/lib/x86_64-linux-gnu/libncurses.so.6 /usr/lib/x86_64-linux-gnu/libncurses.so.5
  ln -sf /usr/lib/x86_64-linux-gnu/libtinfo.so.6 /usr/lib/x86_64-linux-gnu/libtinfo.so.5

  apt-get clean
  rm -rf /var/lib/apt/lists/*
}

main() {
  install_aosp_tooling
}

if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
  main "$@"
fi
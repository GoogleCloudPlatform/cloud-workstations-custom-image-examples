#!/bin/bash
#
# Copyright 2024 Google Inc. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
# Startup script to configure browse-lite (machine) extension settings.
#

if [[ "${EUID:-$(id -u)}" -eq 0 ]]; then
  exec runuser user "${BASH_SOURCE[0]}"
fi

echo "Configuring browse lite"

settings_file="/home/user/.codeoss-cloudworkstations/data/Machine/settings.json"

if [[ ! -f ${settings_file} ]]; then
  mkdir -p /home/user/.codeoss-cloudworkstations/data/Machine/
  echo "{}" > ${settings_file}
fi

if [[ ! $(grep "browse-lite.chromeExecutable" "${settings_file}") ]]; then
  jq '{"browse-lite.chromeExecutable": "/usr/bin/google-chrome"} + .' ${settings_file} > ${settings_file}.tmp
  mv ${settings_file}.tmp ${settings_file}
fi

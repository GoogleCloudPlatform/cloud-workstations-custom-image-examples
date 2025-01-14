#!/bin/bash

# Copyright 2025 Google LLC
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

if [[ "${EUID:-$(id -u)}" -eq 0 ]]; then
    exec runuser user "${BASH_SOURCE[0]}"
fi

echo "Writing Android Studio Defaults."

config_dir="${HOME}/.config/Google/AndroidStudio2024.1"

if [[ ! -d "${config_dir}/options" ]]; then
  mkdir -p ${config_dir}/options
fi

if [[ ! -f ${config_dir}/options/other.xml ]]; then
  cat >> ${config_dir}/options/other.xml <<-EOF
<application>
  <component name="PropertyService"><![CDATA[{
  "keyToString": {
    "android.sdk.path": "${HOME}/Android/Sdk"
  }
  }]]></component>
</application>
EOF
fi

if [[ ! -f ${config_dir}/options/androidStudioFirstRun.xml ]]; then
  cat >> ${config_dir}/options/androidStudioFirstRun.xml <<-EOF
<application>
  <component name="AndroidFirstRunPersistentData">
    <version>1</version>
  </component>
</application>
EOF
fi

if [[ ! -f ${config_dir}/early-access-registry.txt ]]; then
  cat >> ${config_dir}/early-access-registry.txt <<-EOF
ide.experimental.ui
true
ide.experimental.ui.inter.font
false
idea.plugins.compatible.build
EOF
fi

# Delete the lock file if it exists (previous session was not closed before the workstation was stopped).
if [[ -f ${config_dir}/.lock  ]]; then
  rm ${config_dir}/.lock
fi

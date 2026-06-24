#!/bin/bash
#
# Copyright 2026 Google Inc. All Rights Reserved.
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
# Startup script to add default user to workstation container.
#

# Some distributions include a user with uid=1000 by default, such as Ubuntu 24+.
userid=1000
userdel $(id -nu $userid)

groups=sudo,users
if [ ${CLOUD_WORKSTATIONS_CONFIG_DISABLE_SUDO:-false} == "true" ]; then
	groups="users"
fi

useradd -m user -u $userid -G $groups --shell /bin/bash > /dev/null
passwd -d user >/dev/null
echo "%sudo ALL=NOPASSWD: ALL" >> /etc/sudoers

# Relay known environment variables configured at container runtime
# to the default profile. This ensures that they are available over SSH sessions
echo "CLOUD_WORKSTATIONS_ENABLE_METADATA_CREDS_CHECK=${CLOUD_WORKSTATIONS_ENABLE_METADATA_CREDS_CHECK}" >> /etc/environment
echo "WEB_HOST=${WEB_HOST}" >> /etc/environment
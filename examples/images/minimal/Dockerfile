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

FROM us-west1-docker.pkg.dev/cloud-workstations-images/predefined/base AS workstations-base

FROM marketplace.gcr.io/google/ubuntu2404

RUN \
  apt-get update && \
  apt-get upgrade -y && \
  apt-get install -y openssh-server

  # Copy configuration scripts from base image.
COPY --from=workstations-base /etc/workstation-startup.d/020_start-sshd.sh /etc/workstation-startup.d/
COPY --from=workstations-base /etc/ssh/sshd_config /etc/ssh/
COPY --from=workstations-base /google/scripts/entrypoint.sh /google/scripts/
COPY --from=workstations-base /usr/bin/workstation-startup /usr/bin/

# Merge in files from the assets directory
COPY ./assets/. /

ENTRYPOINT ["/google/scripts/entrypoint.sh"]
# syntax=docker/dockerfile:1

# Copyright 2024 Google LLC
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

FROM us-central1-docker.pkg.dev/cloud-workstations-images/predefined/code-oss

ARG EXT_VERSION=0.3.9

# Install google-chrome
RUN wget https://dl.google.com/linux/direct/google-chrome-stable_current_amd64.deb && \
  apt-get update && \
  apt install -y ./google-chrome*.deb && \
  rm google-chrome*.deb && \
  apt-get clean && \
  rm -rf /var/lib/apt/lists/*

# Install browse-lite as a @builtin extension in Code OSS
RUN wget https://open-vsx.org/api/antfu/browse-lite/${EXT_VERSION}/file/antfu.browse-lite-${EXT_VERSION}.vsix && \
  unzip antfu.browse-lite-${EXT_VERSION}.vsix "extension/*" && \
  mv extension /opt/code-oss/extensions/antfu.browse-lite && \
  rm antfu.browse-lite-$EXT_VERSION.vsix

# Merge assets into the container.
COPY ./assets/. /

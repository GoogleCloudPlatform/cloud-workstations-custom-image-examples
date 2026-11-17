# syntax=docker/dockerfile:1

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

FROM us-central1-docker.pkg.dev/cloud-workstations-images/predefined/base

# Install repo and rsync
RUN apt-get update && apt-get install -y \
  repo \
  rsync && \
  #smoke tests
  repo version && \
  rsync --version && \
  apt-get clean && \
  rm -rf /var/lib/apt/lists/*

# Install Cuttlefish from Google Artifact Registry
RUN apt-get update && \
  DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends curl ca-certificates && \
  install -m 0755 -d /etc/apt/keyrings && \
  curl -fsSL https://us-apt.pkg.dev/doc/repo-signing-key.gpg -o /etc/apt/keyrings/android-cuttlefish-artifacts.asc && \
  chmod a+r /etc/apt/keyrings/android-cuttlefish-artifacts.asc && \
  echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/android-cuttlefish-artifacts.asc] https://us-apt.pkg.dev/projects/android-cuttlefish-artifacts android-cuttlefish main" \
    > /etc/apt/sources.list.d/android-cuttlefish-artifacts.list && \
  apt-get update && \
  DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends cuttlefish-base cuttlefish-user && \
  apt-get clean && \
  rm -rf /var/lib/apt/lists/*

# Install recommended tooling: https://source.android.com/docs/setup/start
RUN apt-get update && apt-get install -y \
    git-core \
    gnupg \
    flex \
    bison \
    build-essential \
    zip \
    curl \
    zlib1g-dev \
    libc6-dev-i386 \
    x11proto-core-dev \
    libx11-dev \
    lib32z1-dev \
    libgl1-mesa-dev \
    libxml2-utils \
    xsltproc \
    unzip \
    fontconfig && \
  apt-get clean && \
  rm -rf /var/lib/apt/lists/*

# Merge assets into the container.
COPY ./assets/. /

ENTRYPOINT ["/google/scripts/entrypoint.sh"]

# syntax=docker/dockerfile:1.4

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

# This image layers the Antigravity application onto the GNOME Workstation base.
# It leverages the blueprint's build-time hook mechanism for installation.

# The BASE_IMAGE is resolved by Skaffold during the build process as defined
# as a dependency.
ARG BASE_IMAGE
FROM ${BASE_IMAGE}
ARG DEBIAN_FRONTEND=noninteractive
ARG GCP_REGION=us-central1
ARG INSTALL_AGENT_DEVELOPMENT_KIT_PYTHON=true

# Copy the antigravity-specific hooks and assets into the image.
COPY assets/ /

# Run the build-time hooks provided by this layer to install Antigravity.
RUN export INSTALL_AGENT_DEVELOPMENT_KIT_PYTHON="${INSTALL_AGENT_DEVELOPMENT_KIT_PYTHON}" && \
    if [ -d "/build-hooks.d/" ]; then \
      for hook in /build-hooks.d/*.sh; do \
        [ -e "$hook" ] || continue; \
        echo "Running build hook: $hook"; \
        /bin/bash "$hook"; \
      done; \
      rm -rf /build-hooks.d/; \
    fi

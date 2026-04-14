# syntax=docker/dockerfile:1.4

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
# The BASE_IMAGE is resolved by Skaffold during the build process

ARG BASE_IMAGE

###############################################################
# Stage: Base Provider (to get common.sh)
###############################################################
FROM ${BASE_IMAGE} AS base-provider

###############################################################
# Stage: Cuttlefish Builder
###############################################################
FROM us-central1-docker.pkg.dev/cloud-workstations-images/predefined/code-oss AS cuttlefish-builder

ARG DEBIAN_FRONTEND=noninteractive
ARG APT_OPTS="-y --no-install-recommends"
ARG CURL_OPTS="-fsSL --retry 3"
ARG BUILDER_PKGS="config-package-dev curl debhelper-compat devscripts equivs git golang"
ARG LIBTINFO5_VERSION="6.3-2ubuntu0.1_amd64"
ARG RETRY_ATTEMPTS=3
ARG RETRY_DELAY=5
ARG CUTTLEFISH_RETRY_ATTEMPTS=8

SHELL ["/bin/bash", "-c"]

# Use common.sh retry helper from base image
COPY --from=base-provider /google/scripts/common.sh /google/scripts/common.sh
RUN . /google/scripts/common.sh && \
    export RETRY_ATTEMPTS=${RETRY_ATTEMPTS} && \
    export RETRY_DELAY=${RETRY_DELAY} && \
    retry "apt-get update" && \
    retry "apt-get install ${APT_OPTS} ${BUILDER_PKGS}" && \
    curl ${CURL_OPTS} -O "http://security.ubuntu.com/ubuntu/pool/universe/n/ncurses/libtinfo5_${LIBTINFO5_VERSION}.deb" && \
    retry "apt-get install ${APT_OPTS} ./libtinfo5_${LIBTINFO5_VERSION}.deb" && \
    retry "${CUTTLEFISH_RETRY_ATTEMPTS}" "${RETRY_DELAY}" "git clone https://github.com/google/android-cuttlefish" && \
    find /android-cuttlefish -type f -exec sed -i 's|https://git.kernel.org/pub/scm/linux/kernel/git/jaegeuk/f2fs-tools/|https://github.com/jaegeuk/f2fs-tools.git|g' {} + && \
    retry 5 "${RETRY_DELAY}" "/android-cuttlefish/tools/buildutils/build_packages.sh" && \
    mkdir out && \
    cp /android-cuttlefish/cuttlefish-base_*_*64.deb /out && \
    cp /android-cuttlefish/cuttlefish-user_*_*64.deb /out && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

###############################################################
# Stage: Main System
###############################################################
FROM ${BASE_IMAGE}
ARG ASFP_VERSION=canary
ARG INSTALL_ABFS_CLIENT=false
ARG DEBIAN_FRONTEND=noninteractive

ARG APT_OPTS="-y --no-install-recommends"
ARG CURL_OPTS="-fsSL --retry 3"
ARG AOSP_PKGS="bison build-essential curl flex fontconfig git-core gnome-terminal gnupg lib32z1-dev libc6-dev-i386 libgl1-mesa-dev libncurses6 libx11-dev libxml2-utils repo rsync unzip x11proto-core-dev xsltproc zip zlib1g-dev"
ARG ABFS_PKGS="abfs-client apt-transport-artifact-registry git psmisc rsync unzip zip"

ENV ASFP_VERSION=${ASFP_VERSION}
ENV INSTALL_ABFS_CLIENT=${INSTALL_ABFS_CLIENT}
ENV APT_OPTS=${APT_OPTS}
ENV CURL_OPTS=${CURL_OPTS}
ENV AOSP_PKGS=${AOSP_PKGS}
ENV ABFS_PKGS=${ABFS_PKGS}

# Copy cuttlefish builder artifacts so they can be installed by hooks
COPY --from=cuttlefish-builder /out /cuttlefish

# Merge assets into the container.
COPY assets/ /

# Run build hooks
RUN export ASFP_VERSION="${ASFP_VERSION}" && \
    export INSTALL_ABFS_CLIENT="${INSTALL_ABFS_CLIENT}" && \
    export APT_OPTS="${APT_OPTS}" && \
    export CURL_OPTS="${CURL_OPTS}" && \
    export AOSP_PKGS="${AOSP_PKGS}" && \
    export ABFS_PKGS="${ABFS_PKGS}" && \
    if [ -d "/build-hooks.d/" ]; then \
      for hook in /build-hooks.d/*.sh; do \
        [ -e "$hook" ] || continue; \
        echo "Running build hook: $hook"; \
        /bin/bash "$hook" || exit 1; \
      done; \
      rm -rf /build-hooks.d/; \
    fi && \
    rm -rf /cuttlefish && \
    dconf update

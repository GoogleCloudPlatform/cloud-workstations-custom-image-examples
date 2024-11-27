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

#######################################################
# NoVNC Builder Container
#######################################################
FROM us-central1-docker.pkg.dev/cloud-workstations-images/predefined/base as novnc-builder

ARG NOVNC_BRANCH=v1.5.0
ARG WEBSOCKIFY_BRANCH=v0.12.0

WORKDIR /out

RUN git clone --quiet --depth 1 --branch $NOVNC_BRANCH https://github.com/novnc/noVNC.git && \
  cd noVNC/utils && \
  git clone  --quiet --depth 1 --branch $WEBSOCKIFY_BRANCH https://github.com/novnc/websockify.git

#######################################################
# End NoVNC Builder Container
#######################################################

# Main container build
FROM us-central1-docker.pkg.dev/cloud-workstations-images/predefined/base

# Install and configure systemd. Alternatively, you can omit this command and
# instead source from the image built using ../../systemd/Dockerfile
RUN apt-get update && apt-get install -y \
  systemd && \
  apt-get clean && \
  rm -rf /var/lib/apt/lists/* &&\
  ln -s /dev/null /etc/systemd/system/apache2.service && \
  ln -s /dev/null /etc/systemd/system/getty@tty1.service && \
  ln -s /dev/null /etc/systemd/system/ldconfig.service && \
  /sbin/ldconfig -Xv && \
  ln -s /dev/null /etc/systemd/system/systemd-modules-load.service && \
  ln -s /dev/null /etc/systemd/system/ssh.socket && \
  ln -s /dev/null /etc/systemd/system/ssh.service && \
  echo "d /run/sshd 0755 root root" > /usr/lib/tmpfiles.d/sshd.conf && \
  echo -e "x /run/docker.socket - - - - -\nx /var/run/docker.socket - - - - -" > /usr/lib/tmpfiles.d/docker.conf

# Install GNOME
RUN apt-get update && apt-get install -y \
    gnome-software \
    gnome-software-common \
    gnome-software-plugin-snap \
    libappstream-glib8 \
    libgd3 \
    colord \
    gnome-control-center \
    gvfs-backends \
    hplip \
    libgphoto2-6 \
    libsane1 \
    sane-utils \
    ubuntu-desktop-minimal && \
  apt-get remove -y gnome-initial-setup && \
  apt-get remove -y --purge cloud-init && \
  apt-get clean && \
  rm -rf /var/lib/apt/lists/* && \
  chmod -x /usr/lib/ubuntu-release-upgrader/check-new-release-gtk

# Install TigerVNC and noVNC
COPY --from=novnc-builder /out/noVNC /opt/noVNC
RUN apt-get update && apt-get install -y \
    dbus-x11 \
    tigervnc-common \
    tigervnc-scraping-server \
    tigervnc-standalone-server \
    tigervnc-xorg-extension \
    python3-numpy  && \
  apt-get clean && \
  rm -rf /var/lib/apt/lists/*

# Merge assets into the container.
COPY assets/. /

# Run TigerVNC and noVNC as services.
RUN ln -s /etc/systemd/system/tigervnc.service /etc/systemd/system/multi-user.target.wants/ && \
  ln -s /etc/systemd/system/novnc.service /etc/systemd/system/multi-user.target.wants/ && \
  systemctl enable tigervnc && \
  systemctl enable novnc

# This is implicit when extending workstations predefined images, however we are
# including it in the sample to explicitly call-out we are overriding the
# default entrypoint when merging assets.
ENTRYPOINT ["/google/scripts/entrypoint.sh"]

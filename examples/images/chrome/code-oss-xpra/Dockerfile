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

# Install Xpra
RUN curl http://xpra.org/gpg.asc > /etc/apt/keyrings/xpra.asc && \
echo "deb [signed-by=/etc/apt/keyrings/xpra.asc] http://xpra.org/ noble main" > \
  /etc/apt/sources.list.d/xpra.list && \
  apt-get update && apt-get install -y \
  xpra \
  xpra-x11 \
  xpra-html5 \
  python3-requests \
  python3-paramiko \
  python3-zeroconf \
  python3-avahi && \
  apt-get clean && \
  rm -rf /var/lib/apt/lists/*

# Install google-chrome
RUN wget https://dl.google.com/linux/direct/google-chrome-stable_current_amd64.deb && \
  apt-get update && \
  apt install -y ./google-chrome*.deb && \
  rm google-chrome*.deb && \
  apt-get clean && \
  rm -rf /var/lib/apt/lists/* && \
  # Ensure chrome will work in a headless environment without gpu support by
  # specifying environment variables and flags
  echo "QT_X11_NO_MITSHM=1 _X11_NO_MITSHM=1 _MITSHM=0 $(tail -n -1 /opt/google/chrome/google-chrome) --disable-gpu --disable-dev-shm-usage --use-gl=swiftshader" >> /opt/google/chrome/google-chrome && \
  sed -i 'N;$!P;D' /opt/google/chrome/google-chrome

# Merge assets into the container.
COPY ./assets/. /

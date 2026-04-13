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

ARG NODE_VERSION=20-slim
ARG ALPINE_VERSION=3.19

###############################################################
# Stage: web-builder
###############################################################
FROM node:${NODE_VERSION} AS web-builder
WORKDIR /build/web
COPY web/package*.json ./
RUN npm install --silent
COPY docs/ ../docs/
COPY web/ ./
RUN npm run build

###############################################################
# Stage: Preflight Provider
###############################################################
FROM alpine:${ALPINE_VERSION}
COPY assets/ /export/
COPY --from=web-builder /build/web/dist/ /export/var/www/html/

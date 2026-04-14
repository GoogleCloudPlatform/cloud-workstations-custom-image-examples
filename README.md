<!--
Copyright 2024-2026 Google LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
-->

# Cloud Workstations Custom Image Examples

This repository contains examples intended to help you get started creating [customized container images](https://cloud.google.com/workstations/docs/customize-container-images) for use on [Cloud Workstations](https://cloud.google.com/workstations). The examples provided demonstrate how to extend various Cloud Workstations preconfigured base images to integrate various VDI technologies, IDEs, and other development resources and tooling.

The examples provided are intended as samples only, and are not stable across updates. We recommend forking into a stable repository before deploying.  Additionally, we recommend you follow standard [best practices](https://cloud.google.com/workstations/docs/set-up-security-best-practices) when utilizing Cloud Workstations, including [rebuilding weekly](https://cloud.google.com/workstations/docs/tutorial-automate-container-image-rebuild) to pick up security updates.

## Images

- [**android-studio**](examples/images/android/android-studio): Contains an example image that starts Android Studio running under GNOME that can be accessed by a browser-based client.
- [**android-open-source-project**](examples/images/android-open-source-project): Contains images that can be used to [build](examples/images/android-open-source-project/repo-builder/) and develop against the Android Open Source Project using [Code OSS](examples/images/android-open-source-project/code-oss) or [Android Studio for Platform](examples/images/android-open-source-project/android-studio-for-platform/).
- [**antigravity**](examples/images/antigravity): Custom Image for Cloud Workstations with Antigravity.
- [**chrome**](examples/images/chrome): Contains examples showing how to interact with an instance of chrome running on a remote workstation using either [browse-lite](examples/images/chrome/code-oss-browse-lite/) extension running in Code OSS, or [Xpra](examples/images/chrome/code-oss-xpra/).
- [**gnome**](examples/images/gnome): GNOME Blueprint for Google Cloud Workstations with Guacamole; includes Gemini CLI and Agent Development Toolkit (ADK) by default.
- [**gnome-novnc**](examples/images/gnome-novnc): Example of an image running GNOME desktop that can be accessed by a browser-based client.
- [**systemd**](examples/images/systemd): Adapts a given predefined image to use systemd instead of the default simplified init system.

## Disclaimer

This is not an officially supported Google product. This project is not eligible for the [Google Open Source Software Vulnerability Rewards Program](https://bughunters.google.com/open-source-security).

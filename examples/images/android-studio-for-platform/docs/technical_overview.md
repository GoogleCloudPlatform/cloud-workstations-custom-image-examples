<!--
Copyright 2026 Google LLC

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

# ASFP Technical Overview

This document provides a technical summary of the features and tools provided by the **Android Studio for Platform (ASFP) Custom Image for Cloud Workstations**.

## Core Concept

The ASFP Custom Image for Cloud Workstations is a specialized image layer designed for platform-level Android development. It extends the **GNOME Blueprint** with a full suite of tools for building, testing, and developing the Android Open Source Project (AOSP).

## Feature Stack

| Layer | Component | Description |
| :--- | :--- | :--- |
| **IDE** | Android Studio for Platform (ASFP) | Optimized IDE for navigating and developing the AOSP codebase. |
| **Build Tools** | AOSP Toolchain | Pre-installed packages (bison, build-essential, etc.) and `repo` tool. |
| **Virtualization** | Cuttlefish Emulator | Full support for running Android virtual devices within the workstation. |
| **Storage** | ABFS Client | Integration for the Android Build File System to speed up platform builds. |
| **Automation** | Helper Scripts | Specialized scripts for AOSP builds and Cuttlefish management. |

## Key Technologies

### 1. Specialized AOSP Tooling
The image includes a comprehensive set of packages required for the AOSP build process. This eliminates the need for developers to manually configure their environment, which can often take hours. Tools like `repo` are pre-configured and ready to use.

### 2. Cuttlefish in the Cloud
Running the Android emulator (Cuttlefish) in a cloud environment requires specific kernel support and configurations (e.g., KVM). This image pre-configures these dependencies, enabling developers to run and test Android system images directly within their browser-based workstation.

### 3. ABFS (Android Build File System)
AOSP builds are notoriously large and I/O intensive. The optional ABFS integration allows the workstation to mount the AOSP source tree as a virtual file system, significantly reducing the initial sync time and improving build performance by offloading I/O to Google's backend infrastructure.

### 4. Hook-Based Customization
Like other images in this repository, the ASFP image uses **Build Hooks** (`/build-hooks.d/`). These scripts handle the installation of ASFP, AOSP tooling, and ABFS in a modular way, making it easy to update specific components without rebuilding the entire stack.

## Use Cases

*   **Platform Engineering**: Developing and testing new features or fixes for the Android OS.
*   **AOSP Onboarding**: Providing new developers with a "one-click" environment that is fully compliant with AOSP build requirements.
*   **Automated Testing**: Using the built-in Cuttlefish support to run automated system-level tests in a reproducible environment.

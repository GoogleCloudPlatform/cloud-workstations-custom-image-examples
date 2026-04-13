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

# Antigravity Technical Overview

This document provides a technical summary of the features and tools provided by the **Antigravity Custom Image for Cloud Workstations**.

## Core Concept

The Antigravity Custom Image for Cloud Workstations is a specialized image layer designed for high-performance cloud-native development and hackathon scenarios. It extends the **GNOME Blueprint** with a curated set of tools and a pre-configured desktop experience.

## Feature Stack

| Layer | Component | Description |
| :--- | :--- | :--- |
| **User Interface** | Antigravity Desktop Shell | A customized launcher and dashboard for hackathon participants. |
| **Development** | Agent Development Kit (ADK) | Python-based tools for building and testing AI agents. |
| **AI Integration** | Gemini CLI | Direct access to Gemini Pro and Flash models for terminal automation. |
| **Systems** | Google Artifact Registry | Private repository integration for secure package delivery. |

## Key Technologies

### 1. Layered Image Architecture
Antigravity uses a "Thin Layer" approach. It does not replace the OS or desktop environment but instead "patches" them at build-time. This ensures that any security updates applied to the base GNOME image are automatically inherited by Antigravity without requiring manual reconfiguration.

### 2. Region-Aware Package Management
To optimize build performance and ensure compliance with data residency requirements, the Antigravity build-time hooks dynamically adjust APT sources based on the `${GCP_REGION}` argument. This ensures that the workstation pulls assets from the closest Google Cloud data center.

### 3. Automated Setup Hooks
Unlike traditional images that require complex startup scripts, Antigravity uses **Build Hooks**. These are one-time execution scripts that run during the container creation phase, allowing for:
*   Pre-compilation of application assets.
*   System-wide configuration of developer tools.
*   Registration of background services.

## Use Cases

*   **Hackathons**: Rapidly deploy 100+ identical, high-performance desktop environments with all required tools pre-installed.
*   **AI Research**: Provide researchers with a pre-configured environment containing the ADK and Gemini CLI for immediate productivity.
*   **Corporate Standard Images**: Use Antigravity as a template for building internal developer "golden images" that include proprietary SDKs and configurations.

---
Brought to you by the Google Cloud Professional Services Organization Apps team (PSO AppΣ)

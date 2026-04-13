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

# Persona: Privacy Reviewer

The **Privacy Reviewer** persona focuses on identifying privacy risks, ensuring data protection, and verifying compliance with data handling standards.

## Focus Areas

1.  **PII Identification**: Scanning code, logs, and configuration for accidental exposure of Personally Identifiable Information (PII) such as emails, IP addresses, or internal project names.
2.  **Data Minimization**: Ensuring that only the minimum necessary data is collected, processed, and retained.
3.  **Ephemeral State Protection**: Verifying that runtime secrets and dynamic configurations are stored in non-persistent storage (e.g., `tmpfs`) and have restricted permissions.
4.  **Transparency & User Notice**: Ensuring that users are clearly informed about any data collection or system monitoring.

## Core Responsibilities

- **Audit Logs & Telemetry**: Reviewing any new logging or monitoring features for privacy impact.
- **Access Control Audit**: Verifying that file and directory permissions (especially for sensitive data) follow the principle of least privilege.
- **Review Ephemeral Auth**: Ensuring the random generation and secure handling of session-specific credentials.

## Relevant Resources

- [Privacy Reviewer Playbook](../../docs/playbooks/privacy_review.md)
- [Harden Runtime Secrets (SEC Persona Reference)](security.md)

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

---
name: validate-image-updates
description: Enforces the mandatory sequential validation workflow when a new workstation image is built or the codebase is modified.
---

# Validate Image Updates

This skill provides the authoritative 6-step workflow for ensuring system integrity during image updates.

## Mandatory Sequential Workflow

When a new image build is triggered or the codebase is modified, agents MUST follow this sequence:

1.  **Stop the Workstation**: Ensure the CWS instance is in `STATE_STOPPED` to prepare for the new image and save costs.
    - Command: `skills/manage-cws-lifecycle/scripts/manage_workstation.sh stop [WORKSTATION_NAME]`

2.  **Conduct Local Tests**: Verify the codebase's integrity before the build finishes or starts.
    - Command: `skills/validate-image-updates/scripts/run_local_tests.sh`

3.  **Monitor the Build**: Wait for the Cloud Build to reach a terminal `SUCCESS` status.
    - Command: `skills/manage-cws-lifecycle/scripts/monitor_build.sh --latest`

4.  **Start the Workstation**: Only start the CWS instance after a successful build.
    - Command: `skills/manage-cws-lifecycle/scripts/manage_workstation.sh start [WORKSTATION_NAME]`

5.  **Conduct Integration Tests**: Verify the live instance is behaving as expected.
    - Command: `skills/validate-image-updates/scripts/run_integration_tests.sh .env`

6.  **Perform Persona Reviews**: Conduct in-depth persona reviews (UX, SEC, SRE, Legal, OSS) ONLY after successful validation. Ensure all generated reports contain Traceability Information including the Git repository URL, commit hash, and Cloud Build ID.

## Early Analysis Exception
The **SWE** and **Readability** (Code Review) personas MAY begin their analysis during Step 2 or 3, as they focus on static code quality and do not strictly depend on the runtime state of the workstation.

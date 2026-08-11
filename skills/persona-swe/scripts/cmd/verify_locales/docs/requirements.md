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

# Technical Requirements: Verify Locales

## 1. Functional Requirements (The "What")

- **FR1: Alphabetical Key Sorting**: The tool MUST verify that all keys in i18n locale JSON files are sorted alphabetically.
- **FR2: Key Consistency**: The tool MUST ensure that all locale files in a directory contain the exact same set of keys as the primary `en.json` file.
- **FR3: Format Support**: The tool MUST support both JSON Array (list of objects with `id` and `translation`) and JSON Map (key-value pairs) formats.
- **FR4: Missing Key Detection**: The tool MUST identify and report keys present in `en.json` but missing in other locale files.
- **FR5: Extra Key Detection**: The tool MUST identify and report keys present in other locale files but missing in `en.json`.
- **FR6: Mandatory Trailing Newline**: The tool MUST verify that all locale JSON files end with a mandatory newline character to ensure POSIX compliance and clean git diffs.

## 2. Non-Functional Requirements (The "How Well")

- **NFR1: Strict Order Verification**: The tool MUST use a non-buffering or order-preserving parser to ensure that keys are physically ordered correctly in the file, not just logically present.
- **NFR2: Informative Error Reporting**: The tool MUST provide specific error messages indicating the file and the specific key or ordering issue discovered.
- **NFR3: Go Readability**: The implementation MUST adhere to Google Go Readability standards and pass `golangci-lint`.
- **NFR4: Fast Execution**: The tool MUST be able to verify hundreds of locale files across multiple directories in under 1 second.

## 3. Compliance & Legal

- **CL1: I18n Style Guide Adherence**: The tool MUST enforce the rules defined in the [i18n Style Guide](../../../../../docs/style_guides/i18n.md).
- **CL2: Pre-commit Integration**: The tool MUST be compatible with `pre-commit` for automated validation during the development lifecycle.

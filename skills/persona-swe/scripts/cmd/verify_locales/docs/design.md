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

# Design Document: Verify Locales

## 1. Introduction

`verify_locales` is a synchronization and validation utility used to maintain the integrity of internationalization (i18n) assets. In a globalized project, ensuring that all supported languages have consistent and well-formatted translation files is critical for user experience and automated processing.

## 2. Architecture

The tool is divided into a thin CLI wrapper and a reusable internal library:

- **`cmd/verify_locales`**: Handles command-line arguments and directory discovery.
- **`internal/locales`**: Contains the core logic for parsing JSON files, extracting keys while preserving order, and performing cross-file comparisons.

## 3. Implementation Details

### 3.1 Order-Preserving JSON Parsing

Standard JSON decoders (like Go's `json.Unmarshal`) into a map do not preserve the order of keys. To enforce the **Alphabetical Key Sorting (FR1)** requirement, `verify_locales` utilizes `json.Decoder.Token()` to manually walk the JSON structure. This allows it to record the sequence of keys as they appear in the physical file.

### 3.2 Consistency Engine

The tool follows a "Primary-Secondary" model:

1. **Primary Analysis**: `en.json` is analyzed first to establish the "source of truth" for keys.
2. **Secondary Validation**: All other `.json` files in the same directory are compared against the primary key set.
3. **Set Comparison**: It performs a symmetric difference check to find both missing and extra keys.

### 3.3 Multi-Format Handling

The tool supports both the Array format (used by some i18n libraries for metadata support) and the standard Map format. It dynamically detects the format by inspecting the first token of the JSON file (`[` vs `{`).

## 4. Integration

The tool is designed primarily as a **Pre-commit Hook (CL2)**. It is invoked automatically during `git commit` to ensure that no developer accidentally introduces an unsorted or incomplete locale file into the repository.

## 5. Performance Considerations

By using a streaming `json.Decoder` and performing set-based comparisons, the tool maintains $O(N)$ performance where $N$ is the total number of keys across all files, ensuring it remains fast even for large translation sets.

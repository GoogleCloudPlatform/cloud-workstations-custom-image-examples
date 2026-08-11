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

# Verify Locales

A validation utility to ensure that i18n locale JSON files are consistent, properly sorted, and adhere to project standards.

## Features

- **Sorting Validation**: Ensures all keys are alphabetically sorted for clean diffs.
- **Consistency Check**: Verifies that all translations match the keys defined in the primary `en.json`.
- **Format Agnostic**: Supports both JSON Map and JSON Array formats.
- **POSIX Compliance**: Enforces trailing newlines in all JSON files.

## Usage

```bash
go run skills/persona-swe/scripts/cmd/verify_locales/main.go <paths...>
```

The tool accepts directories or individual files. When a directory is provided, it validates all `.json` files within that directory against the local `en.json`.

### Example

```bash
go run skills/persona-swe/scripts/cmd/verify_locales/main.go examples/preflight-web/public/locales/
```

## Integration

This tool is integrated into the project's `.pre-commit-config.yaml`:

```yaml
- id: verify-locales
  name: Verify i18n Locales
  entry: go run skills/persona-swe/scripts/cmd/verify_locales/main.go
  language: system
  files: ".*locale.*\\.json$"
```

## Documentation

- [Technical Requirements](./docs/requirements.md)
- [Design Document](./docs/design.md)

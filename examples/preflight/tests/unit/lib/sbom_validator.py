#!/usr/bin/env python3

# Copyright 2026 Google LLC
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

"""Validation utility for the Software Bill of Materials and its assets."""

import json
import os
import sys


def validate_manifest(sbom_path: str):
    """Checks if manifest exists, is valid JSON, and has required metadata."""
    try:
        with open(sbom_path, "r") as f:
            data = json.load(f)
            meta = data.get("metadata", {})
            required = ["project", "copyright", "organization", "repository", "url"]
            for field in required:
                if field not in meta:
                    print(f"Missing metadata field: {field}", file=sys.stderr)
                    sys.exit(1)
    except Exception as e:
        print(f"Manifest validation failed: {e}", file=sys.stderr)
        sys.exit(1)


def validate_assets(sbom_path: str, assets_root: str):
    """Checks if all unique licenses in SBOM have valid local files."""
    signatures = {
        "Apache-2.0": "Apache License",
        "GPL-2.0-only": "GNU GENERAL PUBLIC LICENSE",
        "BSD-2-Clause": "Redistributions in binary form",
        "MIT": "Permission is hereby granted",
    }

    try:
        with open(sbom_path, "r") as f:
            data = json.load(f)
            licenses = data.get("licenses", {})

            for lid, info in licenses.items():
                local_path = info["local_text"].lstrip("/")
                path = os.path.join(assets_root, local_path)

                if not os.path.isfile(path):
                    print(f"Missing asset for {lid}: {path}", file=sys.stderr)
                    sys.exit(1)

                if os.path.getsize(path) == 0:
                    print(f"Empty asset for {lid}: {path}", file=sys.stderr)
                    sys.exit(1)

                with open(path, "r") as tf:
                    content = tf.read().lower()

                    # 404 Check
                    if "404" in content and "not found" in content:
                        print(f"Asset for {lid} is a 404 page: {path}", file=sys.stderr)
                        sys.exit(1)

                    # Signature Check
                    sig = signatures.get(lid, "").lower()
                    if sig and sig not in content:
                        print(f"Asset for {lid} missing signature '{sig}': {path}", file=sys.stderr)
                        sys.exit(1)

            # Integrity Check: Ensure components reference valid IDs
            for comp in data.get("components", []):
                if comp.get("license_id") not in licenses:
                    print(f"Component {comp['name']} unknown license: {comp['license_id']}", file=sys.stderr)
                    sys.exit(1)

    except Exception as e:
        print(f"Asset validation failed: {e}", file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    if len(sys.argv) < 3:
        print("Usage: sbom_validator.py <action:manifest|assets> <sbom_path> [assets_root]", file=sys.stderr)
        sys.exit(1)

    action = sys.argv[1]
    path = sys.argv[2]

    if action == "manifest":
        validate_manifest(path)
    elif action == "assets":
        root = sys.argv[3] if len(sys.argv) > 3 else os.path.dirname(path)
        validate_assets(path, root)
    else:
        print(f"Unknown action: {action}", file=sys.stderr)
        sys.exit(1)

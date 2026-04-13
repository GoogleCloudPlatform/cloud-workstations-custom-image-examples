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

"""Helper utility for parsing the Software Bill of Materials manifest."""

import json
import sys
import os

def get_license_mappings(sbom_path: str):
    """Extracts unique license URLs and their local target paths.

    Returns:
        Prints space-separated strings: "lid url local_text"
    """
    if not os.path.exists(sbom_path):
        print(f"Error: SBOM not found at {sbom_path}", file=sys.stderr)
        sys.exit(1)

    try:
        with open(sbom_path, "r") as f:
            data = json.load(f)
            licenses = data.get("licenses", {})
            for lid, info in licenses.items():
                url = info.get('url')
                local_text = info.get('local_text')
                if url and local_text:
                    # We print space separated values for the bash script to consume
                    print(f"{lid} {url} {local_text}")
    except Exception as e:
        print(f"Error parsing SBOM: {e}", file=sys.stderr)
        sys.exit(1)

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Usage: sbom_helper.py <sbom_json_path>", file=sys.stderr)
        sys.exit(1)
    get_license_mappings(sys.argv[1])

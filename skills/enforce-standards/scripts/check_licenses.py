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

import argparse
import datetime
import logging
import os
import re
import subprocess
import sys

# Configure logging to only print messages without level prefixes
logging.basicConfig(level=logging.INFO, format="%(message)s")

LICENSES = {
    "Apache-2.0": """\
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.""",
    "MIT": """\
Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.""",
    "GPL-2.0-only": """\
This program is free software; you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation; version 2 of the License.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program; if not, write to the Free Software
Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA  02110-1301, USA."""
}

def get_tracked_files(filter_regex):
    try:
        result = subprocess.run(['git', 'ls-files'], stdout=subprocess.PIPE, text=True, check=True)
        files = result.stdout.splitlines()
        filter_pattern = re.compile(filter_regex, re.IGNORECASE)
        return [f for f in files if filter_pattern.search(f)]
    except subprocess.CalledProcessError as e:
        logging.error(f"Error running git ls-files: {e}")
        sys.exit(1)

def process_file_content(content, ext, current_year, holder, target_license):
    """Processes the file content to ensure license compliance."""
    original_content = content
    license_added = False
    year_updated = False
    different_holder = None

    # 0. Extract start year if it exists for the target holder before any deletions
    start_year = current_year
    year_pattern = re.compile(rf"Copyright ([0-9]{{4}})(?:-[0-9]{{4}})? {re.escape(holder)}", re.IGNORECASE)
    year_match = year_pattern.search(original_content)
    if year_match:
        start_year = int(year_match.group(1))

    # Check for different copyright holders (for warnings)
    holder_pattern = re.compile(r"Copyright [0-9]{4}(?:-[0-9]{4})? (.*?)(?:\n|$)", re.IGNORECASE)
    for match in holder_pattern.finditer(original_content):
        found_holder = match.group(1).strip()
        if holder.lower() not in found_holder.lower():
            different_holder = found_holder

    # 1. Remove all redundant short headers (if any)
    content = re.sub(r'/\*\*\s*\n\s*\* Copyright.*?\n\s*\*/(?:\s*\n)?', '', content, flags=re.DOTALL)
    content = re.sub(r'<!--\s*\n\s*Copyright.*?\n\s*-->(?:\s*\n)?', '', content, flags=re.DOTALL)

    # 2. Check for full license string
    if target_license in LICENSES and LICENSES[target_license].splitlines()[0] not in content:
        license_added = True
        effective_year = f"{start_year}-{current_year}" if start_year < current_year else str(current_year)
        copyright_line = f"Copyright {effective_year} {holder}\n\n"
        full_header_text = copyright_line + LICENSES[target_license]

        if ext in ['sh', 'bash', 'bats', 'py', 'desktop']:
            formatted_header = '\n'.join([f"# {line}" if line else "#" for line in full_header_text.splitlines()]) + '\n\n'
            if content.startswith('#!'):
                lines = content.splitlines(True)
                content = lines[0] + '\n' + formatted_header + ''.join(lines[1:])
            else:
                content = formatted_header + content
        elif ext in ['html', 'md']:
            formatted_header = f"<!--\n{full_header_text}\n-->\n\n"
            content = formatted_header + content
        elif ext in ['css', 'ts', 'js']:
            formatted_header = "/**\n" + '\n'.join([f" * {line}" if line else " *" for line in full_header_text.splitlines()]) + "\n */\n\n"
            content = formatted_header + content

    # 3. Update copyright year to range if in the past
    pattern = re.compile(rf"Copyright ([0-9]{{4}})(?:-[0-9]{{4}})? {re.escape(holder)}", re.IGNORECASE)
    def year_replacer(match):
        nonlocal year_updated
        start_year = int(match.group(1))
        if start_year < current_year:
            year_updated = True
            return f"Copyright {start_year}-{current_year} {holder}"
        return match.group(0)

    content = pattern.sub(year_replacer, content)

    # 4. Remove extra leading blank lines
    content = re.sub(r'^\n+', '', content)

    # 5. Enforce blank line spacing rules even if license already existed
    end_markers = [
        r'limitations under the License\.',
        r'DEALINGS IN THE SOFTWARE\.',
        r'02110-1301, USA\.'
    ]
    end_pattern = r'(' + r'|'.join(end_markers) + r')'

    if ext in ['sh', 'bash', 'bats', 'py', 'desktop']:
        content = re.sub(r'^(#!.*?\n)\n*(# Copyright)', r'\1\n\2', content)
        content = re.sub(rf'(# {end_pattern})\n+(?!$)', r'\1\n\n', content)
    elif ext in ['html', 'md']:
        content = re.sub(rf'({end_pattern}\n-->)\n+(?!$)', r'\1\n\n', content)
    elif ext in ['css', 'ts', 'js']:
        content = re.sub(rf'({end_pattern}\n \*/)\n+(?!$)', r'\1\n\n', content)

    return content, license_added, year_updated, different_holder

def main():
    parser = argparse.ArgumentParser(description="Check and apply license headers.")
    parser.add_argument("--holder", default="Google LLC", help="Copyright holder name.")
    parser.add_argument("--license", default="Apache-2.0", help="SPDX license identifier to enforce.")
    parser.add_argument("--filter", default=r"\.(ts|js|sh|css|html|md|bash|bats|py|desktop)$", help="Regex to filter files.")
    parser.add_argument("--exclude", default="node_modules/ dist/ licenses/ check_licenses.py", help="Space-separated list of strings to exclude.")
    args = parser.parse_args()

    exclude_list = args.exclude.split()
    current_year = datetime.datetime.now().year

    files_to_check = get_tracked_files(args.filter)

    modified_files = []

    for file_path in files_to_check:
        if any(ex in file_path for ex in exclude_list):
            continue

        if not os.path.exists(file_path) or os.path.islink(file_path):
            continue

        with open(file_path, 'r') as f:
            original_content = f.read()

        ext = file_path.split('.')[-1].lower()
        content, license_added, year_updated, different_holder = process_file_content(
            original_content, ext, current_year, args.holder, args.license
        )

        if content != original_content:
            with open(file_path, 'w') as f:
                f.write(content)
            modified_files.append(file_path)
            if license_added:
                logging.info(f"  - Applying {args.license} license to {file_path}")
            elif year_updated:
                logging.info(f"  - Updating copyright year in {file_path}")
            else:
                logging.info(f"  - Fixing license formatting in {file_path}")

        if different_holder:
            logging.warning(f"  - Warning: Different copyright holder ('{different_holder}') found in {file_path}")

    if modified_files:
        logging.info(f"⚠️  Applied license headers or fixes to {len(modified_files)} files.")
        sys.exit(1)
    else:
        logging.info("✅ All tracked files already have valid license headers and formatting.")
        sys.exit(0)

if __name__ == "__main__":
    main()
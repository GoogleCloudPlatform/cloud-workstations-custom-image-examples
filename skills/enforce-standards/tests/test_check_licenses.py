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

import sys
import os
import unittest

# Ensure the scripts directory is in the path to import check_licenses
SCRIPT_DIR = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
sys.path.append(os.path.join(SCRIPT_DIR, "scripts"))

import check_licenses

class TestCheckLicenses(unittest.TestCase):
    def test_adds_license_to_empty_file(self):
        content, added, updated, diff_holder = check_licenses.process_file_content(
            "", "py", 2026, "Google LLC", "Apache-2.0"
        )
        self.assertTrue(added)
        self.assertFalse(updated)
        self.assertIsNone(diff_holder)
        self.assertIn("Copyright 2026 Google LLC", content)
        self.assertIn('Licensed under the Apache License, Version 2.0 (the "License");', content)

    def test_updates_past_year(self):
        initial = '# Copyright 2024 Google LLC\n#\n# Licensed under the Apache License, Version 2.0 (the "License");'
        content, added, updated, diff_holder = check_licenses.process_file_content(
            initial, "py", 2026, "Google LLC", "Apache-2.0"
        )
        self.assertFalse(added)
        self.assertTrue(updated)
        self.assertIsNone(diff_holder)
        self.assertIn("Copyright 2024-2026 Google LLC", content)

    def test_updates_past_year_range(self):
        initial = '# Copyright 2022-2024 Google LLC\n#\n# Licensed under the Apache License, Version 2.0 (the "License");'
        content, added, updated, diff_holder = check_licenses.process_file_content(
            initial, "py", 2026, "Google LLC", "Apache-2.0"
        )
        self.assertFalse(added)
        self.assertTrue(updated)
        self.assertIsNone(diff_holder)
        self.assertIn("Copyright 2022-2026 Google LLC", content)

    def test_warns_on_different_holder(self):
        initial = "# Copyright 2024 Google Inc. All Rights Reserved.\n"
        content, added, updated, diff_holder = check_licenses.process_file_content(
            initial, "py", 2026, "Google LLC", "Apache-2.0"
        )
        self.assertTrue(added) # Since it didn't find the exact license text, it might add it depending on the setup
        self.assertFalse(updated)
        self.assertEqual("Google Inc. All Rights Reserved.", diff_holder)

    def test_ignores_current_year(self):
        initial = '# Copyright 2026 Google LLC\n#\n# Licensed under the Apache License, Version 2.0 (the "License");'
        content, added, updated, diff_holder = check_licenses.process_file_content(
            initial, "py", 2026, "Google LLC", "Apache-2.0"
        )
        self.assertFalse(added)
        self.assertFalse(updated)
        self.assertIsNone(diff_holder)
        self.assertEqual(initial, content)

    def test_preserves_old_year_on_new_license_addition(self):
        # File has copyright but no license body
        initial = "# Copyright 2024 Google LLC\n\nprint('hello')"
        content, added, updated, diff_holder = check_licenses.process_file_content(
            initial, "py", 2026, "Google LLC", "Apache-2.0"
        )
        self.assertTrue(added)
        self.assertIn("Copyright 2024-2026 Google LLC", content)

    def test_preserves_old_year_on_header_replacement_md(self):
        # Markdown block will be removed and re-added
        initial = "<!--\nCopyright 2024 Google LLC\n\nLicensed under the Apache License, Version 2.0 (the \"License\");\n... \n-->\n\n# Body"
        content, added, updated, diff_holder = check_licenses.process_file_content(
            initial, "md", 2026, "Google LLC", "Apache-2.0"
        )
        self.assertTrue(added) # Re-added because '...' doesn't match full license
        self.assertIn("Copyright 2024-2026 Google LLC", content)

    def test_preserves_old_year_range_on_header_replacement(self):
        # Existing range should be preserved as start-current
        initial = "/**\n * Copyright 2022-2024 Google LLC\n */\n\nexport const x = 1;"
        content, added, updated, diff_holder = check_licenses.process_file_content(
            initial, "ts", 2026, "Google LLC", "Apache-2.0"
        )
        self.assertTrue(added) # Short header removed and full license added
        self.assertIn("Copyright 2022-2026 Google LLC", content)

if __name__ == "__main__":
    unittest.main()

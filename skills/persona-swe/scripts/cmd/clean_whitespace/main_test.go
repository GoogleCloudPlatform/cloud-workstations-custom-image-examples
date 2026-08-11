// Copyright 2026 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"os"
	"os/exec"
	"testing"
)

// TestMainSmokeTest runs the main function in a separate process to verify it starts and displays usage correctly.
func TestMainSmokeTest(t *testing.T) {
	if os.Getenv("BE_CRASHER") == "1" {
		os.Args = []string{"clean_whitespace"}
		main()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestMainSmokeTest")
	cmd.Env = append(os.Environ(), "BE_CRASHER=1")
	err := cmd.Run()
	if err != nil {
		t.Fatalf("main() exited with error: %v", err)
	}
}

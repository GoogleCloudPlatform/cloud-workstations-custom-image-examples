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

// Package models defines shared data structures and utilities for GitSeep.
package models

import "errors"

// Standard Pipeline Control Errors
var (
	// ErrPipelineAborted signals that the pipeline should stop immediately and clean up,
	// typically due to user action (e.g., no migrations approved) or a non-fatal short-circuit.
	ErrPipelineAborted = errors.New("STOP_PIPELINE")

	// ErrConflictDetected signals that a mathematical conflict was predicted in the stratigraphy.
	ErrConflictDetected = errors.New("stratigraphy conflict detected")

	// ErrValidationFailed signals that geological validation (pre-commit) failed for one or more strata.
	ErrValidationFailed = errors.New("geological validation failed")

	// ErrInitializationFailed signals a failure to resolve the repository context or seepage rules.
	ErrInitializationFailed = errors.New("initialization failed")
)

// RichPipelineError is an error that carries a full PipelineEvent for detailed reporting.
type RichPipelineError struct {
	Event PipelineEvent
	Err   error
}

var _ error = (*RichPipelineError)(nil)

func (e *RichPipelineError) Error() string {
	if e.Event.Message != "" {
		return e.Event.Message
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	return "pipeline error"
}

// Unwrap returns the underlying error.
func (e *RichPipelineError) Unwrap() error {
	return e.Err
}

// IsAborted returns true if the error is a pipeline abort signal.
func IsAborted(err error) bool {
	return errors.Is(err, ErrPipelineAborted)
}

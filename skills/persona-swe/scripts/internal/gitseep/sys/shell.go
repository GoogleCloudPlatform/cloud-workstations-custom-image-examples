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

// Package sys provides system-level abstractions for GitSeep.
package sys

import (
	"io"
	"os/exec"
)

// ShellService defines the interface for executing system commands.
type ShellService interface {
	Command(name string, arg ...string) ShellCommand
}

// ShellCommand defines the subset of exec.Cmd methods needed by GitSeep.
type ShellCommand interface {
	SetDir(dir string)
	SetEnv(env []string)
	SetStdout(w io.Writer)
	SetStderr(w io.Writer)
	CombinedOutput() ([]byte, error)
	Run() error
	Start() error
	Wait() error
	StdoutPipe() (io.ReadCloser, error)
}

// RealShellService implements ShellService using the standard os/exec package.
type RealShellService struct{}

func (s *RealShellService) Command(name string, arg ...string) ShellCommand {
	return &realShellCommand{cmd: exec.Command(name, arg...)}
}

type realShellCommand struct {
	cmd *exec.Cmd
}

func (c *realShellCommand) SetDir(dir string)                  { c.cmd.Dir = dir }
func (c *realShellCommand) SetEnv(env []string)                { c.cmd.Env = env }
func (c *realShellCommand) SetStdout(w io.Writer)              { c.cmd.Stdout = w }
func (c *realShellCommand) SetStderr(w io.Writer)              { c.cmd.Stderr = w }
func (c *realShellCommand) CombinedOutput() ([]byte, error)    { return c.cmd.CombinedOutput() }
func (c *realShellCommand) Run() error                         { return c.cmd.Run() }
func (c *realShellCommand) Start() error                       { return c.cmd.Start() }
func (c *realShellCommand) Wait() error                        { return c.cmd.Wait() }
func (c *realShellCommand) StdoutPipe() (io.ReadCloser, error) { return c.cmd.StdoutPipe() }

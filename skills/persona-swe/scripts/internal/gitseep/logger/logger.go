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

// Package logger provides a standardized, colorful logging utility for GitSeep.
package logger

import (
	"fmt"
	"hash/fnv"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/i18n"
)

// QuietMode suppresses all non-error output when true.
var QuietMode bool

var (
	// StyleBold renders text in bold.
	StyleBold = lipgloss.NewStyle().Bold(true)
	// StyleGrey renders text in grey.
	StyleGrey = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	// StyleRed renders text in red.
	StyleRed = lipgloss.NewStyle().Foreground(lipgloss.Color("31"))
	// StyleGreen renders text in green.
	StyleGreen = lipgloss.NewStyle().Foreground(lipgloss.Color("32"))
	// StyleYellow renders text in yellow.
	StyleYellow = lipgloss.NewStyle().Foreground(lipgloss.Color("33"))
	// StylePath renders file paths in cyan for high visibility.
	StylePath = lipgloss.NewStyle().Foreground(lipgloss.Color("6")).Bold(true)
	// StyleAnchor renders bedrock anchors in a muted, low-contrast color.
	StyleAnchor = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "251", Dark: "237"})
)

const (
	// keep-sorted start
	// IconBedrock represents a bedrock commit.
	IconBedrock = "🪨"
	// IconBranch represents a branch creation/sync action.
	IconBranch = "🏞️"
	// IconExcluded represents an excluded item.
	IconExcluded = "❌"
	// IconExecution represents the reconstruction of history.
	IconExecution = "⚙️"
	// IconFinalize represents the finalization of references.
	IconFinalize = "🏆"
	// IconLab represents a validation or test action.
	IconLab = "🧪"
	// IconLithify represents a lithification action.
	IconLithify = "💎"
	// IconPercolate represents a percolate down action.
	IconPercolate = "💧"
	// IconReview represents an interactive review phase.
	IconReview = "✋"
	// IconSearch represents a search action.
	IconSearch = "🔍"
	// IconSeep represents a seep up action.
	IconSeep = "🫧"
	// IconSelected represents a selected item.
	IconSelected = "✅"
	// IconSuccess represents a successful action.
	IconSuccess = "🏆"
	// IconSync represents a synchronization or amendment action.
	IconSync = "🔄"
	// IconWarn represents a warning.
	IconWarn = "⚠️"
	// keep-sorted end
)

// LoggerOptions defines configuration for the global logger.
type LoggerOptions struct {
	// Quiet suppresses all non-error output.
	Quiet bool
}

// Init initializes the logger configuration.
func Init(opts LoggerOptions) {
	QuietMode = opts.Quiet
}

// PhasedLogger provides contextual logging prefixed with a specific phase icon and name.
type PhasedLogger struct {
	Icon  string
	Phase string
}

// WithPhase creates a new scoped logger for a specific geological phase.
func WithPhase(icon, phase string) *PhasedLogger {
	return &PhasedLogger{Icon: icon, Phase: phase}
}

func (l *PhasedLogger) format(msg string) string {
	return fmt.Sprintf("[%s %s] %s", l.Icon, l.Phase, msg)
}

// Info logs a contextual message.
func (l *PhasedLogger) Info(format string, a ...interface{}) {
	Info("%s", StyleGrey.Render(l.format(fmt.Sprintf(format, a...))))
}

// Success logs a contextual success message with the icon after the prefix.
func (l *PhasedLogger) Success(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	prefix := StyleGrey.Render(fmt.Sprintf("[%s %s] ", l.Icon, l.Phase))
	fullMsg := fmt.Sprintf("%s%s %s", prefix, IconSuccess, msg)
	if LogHook != nil {
		LogHook(fullMsg)
	}
	if !QuietMode {
		fmt.Fprintln(os.Stdout, fullMsg)
	}
}

// Warn logs a contextual warning message with the icon after the prefix.
func (l *PhasedLogger) Warn(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	prefix := StyleGrey.Render(fmt.Sprintf("[%s %s] ", l.Icon, l.Phase))
	fullMsg := fmt.Sprintf("%s%s %s %s", prefix, IconWarn, StyleBold.Render(i18n.T("prefix_warning")), StyleGrey.Render(msg))
	if LogHook != nil {
		LogHook(fullMsg)
	}
	if !QuietMode {
		fmt.Fprintln(os.Stdout, fullMsg)
	}
}

// Error logs a contextual error message with the prefix first.
func (l *PhasedLogger) Error(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	prefix := StyleGrey.Render(fmt.Sprintf("[%s %s] ", l.Icon, l.Phase))
	fullMsg := prefix + StyleRed.Render(i18n.T("prefix_error")+" "+msg)
	if LogHook != nil {
		LogHook(fullMsg)
	}
	if !QuietMode {
		fmt.Fprintln(os.Stderr, fullMsg)
	}
}

// LogHook is an optional function called for every log message.
// It can be used to broadcast logs to a TUI or other external listeners.
var LogHook func(msg string)

// Info logs a message to stdout if QuietMode is false.
func Info(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	if LogHook != nil {
		LogHook(msg)
	}
	if !QuietMode {
		fmt.Fprintln(os.Stdout, msg)
	}
}

// Warn logs a warning message to stdout with an icon if QuietMode is false.
func Warn(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fullMsg := fmt.Sprintf("%s %s %s", IconWarn, StyleBold.Render(i18n.T("prefix_warning")), msg)
	if LogHook != nil {
		LogHook(fullMsg)
	}
	if !QuietMode {
		fmt.Fprintln(os.Stdout, fullMsg)
	}
}

// Error logs an error message to stderr.
func Error(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fullMsg := StyleRed.Render(i18n.T("prefix_error") + " " + msg)
	if LogHook != nil {
		LogHook(fullMsg)
	}
	if !QuietMode {
		fmt.Fprintln(os.Stderr, fullMsg)
	}
}

// Success logs a success message to stdout with an icon if QuietMode is false.
func Success(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fullMsg := fmt.Sprintf("%s %s", IconSuccess, msg)
	if LogHook != nil {
		LogHook(fullMsg)
	}
	if !QuietMode {
		fmt.Fprintln(os.Stdout, fullMsg)
	}
}

// ColorHash returns a truncated, deterministically colored, and bolded Git hash.
func ColorHash(h string) string {
	if len(h) < 7 {
		return h
	}
	short := h[:7]

	// High-contrast, vibrant colors from the xterm-256 palette.
	// Avoiding very dark or very light colors for readability.
	colors := []string{
		"1",   // Red
		"2",   // Green
		"3",   // Yellow
		"4",   // Blue
		"5",   // Magenta
		"6",   // Cyan
		"9",   // Bright Red
		"10",  // Bright Green
		"11",  // Bright Yellow
		"12",  // Bright Blue
		"13",  // Bright Magenta
		"14",  // Bright Cyan
		"202", // Orange
		"208", // Dark Orange
		"214", // Orange1
		"165", // Magenta3
		"135", // MediumPurple3
		"81",  // SteelBlue1
		"118", // Chartreuse1
		"226", // Yellow1
	}

	// Use FNV-1a hashing for deterministic and uniform color distribution.
	f := fnv.New32a()
	f.Write([]byte(h))
	hashVal := f.Sum32()

	c := colors[hashVal%uint32(len(colors))]
	return lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(c)).Render(short)
}

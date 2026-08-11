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

package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/sce-taid/cloud-workstations-custom-image-examples/skills/persona-swe/scripts/internal/gitseep/models"
)

// PipelineEventMsg is a wrapper for models.PipelineEvent to be used as a tea.Msg.
type PipelineEventMsg models.PipelineEvent

// EventBus provides a centralized mechanism for distributing pipeline events to UI components.
type EventBus struct {
	subscribers []eventHandler
}

// eventHandler is a function that processes a pipeline event and returns an optional command.
type eventHandler func(models.PipelineEvent) tea.Cmd

// NewEventBus creates a new EventBus instance.
func NewEventBus() *EventBus {
	return &EventBus{}
}

// Subscribe adds a handler to the bus.
func (b *EventBus) Subscribe(h eventHandler) {
	b.subscribers = append(b.subscribers, h)
}

// Publish broadcasts an event to all subscribers.
func (b *EventBus) Publish(ev models.PipelineEvent) tea.Cmd {
	var cmds []tea.Cmd
	for _, h := range b.subscribers {
		if cmd := h(ev); cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	return tea.Batch(cmds...)
}

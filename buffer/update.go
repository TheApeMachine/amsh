package buffer

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Define message types within the buffer package
type SetFilenameMsg string

// Mode represents the buffer mode
type Mode int

const (
	NormalMode Mode = iota
	InsertMode
)

type ComponentMsg struct {
	ComponentName string
	InnerMsg      tea.Msg
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case ComponentMsg:
		if component, ok := m.components[msg.ComponentName]; ok {
			_, cmd := component.Update(msg.InnerMsg)
			return m, cmd
		}
	case tea.WindowSizeMsg:
		// Handle window size changes
	}

	var cmds []tea.Cmd
	for name, component := range m.components {
		_, cmd := component.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		m.UpdateComponentView(name, component.View())
	}

	return m, tea.Batch(cmds...)
}

type StatusUpdateMsg struct {
	Filename string
	Mode     Mode
}

package editor

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/theapemachine/amsh/messages"
)

// Define the custom message type
type SetFilenameMsg string

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.mode {
		case NormalMode:
			if msg.String() == "i" {
				m.mode = InsertMode
				return m, m.sendStatusUpdate()
			}
		case InsertMode:
			if msg.String() == "esc" {
				m.mode = NormalMode
				return m, m.sendStatusUpdate()
			}
		}
	case tea.WindowSizeMsg:
		m.SetSize(msg.Width, msg.Height)
	case SetFilenameMsg:
		m.filename = string(msg)
		m.loadContent(m.filename)
	}

	// Update all components and sync content
	for i := range m.components {
		newModel, newCmd := m.components[i].Update(msg)
		if newCmd != nil {
			cmd = tea.Batch(cmd, newCmd)
		}
		m.components[i] = newModel
		m.syncContent(i)
	}

	m.StatusBar, _ = m.StatusBar.Update(msg)
	return m, cmd
}

func (m *Model) syncContent(index int) {
	newContent := m.components[index].Value()
	if newContent != m.content {
		m.content = newContent
		m.saveMutex.Lock()
		m.lastSaved = newContent
		m.saveMutex.Unlock()
	}
}

func (m *Model) sendStatusUpdate() tea.Cmd {
	return func() tea.Msg {
		return messages.StatusUpdateMsg{
			Filename: m.filename,
			Mode:     messages.Mode(m.mode.String()),
		}
	}
}

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
		m.loadContent()
	}

	m.updateContent()
	return m, cmd
}

func (m *Model) sendStatusUpdate() tea.Cmd {
	return func() tea.Msg {
		return StatusUpdateMsg{
			Filename: m.filename,
			Mode:     m.mode,
		}
	}
}

type StatusUpdateMsg struct {
	Filename string
	Mode     Mode
}

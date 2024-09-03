package statusbar

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Define message types within the statusbar package
type StatusUpdateMsg struct {
	Filename string
	Mode     Mode
}

// Mode represents the editor mode
type Mode int

const (
	NormalMode Mode = iota
	InsertMode
)

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case StatusUpdateMsg:
		m.filename = msg.Filename
		m.mode = "Normal"
	case tea.WindowSizeMsg:
		m.SetSize(msg.Width)
	}
	return m, nil
}

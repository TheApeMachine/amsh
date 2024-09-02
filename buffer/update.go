package buffer

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/theapemachine/amsh/messages"
)

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.mode {
		case messages.NormalMode:
			if msg.String() == "i" {
				m.mode = messages.InsertMode
				return m, m.sendStatusUpdate()
			}
		case messages.InsertMode:
			if msg.String() == "esc" {
				m.mode = messages.NormalMode
				return m, m.sendStatusUpdate()
			}
		}
	case tea.WindowSizeMsg:
		m.SetSize(msg.Width, msg.Height)
	}

	m.updateContent()
	return m, cmd
}

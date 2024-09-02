package statusbar

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/theapemachine/amsh/messages"
)

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case messages.StatusUpdateMsg:
		m.filename = msg.Filename
		m.mode = msg.Mode
	case tea.WindowSizeMsg:
		m.width = msg.Width
	}
	return m, nil
}

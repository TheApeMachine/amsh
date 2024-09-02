package statusbar

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/theapemachine/amsh/messages"
)

type Model struct {
	filename string
	mode     messages.Mode
	width    int
	style    lipgloss.Style
}

func New() Model {
	return Model{
		mode:  messages.NormalMode,
		width: 80,
		style: lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("57")).
			Padding(0, 1),
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m *Model) SetSize(width int) {
	m.width = width
}

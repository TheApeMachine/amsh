package statusbar

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	filename string
	mode     string
	width    int
	style    lipgloss.Style
}

func New() Model {
	return Model{
		mode:  "Normal",
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

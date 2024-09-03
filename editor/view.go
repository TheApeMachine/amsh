package editor

import (
	"github.com/charmbracelet/lipgloss"
)

func (m *Model) View() string {
	var views []string
	for i := range m.inputs {
		views = append(views, m.inputs[i].View())
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, views...) + "\n\n"
}

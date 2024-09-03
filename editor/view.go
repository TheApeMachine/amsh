package editor

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

/*
View renders the current state of the editor.
This method is part of the tea.Model interface and is responsible for generating
the visual representation of the editor component.
It uses lipgloss to create a horizontally joined layout of all input areas,
allowing for a flexible and visually appealing editor interface.
*/
func (m *Model) View() string {
	builder := strings.Builder{}

	// Render each input area
	for _, input := range m.inputs {
		builder.WriteString(lipgloss.NewStyle().Width(m.width / 2).Render(input.View()))
	}

	// Join all input areas horizontally
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		builder.String(),
	)
}

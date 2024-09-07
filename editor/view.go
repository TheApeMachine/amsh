package editor

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/theapemachine/amsh/components"
)

/*
View renders the current state of the editor.
This method is part of the tea.Model interface and is responsible for generating
the visual representation of the editor component.
It uses lipgloss to create a horizontally joined layout of all input areas,
allowing for a flexible and visually appealing editor interface.
*/
func (m *Model) View() string {
	if m.state == components.Inactive {
		return ""
	}

	v := []string{}

	// Render each input area
	for _, input := range m.inputs {
		v = append(v, input.View())
	}

	// Join all input areas horizontally
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		v...,
	)
}

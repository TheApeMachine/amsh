package editor

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/theapemachine/amsh/components"
	"github.com/theapemachine/amsh/logger"
)

/*
View renders the current state of the editor.
This method is part of the tea.Model interface and is responsible for generating
the visual representation of the editor component.
It uses lipgloss to create a horizontally joined layout of all input areas,
allowing for a flexible and visually appealing editor interface.
*/
func (model *Model) View() string {
	if model.state == components.Inactive {
		return ""
	}

	v := []string{}

	// Render each input area
	for _, input := range model.inputs {
		logger.Debug("Rendering textarea")
		v = append(v, input.View())
	}

	// Join all input areas horizontally
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		v...,
	)
}

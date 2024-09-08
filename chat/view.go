package chat

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/theapemachine/amsh/components"
)

func (model *Model) View() string {
	if model.state != components.Focused {
		return ""
	}

	// Create a styled border
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63"))

	// Combine the viewport and textarea with vertical padding
	content := lipgloss.JoinVertical(
		lipgloss.Top,
		model.viewport.View(),
		"\n", // Add a newline for spacing
		model.textarea.View(),
	)

	// Apply the border to the content
	borderedContent := borderStyle.Render(content)

	return lipgloss.Place(
		model.width,
		model.height,
		lipgloss.Center,
		lipgloss.Center,
		borderedContent,
	)
}

package chat

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

func (model *Model) View() string {
	if !model.active {
		return ""
	}

	if model.err != nil {
		return model.styles.FocusedBorderStyle.Render(
			fmt.Sprintf("Error: %s", model.err),
		)
	}

	return lipgloss.JoinVertical(
		lipgloss.Top,
		model.viewport.View(),
		model.textarea.View(),
	)
}

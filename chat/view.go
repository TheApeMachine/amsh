package chat

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/theapemachine/amsh/ui"
)

func (m *Model) View() string {
	if !m.active {
		return ""
	}

	if m.err != nil {
		return ui.FocusedBorderStyle.Render(
			fmt.Sprintf("Error: %s", m.err),
		)
	}

	return lipgloss.JoinVertical(
		lipgloss.Top,
		m.viewport.View(),
		m.textarea.View(),
	)
}

package statusbar

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

/*
View renders the current state of the statusbar.
This method is part of the tea.Model interface and is responsible for generating
the visual representation of the statusbar component.
It formats the current mode and filename into a string and applies the statusbar's style,
ensuring a consistent and informative display at the bottom of the application.
*/
func (m Model) View() string {
	if !m.active {
		return ""
	}

	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFDF5")).
		Background(lipgloss.Color("#FF5F87")).
		Padding(0, 1)

	modeStyle := statusStyle.Copy().
		Background(lipgloss.Color("#6124DF"))

	fileStyle := statusStyle.Copy().
		Background(lipgloss.Color("#A550DF"))

	content := lipgloss.JoinHorizontal(
		lipgloss.Center,
		modeStyle.Render(fmt.Sprintf("MODE: %s", m.mode)),
		fileStyle.Render(fmt.Sprintf("FILE: %s", m.filename)),
	)

	if m.err != nil {
		errorStyle := statusStyle.Copy().
			Foreground(lipgloss.Color("#FF0000")).
			Background(lipgloss.Color("#FFFF00"))
		content = errorStyle.Render(fmt.Sprintf("ERROR: %s", m.err))
	}

	return lipgloss.NewStyle().
		Width(m.width).
		Render(content)
}

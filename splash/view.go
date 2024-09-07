package splash

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/theapemachine/amsh/components"
	"github.com/theapemachine/amsh/ui"
)

func (model *Model) View() string {
	if model.state == components.Inactive {
		return ""
	}

	return lipgloss.Place(
		model.width, model.height,
		lipgloss.Center, lipgloss.Center,
		model.styles.DialogBox.Render(
			lipgloss.NewStyle().Width(50).Align(lipgloss.Center).Render(ui.Logo),
		),
		lipgloss.WithWhitespaceChars("猫咪"),
		lipgloss.WithWhitespaceForeground(ui.Subtle),
	)
}

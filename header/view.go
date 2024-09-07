package header

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/theapemachine/amsh/components"
)

func (model *Model) View() string {
	if model.state == components.Inactive {
		return ""
	}

	return lipgloss.PlaceHorizontal(
		model.width,
		lipgloss.Left,
		model.Select(model.err),
		lipgloss.WithWhitespaceChars("/"),
		lipgloss.WithWhitespaceForeground(model.styles.HeaderText.GetForeground()),
	)
}

func (model *Model) Select(err error) string {
	if err != nil {
		return model.Error(err)
	}

	return model.Healthy()
}

func (model *Model) Healthy() string {
	return lipgloss.NewStyle().Foreground(model.styles.HeaderText.GetForeground()).Bold(true).Padding(0, 1, 0, 2).Render("amsh")
}

func (model *Model) Error(err error) string {
	return model.styles.ErrorHeaderText.Render(err.Error())
}

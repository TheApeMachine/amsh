package statusbar

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/theapemachine/amsh/components"
)

/*
View renders the current state of the statusbar.
This method is part of the tea.Model interface and is responsible for generating
the visual representation of the statusbar component.
It formats the current mode and filename into a string and applies the statusbar's style,
ensuring a consistent and informative display at the bottom of the application.
*/
func (model *Model) View() string {
	if model.state == components.Inactive {
		return ""
	}

	w := lipgloss.Width

	statusKey := model.styles.StatusStyle.Render(model.mode)
	encoding := model.styles.EncodingStyle.Render("UTF-8")
	fishCake := model.styles.StatusNuggetStyle.Render("fishcake")
	statusVal := model.styles.StatusText.Width(model.width - w(statusKey) - w(encoding) - w(fishCake)).Render(model.filename)

	bar := lipgloss.JoinHorizontal(lipgloss.Top,
		statusKey,
		statusVal,
		encoding,
		fishCake,
	)

	return model.styles.StatusBarStyle.Width(model.width).Render(bar)
}

func (model *Model) Select(err error) string {
	if err != nil {
		return model.Error(err)
	}

	return model.Healthy()
}

func (model *Model) Healthy() string {
	return lipgloss.NewStyle().Foreground(
		model.styles.HeaderText.GetForeground(),
	).Bold(true).Padding(0, 1, 0, 2).Render(model.mode)
}

func (model *Model) Error(err error) string {
	return model.styles.ErrorHeaderText.Render(err.Error())
}

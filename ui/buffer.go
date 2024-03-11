package ui

import (
	"github.com/charmbracelet/lipgloss"
)

func (state *State) View() string {
	var views []string

	for _, widget := range state.widgets {
		views = append(views, widget.View())
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, views...)
}

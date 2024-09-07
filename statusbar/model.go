package statusbar

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/theapemachine/amsh/ui"
)

/*
Model represents the state of the statusbar component.
It manages the current filename, mode, width, and styling of the statusbar.
This structure is crucial for displaying relevant information at the bottom of the application.
*/
type Model struct {
	filename string
	mode     string
	width    int
	style    lipgloss.Style
	active   bool
	layout   ui.LayoutPreference
	err      error
}

/*
New creates a new statusbar model with default values.
It initializes the statusbar with a default mode, width, and style.
This factory function ensures that every new statusbar instance starts with a consistent initial state.
*/
func New() Model {
	return Model{
		mode:  "Normal",
		width: 80,
		style: lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("57")).
			Padding(0, 1),
		active: true,
		layout: ui.Bottom,
	}
}

/*
Init initializes the statusbar model.
This method is part of the tea.Model interface and is called when the statusbar component starts.
Currently, it doesn't perform any initialization actions, but it's included for consistency
and potential future use.
*/
func (m Model) Init() tea.Cmd {
	return nil
}

/*
SetSize adjusts the width of the statusbar.
This method is crucial for responsive design, ensuring the statusbar
adapts to window size changes and maintains a consistent appearance.
*/
func (m *Model) SetSize(width int) {
	m.width = width
}

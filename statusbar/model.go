package statusbar

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/theapemachine/amsh/components"
	"github.com/theapemachine/amsh/ui"
)

var modeMap = map[ui.Mode]string{
	ui.ModeNormal: "Normal",
	ui.ModeInsert: "Insert",
	ui.ModeVisual: "Visual",
}

/*
Model represents the state of the statusbar component.
It manages the current filename, mode, width, and styling of the statusbar.
This structure is crucial for displaying relevant information at the bottom of the application.
*/
type Model struct {
	filename   string
	mode       string
	width      int
	styles     *ui.Styles
	modeStyles map[string]lipgloss.Style
	plugin     string
	state      components.State
}

/*
New creates a new statusbar model with default values.
It initializes the statusbar with a default mode, width, and style.
This factory function ensures that every new statusbar instance starts with a consistent initial state.
*/
func New(width int) *Model {
	styles := ui.NewStyles()

	return &Model{
		mode:   modeMap[ui.ModeNormal],
		width:  width,
		styles: ui.NewStyles(),
		modeStyles: map[string]lipgloss.Style{
			"normal": styles.ModeNormalStyle,
			"insert": styles.ModeInsertStyle,
			"visual": styles.ModeVisualStyle,
		},
		state: components.Inactive,
	}
}

/*
Init initializes the statusbar model.
This method is part of the tea.Model interface and is called when the statusbar component starts.
Currently, it doesn't perform any initialization actions, but it's included for consistency
and potential future use.
*/
func (model *Model) Init() tea.Cmd {
	return nil
}

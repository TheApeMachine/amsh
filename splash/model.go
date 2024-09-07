package splash

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/theapemachine/amsh/components"
	"github.com/theapemachine/amsh/ui"
)

type Model struct {
	width              int
	height             int
	styles             *ui.Styles
	whitespaceProgress int
	dialogBoxYPosition int
	state              components.State
}

func New(width, height int) *Model {
	return &Model{
		width:              width,
		height:             height,
		styles:             ui.NewStyles(),
		whitespaceProgress: 0,
		dialogBoxYPosition: 0, // Start from the top
		state:              components.Active,
	}
}

func (model *Model) Init() tea.Cmd {
	return nil
}

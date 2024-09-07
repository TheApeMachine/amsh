package header

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/theapemachine/amsh/components"
	"github.com/theapemachine/amsh/ui"
)

type Model struct {
	width  int
	height int
	styles *ui.Styles
	state  components.State
	err    error
}

func New(width, height int) *Model {
	return &Model{
		width:  width,
		height: height,
		styles: ui.NewStyles(),
		state:  components.Inactive,
	}
}

func (model *Model) Init() tea.Cmd {
	return nil
}

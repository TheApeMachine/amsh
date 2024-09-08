package chat

import (
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/theapemachine/amsh/components"
	"github.com/theapemachine/amsh/ui"
)

type Model struct {
	viewport viewport.Model
	textarea textarea.Model
	styles   *ui.Styles
	state    components.State
	width    int
	height   int
}

func New(width, height int) *Model {
	ta := textarea.New()
	ta.SetHeight(height / 4)
	ta.SetWidth(width / 4)
	ta.Focus()

	return &Model{
		viewport: viewport.New(width/4, height/4),
		textarea: textarea.New(),
		styles:   ui.NewStyles(),
		state:    components.Inactive,
		width:    width,
		height:   height,
	}
}

func (model *Model) Init() tea.Cmd {
	return nil
}

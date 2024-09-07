package chat

import (
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/theapemachine/amsh/ui"
)

type Model struct {
	viewport viewport.Model
	textarea textarea.Model
	styles   *ui.Styles
	active   bool
	err      error
}

func New() *Model {
	return &Model{
		viewport: viewport.New(30, 5),
		textarea: textarea.New(),
		styles:   ui.NewStyles(),
		active:   false,
	}
}

func (model *Model) Init() tea.Cmd {
	return textarea.Blink
}

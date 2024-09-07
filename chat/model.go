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
	layout   ui.LayoutPreference
	active   bool
	err      error
}

func New() *Model {
	return &Model{
		viewport: viewport.New(30, 5),
		textarea: textarea.New(),
		layout:   ui.Overlay,
		active:   false,
	}
}

func (m *Model) Init() tea.Cmd {
	return textarea.Blink
}

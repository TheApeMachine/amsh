package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/theapemachine/amsh/ui/widgets"
)

type State struct {
	widgets []widgets.Widget
}

func NewState() *State {
	state := &State{
		widgets: []widgets.Widget{widgets.NewTextArea()},
	}

	state.widgets[0].Focus()
	return state
}

func (state *State) Init() tea.Cmd {
	return state.widgets[0].Init()
}

func (state *State) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return state, tea.Quit
		}
	}

	for _, widget := range state.widgets {
		_, cmd := widget.Update(msg)
		cmds = append(cmds, cmd)
	}

	return state, tea.Batch(cmds...)
}

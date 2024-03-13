package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/theapemachine/amsh/ui/widgets"
)

type State struct {
	widgets []tea.Model
	ptr     int
}

func NewState() *State {
	state := &State{
		widgets: []tea.Model{widgets.NewMiniScript()},
		ptr:     0,
	}

	return state
}

func (state *State) Init() tea.Cmd {
	var cmds []tea.Cmd

	for _, widget := range state.widgets {
		cmds = append(cmds, widget.Init())
	}

	return tea.Batch(cmds...)
}

func (state *State) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return state, tea.Quit
		}
	}

	var cmd tea.Cmd

	for idx, widget := range state.widgets {
		state.widgets[idx], cmd = widget.Update(msg)
		cmds = append(cmds, cmd)
	}

	// for _, widget := range state.widgets {
	// 	_, cmd := widget.Update(msg)
	// 	cmds = append(cmds, cmd)
	// }

	return state, tea.Batch(cmds...)
}

func (state *State) View() string {
	var builder []string

	for _, widget := range state.widgets {
		builder = append(builder, widget.View())
	}

	return lipgloss.JoinVertical(lipgloss.Top, builder...)
}

package buffer

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/theapemachine/amsh/ui"
)

/*
Update is a message broker for the components in the buffer.
This method is part of the tea.Model interface and is responsible for handling
all incoming messages, updating the appropriate components, and generating any
necessary commands.
It acts as a central hub for message routing and state management, ensuring that
all components are updated correctly based on the received messages.
*/
func (model *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" {
			cmds = append(cmds, tea.Quit)
		}

		cmds = append(cmds, model.keyHandler.Handle(msg))
	}

	model.dispatchMsg(msg, cmds)
	return model, tea.Batch(cmds...)
}

/*
SetMode sets the current mode of the buffer and updates the key handler.
*/
func (model *Model) SetMode(mode ui.Mode) {
	model.mode = mode
	model.keyHandler.SetMode(mode)
}

/*
SetSize sets the size of the buffer.
*/
func (model *Model) SetSize(width, height int) {
	model.width, model.height = width, height
}

/*
dispatchMsg dispatches a message to all components.
Each component is responsible for handling, or not handling, its own messages.
*/
func (model *Model) dispatchMsg(msg tea.Msg, cmds []tea.Cmd) []tea.Cmd {
	for idx, component := range model.components {
		var cmd tea.Cmd
		model.components[idx], cmd = component.Update(msg)
		cmds = append(cmds, cmd)
	}

	return cmds
}

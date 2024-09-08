package buffer

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/theapemachine/amsh/logger"
	"github.com/theapemachine/amsh/messages"
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
	logger.StartTick()
	defer logger.EndTick()

	EndSection := logger.StartSection("buffer.Update", "update")
	defer EndSection()

	if msg, ok := msg.(messages.Message[string]); ok {
		logger.Debug("Received message type: %v, Data: %v", msg.Type, msg.Data)
	}

	var (
		cmds []tea.Cmd
	)
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" {
			cmds = append(cmds, tea.Quit)
		}

		if model.mode == ui.ModeNormal {
			model.cmdChan <- msg
		}
	case tea.WindowSizeMsg:
		logger.Debug("<- <tea.WindowSizeMsg> %d, %d", msg.Width, msg.Height)
		model.SetSize(msg.Width, msg.Height)
	case messages.Message[[]int]:
		switch msg.Type {
		case messages.MessageWindowSize:
			model.SetSize(msg.Data[0], msg.Data[1])
		}
	case messages.Message[ui.Mode]:
		model.mode = msg.Data
		cmds = append(cmds, model.dispatchModeMsg(cmds)...)
	case messages.Message[string]:
		switch msg.Type {
		case messages.MessageOpenFile:
			model.dispatchMsg(msg)
		case messages.MessageShow:
			model.dispatchMsg(msg)
		}
	}

	return model, tea.Batch(cmds...)
}

// func (model *Model) handleKeyMsg(msg tea.KeyMsg, cmds []tea.Cmd) []tea.Cmd {
// 	logger.Debug("<- <tea.KeyMsg> %s", msg.String())

// 	switch msg.String() {
// 	case "q":
// 		cmds = append(cmds, tea.Quit)
// 	case "esc":
// 		model.mode = ui.ModeNormal
// 		cmds = append(cmds, model.dispatchModeMsg(cmds)...)
// 	case "i":
// 		model.mode = ui.ModeInsert
// 		cmds = append(cmds, model.dispatchModeMsg(cmds)...)
// 	case "v":
// 		model.mode = ui.ModeVisual
// 		cmds = append(cmds, model.dispatchModeMsg(cmds)...)
// 	}

// 	return cmds
// }

func (model *Model) SetSize(width, height int) {
	model.width, model.height = width, height
}

func (model *Model) dispatchMsg(msg tea.Msg) {
	for _, component := range model.components {
		logger.Debug("Dispatching message to component: %T", component)
		component.Update(msg)
	}
}

func (model *Model) dispatchModeMsg(cmds []tea.Cmd) []tea.Cmd {
	for _, component := range model.components {
		_, cmd := component.Update(messages.Message[ui.Mode]{
			Type: messages.MessageMode,
			Data: model.mode,
		})

		cmds = append(cmds, cmd)
	}

	return cmds
}

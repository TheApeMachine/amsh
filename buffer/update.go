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

	logger.Debug("<- <%v>", msg)

	var (
		cmds []tea.Cmd
		cmd  tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		logger.Debug("<- <tea.KeyMsg> %s", msg.String())

		switch msg.String() {
		case "q":
			return model, tea.Quit
		case "esc":
			model.mode = ui.ModeNormal
			cmds = append(cmds, model.dispatchModeMsg(cmds)...)
		case "i":
			model.mode = ui.ModeInsert
			cmds = append(cmds, model.dispatchModeMsg(cmds)...)
		case "v":
			model.mode = ui.ModeVisual
			cmds = append(cmds, model.dispatchModeMsg(cmds)...)
		}
	case tea.WindowSizeMsg:
		logger.Debug("<- <tea.WindowSizeMsg> %d, %d", msg.Width, msg.Height)
		model.SetSize(msg.Width, msg.Height)
	case messages.Message[[]int]:
		switch msg.Type {
		case messages.MessageWindowSize:
			model.SetSize(msg.Data[0], msg.Data[1])
		}
	}

	for idx, component := range model.components {
		model.components[idx], cmd = component.Update(msg)
		cmds = append(cmds, cmd)
	}

	return model, tea.Batch(cmds...)
}

func (model *Model) SetSize(width, height int) {
	model.width, model.height = width, height
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

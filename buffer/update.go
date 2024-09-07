package buffer

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/theapemachine/amsh/logger"
	"github.com/theapemachine/amsh/messages"
)

/*
Update is a message broker for the components in the buffer.
This method is part of the tea.Model interface and is responsible for handling
all incoming messages, updating the appropriate components, and generating any
necessary commands.

It acts as a central hub for message routing and state management, ensuring that
all components are updated correctly based on the received messages.
*/
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		if msg.String() == "q" {
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		logger.Debug("<- <tea.WindowSizeMsg> %d, %d", msg.Width, msg.Height)
		m.width = msg.Width
		m.height = msg.Height
	case messages.Message[[]int]:
		logger.Debug("<- <messages.Message[[]int]> %v, %v, %d", msg.Context, msg.Data, msg.Type)

		switch msg.Type {
		case messages.MessageWindowSize:
			logger.Debug("<- <messages.MessageWindowSize> %d, %d", msg.Data[0], msg.Data[1])
			m.width = msg.Data[0]
			m.height = msg.Data[1]
		}
	}

	for idx, component := range m.components {
		m.components[idx], cmd = component.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

package chat

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/theapemachine/amsh/components"
	"github.com/theapemachine/amsh/logger"
	"github.com/theapemachine/amsh/messages"
)

func (model *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	logger.Debug("<- <%T|%v>", msg, msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		model.textarea.Update(msg)
	case messages.Message[string]:
		switch msg.Type {
		case messages.MessageShow:
			if msg.Data == "chat" {
				model.state = components.Focused
			}
		}
	}

	return model, nil
}

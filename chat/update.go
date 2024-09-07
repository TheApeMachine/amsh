package chat

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/theapemachine/amsh/logger"
	"github.com/theapemachine/amsh/messages"
)

func (model *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case messages.Message[string]:
		logger.Log("chat message received: %v", msg)
	}
	return model, nil
}

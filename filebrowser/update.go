package filebrowser

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/theapemachine/amsh/components"
	"github.com/theapemachine/amsh/messages"
)

func (model *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case messages.Message[string]:
		switch msg.Type {
		case messages.MessageShow:
			switch msg.Data {
			case "filebrowser":
				model.state = components.Focused
			default:
				model.state = components.Inactive
			}
		}
	}

	return model, nil
}

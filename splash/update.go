package splash

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/theapemachine/amsh/components"
	"github.com/theapemachine/amsh/messages"
)

func (model *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		model.SetSize(model.width, model.height)
	case messages.Message[[]int]:
		if msg.Type == messages.MessageWindowSize {
			model.SetSize(msg.Data[0], msg.Data[1])
		}
	case messages.Message[string]:
		if msg.Type == messages.MessageOpenFile || msg.Type == messages.ComponentLoaded {
			model.state = components.Inactive
		}
	}

	return model, nil
}

func (model *Model) SetSize(width, height int) {
	model.width = width
	model.height = height
}

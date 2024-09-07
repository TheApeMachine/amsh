package header

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/theapemachine/amsh/components"
	"github.com/theapemachine/amsh/messages"
)

func (model *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case messages.Message[error]:
		if !messages.ShouldProcessMessage(model.state, msg.Context) {
			return model, nil
		}

		model.err = msg.Data
	case messages.Message[[]int]:
		if !messages.ShouldProcessMessage(model.state, msg.Context) {
			return model, nil
		}

		model.SetSize(msg.Data[0], msg.Data[1])
	case messages.Message[string]:
		if !messages.ShouldProcessMessage(model.state, msg.Context) {
			return model, nil
		}

		if msg.Type == messages.MessageOpenFile {
			model.state = components.Active
		}
	}

	return model, nil
}

func (model *Model) SetSize(width, height int) {
	model.width = width
	model.height = height
}

package splash

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/theapemachine/amsh/components"
	"github.com/theapemachine/amsh/messages"
)

func (model *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		model.SetSize(m.Width, m.Height)
	case messages.Message[[]int]:
		if m.Type == messages.MessageWindowSize {
			model.SetSize(m.Data[0], m.Data[1])
		}
	case messages.Message[string]:
		if m.Type == messages.MessageOpenFile || m.Type == messages.ComponentLoaded {
			model.state = components.Inactive
		}
	}

	return model, nil
}

func (model *Model) SetSize(width, height int) {
	model.width = width
	model.height = height
}

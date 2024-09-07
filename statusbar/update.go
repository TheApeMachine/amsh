package statusbar

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/theapemachine/amsh/components"
	"github.com/theapemachine/amsh/logger"
	"github.com/theapemachine/amsh/messages"
	"github.com/theapemachine/amsh/ui"
)

/*
Update handles all incoming messages for the statusbar component.
This method is part of the tea.Model interface and is responsible for updating the statusbar state
based on various events such as status updates and window size changes.
It ensures that the statusbar always displays the most current information.
*/
func (model *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	EndSection := logger.StartSection("statusbar.Update", "update")
	defer EndSection()

	logger.Debug("<- <%v>", msg)

	switch msg := msg.(type) {
	case messages.Message[ui.Mode]:
		if !messages.ShouldProcessMessage(model.state, msg.Context) {
			return model, nil
		}

		model.mode = modeMap[msg.Data]
	case tea.WindowSizeMsg:
		model.SetSize(msg.Width)
	case messages.Message[[]int]:
		if !messages.ShouldProcessMessage(model.state, msg.Context) {
			return model, nil
		}

		model.SetSize(msg.Data[0])
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

/*
SetSize adjusts the width of the statusbar.
This method is crucial for responsive design, ensuring the statusbar
adapts to window size changes and maintains a consistent appearance.
*/
func (model *Model) SetSize(width int) {
	model.width = width
}

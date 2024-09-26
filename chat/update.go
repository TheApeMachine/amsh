package chat

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/theapemachine/amsh/components"
	"github.com/theapemachine/amsh/data"
	"github.com/theapemachine/amsh/logger"
)

func (model *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	logger.Debug("<- <%T|%v>", msg, msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			model.state = components.Inactive
			return model, nil
		case "enter":
			if _, model.err = model.aiIO.Write([]byte(
				model.textarea.Value(),
			)); model.err != nil {
				return model, nil
			}
			model.textarea.Reset()
		default:
			model.textarea.Update(msg)
		}
	case data.Artifact:
		var scope string

		if scope, model.err = msg.Scope(); model.err != nil {
			logger.Error("Error getting scope: %v", model.err)
			return model, nil
		}

		switch scope {
		case "chat":
			if model.state == components.Focused {
				return model, nil
			}
			model.state = components.Focused
			return model, model.textarea.Focus()
		}
	}

	return model, nil
}

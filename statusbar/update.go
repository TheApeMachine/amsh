package statusbar

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/theapemachine/amsh/logger"
	"github.com/theapemachine/amsh/messages"
)

/*
Update handles all incoming messages for the statusbar component.
This method is part of the tea.Model interface and is responsible for updating the statusbar state
based on various events such as status updates and window size changes.
It ensures that the statusbar always displays the most current information.
*/
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	logger.Log("Statusbar received message: %T", msg)
	switch msg := msg.(type) {
	case messages.Message[string]:
	case tea.WindowSizeMsg:
		m.SetSize(msg.Width)
	}
	return m, nil
}

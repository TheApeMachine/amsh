package statusbar

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/theapemachine/amsh/messages"
)

/*
StatusUpdateMsg is a message type used to update the statusbar's content.
It carries information about the current filename and editing mode.
This custom message type allows for targeted updates to the statusbar's state.
*/
type StatusUpdateMsg struct {
	Filename string
	Mode     messages.Mode
}

/*
Update handles all incoming messages for the statusbar component.
This method is part of the tea.Model interface and is responsible for updating the statusbar state
based on various events such as status updates and window size changes.
It ensures that the statusbar always displays the most current information.
*/
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case messages.StatusUpdateMsg:
		m.filename = msg.Filename
		m.mode = fmt.Sprintf("%d", msg.Mode)
	case tea.WindowSizeMsg:
		m.SetSize(msg.Width)
	}
	return m, nil
}

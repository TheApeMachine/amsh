package buffer

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/theapemachine/amsh/logger"
	"github.com/theapemachine/amsh/messages"
)

// Mode represents the buffer mode
type Mode int

const (
	NormalMode Mode = iota
	InsertMode
)

/*
ComponentMsg is a message that is sent to a component in the buffer.
This custom message type allows for targeted message delivery to specific components,
enabling fine-grained control over component updates and interactions.
*/
type ComponentMsg struct {
	ComponentName string
	InnerMsg      tea.Msg
}

/*
Update is a message broker for the components in the buffer.
This method is part of the tea.Model interface and is responsible for handling
all incoming messages, updating the appropriate components, and generating any
necessary commands.

It acts as a central hub for message routing and state management, ensuring that
all components are updated correctly based on the received messages.
*/
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	logger.Log("Buffer received message: %T", msg)

	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case messages.StatusUpdateMsg:
		logger.Log("Buffer received StatusUpdateMsg")
		if sb, ok := m.components["statusbar"]; ok {
			updatedComponent, cmd := sb.Update(msg)
			m.components["statusbar"] = updatedComponent
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
	case messages.SetActiveComponentMsg:
		logger.Log("Buffer received SetActiveComponentMsg: %s", msg)
		m.SetActiveComponent(string(msg))
	case messages.FileSelectedMsg:
		logger.Log("Buffer received FileSelectedMsg: %s", msg)
		m.SetActiveComponent("editor")
		cmds = append(cmds, func() tea.Msg {
			return messages.ComponentMsg{
				ComponentName: "editor",
				InnerMsg:      messages.SetFilenameMsg(msg),
			}
		})
	case messages.ComponentMsg:
		logger.Log("Buffer received ComponentMsg for %s", msg.ComponentName)
		if component, ok := m.components[msg.ComponentName]; ok {
			updatedComponent, cmd := component.Update(msg.InnerMsg)
			m.components[msg.ComponentName] = updatedComponent
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
	case tea.WindowSizeMsg:
		logger.Log("Buffer received WindowSizeMsg: %v", msg)
		m.width = msg.Width
		m.height = msg.Height
		// Update the size for all components that support it
		for name, component := range m.components {
			if sizer, ok := component.(interface{ SetSize(width, height int) }); ok {
				sizer.SetSize(msg.Width, msg.Height)
			}
			updatedComponent, cmd := component.Update(msg)
			m.components[name] = updatedComponent
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
	default:
		// Pass the message to all components
		// This ensures that components can react to messages even if they're not explicitly handled above
		for name, component := range m.components {
			logger.Log("Buffer passing message to component: %s", name)
			updatedComponent, cmd := component.Update(msg)
			m.components[name] = updatedComponent
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
	}

	return m, tea.Batch(cmds...)
}

/*
StatusUpdateMsg is a message that is sent to the buffer to update the status.
This custom message type allows for updating the status information displayed in the UI,
providing feedback to the user about the current state of the application.
*/
type StatusUpdateMsg struct {
	Filename string
	Mode     Mode
}

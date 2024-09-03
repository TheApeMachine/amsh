package buffer

import (
	"github.com/charmbracelet/lipgloss"
)

/*
View renders the current state of the buffer.
This method is part of the tea.Model interface and is responsible for generating
the visual representation of the buffer and its components.
It uses lipgloss to create a vertically joined layout of the active component and the statusbar.
*/
func (m *Model) View() string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Render the active component
	activeView := m.components[m.activeComponent].View()

	// Render the statusbar
	statusbarView := m.components["statusbar"].View()

	// Combine the views
	// We use lipgloss to create a consistent and visually appealing layout
	return lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.NewStyle().Height(m.height-1).Render(activeView),
		statusbarView,
	)
}

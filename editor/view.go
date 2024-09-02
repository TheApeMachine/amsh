package editor

import (
	"strings"
)

func (m *Model) View() string {
	var s strings.Builder

	// Render the active editor
	activeView := m.components[m.activeEditor].View()
	s.WriteString(activeView)

	// Add padding to push the status bar to the bottom
	for i := 0; i < m.height-strings.Count(s.String(), "\n")-2; i++ {
		s.WriteString("\n")
	}

	// Add the status bar at the bottom
	s.WriteString(m.StatusBar.View())

	return s.String()
}

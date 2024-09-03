package textarea

import (
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

/*
Blink returns the blink command for the textarea.
This method is used to start the cursor blinking animation.
*/
func (m *Model) Blink() tea.Cmd {
	return textarea.Blink
}

/*
View renders the current state of the textarea.
This method is part of the tea.Model interface and is responsible for generating
the visual representation of the textarea component.
It delegates to the underlying bubbles textarea for rendering.
*/
func (m *Model) View() string {
	return m.Model.View()
}

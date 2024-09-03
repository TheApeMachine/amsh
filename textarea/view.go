package textarea

import (
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Blink() tea.Cmd {
	return textarea.Blink
}

func (m Model) View() string {
	return m.textarea.View()
}

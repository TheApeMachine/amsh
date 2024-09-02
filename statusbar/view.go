package statusbar

import "fmt"

func (m Model) View() string {
	statusText := fmt.Sprintf("Mode: %s | File: %s", m.mode, m.filename)
	return m.style.Width(m.width).Render(statusText)
}

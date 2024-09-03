package statusbar

import "fmt"

/*
View renders the current state of the statusbar.
This method is part of the tea.Model interface and is responsible for generating
the visual representation of the statusbar component.
It formats the current mode and filename into a string and applies the statusbar's style,
ensuring a consistent and informative display at the bottom of the application.
*/
func (m Model) View() string {
	statusText := fmt.Sprintf("Mode: %s | File: %s", m.mode, m.filename)
	return m.style.Width(m.width).Render(statusText)
}

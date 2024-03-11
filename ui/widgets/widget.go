package widgets

import tea "github.com/charmbracelet/bubbletea"

type Widget interface {
	tea.Model
	Focus()
}

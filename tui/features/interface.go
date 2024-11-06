package features

import tea "github.com/charmbracelet/bubbletea"

/*
Feature wraps the tea.Model interface and adds a couple of extra methods,
which are used to manage the screens and the layout.
*/
type Feature interface {
	tea.Model
	Model() tea.Model
	Name() string
	Size() (int, int)
}

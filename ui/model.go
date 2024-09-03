package ui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

/*
Model represents the state of the UI component.
It manages key bindings and other UI-related functionality.
*/
type Model struct {
	keys KeyMap
}

/*
KeyMap defines the key bindings for various actions in the UI.
This structure allows for easy customization of keyboard shortcuts.
*/
type KeyMap struct {
	Up    key.Binding
	Down  key.Binding
	Left  key.Binding
	Right key.Binding
	Enter key.Binding
	Back  key.Binding
	Quit  key.Binding
}

/*
New creates a new UI model with default key bindings.
This factory function ensures that every new UI instance
starts with a consistent set of key bindings.
*/
func New() Model {
	return Model{
		keys: KeyMap{
			Up: key.NewBinding(
				key.WithKeys("up", "k"),
				key.WithHelp("↑/k", "move up"),
			),
			Down: key.NewBinding(
				key.WithKeys("down", "j"),
				key.WithHelp("↓/j", "move down"),
			),
			Left: key.NewBinding(
				key.WithKeys("left", "h"),
				key.WithHelp("←/h", "move left"),
			),
			Right: key.NewBinding(
				key.WithKeys("right", "l"),
				key.WithHelp("→/l", "move right"),
			),
			Enter: key.NewBinding(
				key.WithKeys("enter"),
				key.WithHelp("enter", "select"),
			),
			Back: key.NewBinding(
				key.WithKeys("esc"),
				key.WithHelp("esc", "back"),
			),
			Quit: key.NewBinding(
				key.WithKeys("q", "ctrl+c"),
				key.WithHelp("q", "quit"),
			),
		},
	}
}

/*
Init initializes the UI model.
This method is part of the tea.Model interface and is called when the UI component starts.
Currently, it doesn't perform any initialization actions, but it's included for consistency
and potential future use.
*/
func (m Model) Init() tea.Cmd {
	return nil
}

/*
Update handles all incoming messages for the UI component.
This method is part of the tea.Model interface and is responsible for updating the UI state
based on various events. Currently, it's a placeholder for future UI-specific logic.
*/
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

/*
View renders the current state of the UI.
This method is part of the tea.Model interface and is responsible for generating
the visual representation of the UI component. Currently, it's a placeholder
for future UI-specific rendering logic.
*/
func (m Model) View() string {
	return ""
}
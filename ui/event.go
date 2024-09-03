package ui

import "github.com/charmbracelet/bubbles/key"

type Event struct {
	keymap struct {
		next, prev, add, remove, quit key.Binding
	}
}

func NewEvent() *Event {
	return &Event{keymap: struct {
		next, prev, add, remove, quit key.Binding
	}{
		next: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next"),
		),
		prev: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "prev"),
		),
		add: key.NewBinding(
			key.WithKeys("ctrl+n"),
			key.WithHelp("ctrl+n", "add an editor"),
		),
		remove: key.NewBinding(
			key.WithKeys("ctrl+w"),
			key.WithHelp("ctrl+w", "remove an editor"),
		),
	}}
}

package ui

import "github.com/charmbracelet/bubbles/key"

type Event struct {
	Keymap struct {
		Next, Prev, Add, Remove, Quit key.Binding
	}
}

func NewEvent() *Event {
	return &Event{Keymap: struct {
		Next, Prev, Add, Remove, Quit key.Binding
	}{
		Next: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next"),
		),
		Prev: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "prev"),
		),
		Add: key.NewBinding(
			key.WithKeys("ctrl+n"),
			key.WithHelp("ctrl+n", "add an editor"),
		),
		Remove: key.NewBinding(
			key.WithKeys("ctrl+w"),
			key.WithHelp("ctrl+w", "remove an editor"),
		),
	}}
}

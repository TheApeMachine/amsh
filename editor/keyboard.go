package editor

import (
	"github.com/charmbracelet/bubbles/key"
)

type KeyboardManager struct {
	KeyMap keymap
}

func NewKeyboardManager() *KeyboardManager {
	return &KeyboardManager{
		KeyMap: keymap{
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
			quit: key.NewBinding(
				key.WithKeys("esc", "ctrl+c"),
				key.WithHelp("esc", "quit"),
			),
		},
	}
}

func (km *KeyboardManager) UpdateKeybindings(inputsLen int) {
	km.KeyMap.add.SetEnabled(inputsLen < maxInputs)
	km.KeyMap.remove.SetEnabled(inputsLen > minInputs)
}

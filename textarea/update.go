package textarea

import (
	tea "github.com/charmbracelet/bubbletea"
)

/*
Update handles all incoming messages for the textarea component.
This method is part of the tea.Model interface and is responsible for updating the textarea state
based on various events such as key presses and window size changes.
It delegates to the underlying bubbles textarea for most of its functionality.
*/
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.SetSize(msg.Width, msg.Height)
	}

	m.updateKeybindings()
	m.sizeInput()

	return m, tea.Batch(cmds...)
}

/*
SetSize adjusts the size of the textarea.
This method is crucial for responsive design, ensuring the textarea
adapts to window size changes and maintains a consistent appearance.
*/
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
}

func (m *Model) sizeInput() {
	m.SetWidth(m.width)
	m.SetHeight(m.height)
}

func (m *Model) updateKeybindings() {
	m.event.Keymap.Add.SetEnabled(m.enabled)
	m.event.Keymap.Remove.SetEnabled(!m.enabled)
}

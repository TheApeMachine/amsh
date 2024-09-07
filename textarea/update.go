package textarea

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/theapemachine/amsh/components"
	"github.com/theapemachine/amsh/logger"
	"github.com/theapemachine/amsh/messages"
)

/*
Update handles all incoming messages for the textarea component.
This method is part of the tea.Model interface and is responsible for updating the textarea state
based on various events such as key presses and window size changes.
It delegates to the underlying bubbles textarea for most of its functionality.
*/
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	EndSection := logger.StartSection("textarea.Update", "update")
	defer EndSection()

	logger.Debug("<- %T", msg)

	var (
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.state == components.Focused {
			m.Model.Update(msg)
		}
	case messages.Message[[]int]:
		if !messages.ShouldProcessMessage(m.state, msg.Context) {
			return m, nil
		}

		m.SetSize(msg.Data[0], msg.Data[1])
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) SetValue(val string) {
	m.state = components.Active
	m.Model.SetValue(val)
}

func (m *Model) Focus() {
	m.Model.Focus()
}

func (m *Model) SetContent(val string) {
	logger.Log("setting content:\n\n%s", val)
	m.Model.SetValue(val)
}

/*
SetSize adjusts the size of the textarea.
This method is crucial for responsive design, ensuring the textarea
adapts to window size changes and maintains a consistent appearance.
*/
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.Model.SetWidth(width / 2)
	m.Model.SetHeight(height)
}

func (m *Model) updateKeybindings() {
	m.event.Keymap.Add.SetEnabled(m.enabled)
	m.event.Keymap.Remove.SetEnabled(!m.enabled)
}

package editor

import (
	"bufio"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/theapemachine/amsh/components"
	"github.com/theapemachine/amsh/logger"
	"github.com/theapemachine/amsh/messages"
	"github.com/theapemachine/amsh/textarea"
)

/*
Update handles all incoming messages for the editor component.
This method is part of the tea.Model interface and is responsible for updating the editor state
based on various events such as key presses, file selection, and window size changes.
It delegates to specific handlers based on the current editing mode and message type.
*/
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	EndSection := logger.StartSection("editor.Update", "update")
	defer EndSection()

	logger.Debug("<- <%v>", msg)
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		m.inputs[m.focus].Update(msg)
	case messages.Message[[]int]:
		if !messages.ShouldProcessMessage(m.state, msg.Context) {
			return m, nil
		}

		m.handleWindowSizeMsg(msg)
	case messages.Message[string]:
		if !messages.ShouldProcessMessage(m.state, msg.Context) {
			return m, nil
		}

		if msg.Type == messages.MessageOpenFile {
			if m.err = m.loadFile(msg.Data); m.err != nil {
				logger.Log("Error opening file: %v", m.err)
				cmds = append(cmds, func() tea.Msg {
					return messages.NewMessage(
						messages.MessageError, m.err, messages.All,
					)
				})
			}
		}
	}

	return m, tea.Batch(cmds...)
}

/*
handleWindowSizeMsg handles window resizing messages.
*/
func (m *Model) handleWindowSizeMsg(msg messages.Message[[]int]) {
	m.width, m.height = msg.Data[0], msg.Data[1]
	m.resizeTextareas()
}

/*
loadFile loads a file into the editor, creating a new textarea for it.
*/
func (m *Model) loadFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	var content []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		content = append(content, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	if len(m.inputs) == 0 {
		m.focus = 0
		m.inputs = append(m.inputs, textarea.New(m.width, m.height))
	}

	m.inputs[m.focus].Focus()
	m.inputs[m.focus].SetContent(strings.Join(content, "\n"))

	m.state = components.Active
	return nil
}

/*
resizeTextareas resizes all textareas based on the current width and height.
*/
func (m *Model) resizeTextareas() {
	for _, input := range m.inputs {
		input.SetSize(m.width, m.height)
	}
}

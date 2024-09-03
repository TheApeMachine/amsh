package editor

import (
	tea "github.com/charmbracelet/bubbletea"
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
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		logger.Log("Editor received KeyMsg: %v", msg)
		switch m.mode {
		case NormalMode:
			return m.handleNormalModeKeyMsg(msg)
		case InsertMode:
			return m.handleInsertModeKeyMsg(msg)
		}
	case messages.SetFilenameMsg:
		logger.Log("Editor received SetFilenameMsg: %s", msg)
		return m, m.SetFile(string(msg))
	case tea.WindowSizeMsg:
		logger.Log("Editor received WindowSizeMsg: %v", msg)
		m.SetSize(msg.Width, msg.Height)
		return m, nil
	}

	return m, cmd
}

/*
handleNormalModeKeyMsg processes key messages when the editor is in Normal mode.
It handles navigation and mode switching based on the pressed key.
This method is crucial for implementing vim-like keybindings and navigation.
*/
func (m *Model) handleNormalModeKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "i":
		m.mode = InsertMode
		return m, m.sendStatusUpdate()
	case "j", "down":
		m.inputs[m.focus].CursorDown()
	case "k", "up":
		m.inputs[m.focus].CursorUp()
	}
	return m, nil
}

/*
handleInsertModeKeyMsg processes key messages when the editor is in Insert mode.
It handles text input and mode switching, updating the textarea content as necessary.
This method is essential for implementing text editing functionality.
*/
func (m *Model) handleInsertModeKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg.Type {
	case tea.KeyEsc:
		m.mode = NormalMode
		return m, m.sendStatusUpdate()
	default:
		newModel, newCmd := m.inputs[m.focus].Update(msg)
		if newTextarea, ok := newModel.(*textarea.Model); ok {
			m.inputs[m.focus] = newTextarea
			m.SetContent(m.inputs[m.focus].Value())
			cmd = newCmd
		}
	}
	return m, cmd
}

/*
sendStatusUpdate creates a command to send a status update message.
This method is used to notify other components about changes in the editor's state,
such as the current file and editing mode.
*/
func (m *Model) sendStatusUpdate() tea.Cmd {
	return func() tea.Msg {
		return messages.StatusUpdateMsg{
			Filename: m.filename,
			Mode:     messages.Mode(m.mode),
		}
	}
}

// OpenFileBrowserMsg is a message type used to trigger the file browser
type OpenFileBrowserMsg struct{}

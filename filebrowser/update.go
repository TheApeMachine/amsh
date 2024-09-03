package filebrowser

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/theapemachine/amsh/logger"
	"github.com/theapemachine/amsh/messages"
)

/*
handleEnterKey processes the action when the Enter key is pressed.
It either changes the current directory or selects a file, depending on the selected item.
This method is crucial for navigating the file system and selecting files for editing.
*/
func (m *Model) handleEnterKey() (tea.Model, tea.Cmd) {
	logger.Log("FileBrowser handling Enter key")
	if i, ok := m.list.SelectedItem().(fileItem); ok {
		if i.info.IsDir() {
			logger.Log("FileBrowser selected directory: %s", i.path)
			m.currentPath = i.path
			m.initList()
			return m, nil
		} else {
			logger.Log("FileBrowser selected file: %s", i.path)
			m.selectedFile = i.path
			return m, tea.Batch(
				func() tea.Msg {
					logger.Log("FileBrowser sending FileSelectedMsg: %s", m.selectedFile)
					return messages.FileSelectedMsg(m.selectedFile)
				},
				func() tea.Msg {
					logger.Log("FileBrowser sending SetActiveComponentMsg: editor")
					return messages.SetActiveComponentMsg("editor")
				},
			)
		}
	}
	logger.Log("FileBrowser handleEnterKey: no item selected")
	return m, nil
}

/*
Update handles all incoming messages for the file browser component.
This method is part of the tea.Model interface and is responsible for updating the file browser state
based on various events such as key presses and window size changes.
It delegates to specific handlers based on the message type and updates the list model as necessary.
*/
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "enter":
			return m.handleEnterKey()
		}
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		m.list.SetHeight(msg.Height - 1) // Leave room for status line
		return m, nil
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

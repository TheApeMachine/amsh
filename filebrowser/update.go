package filebrowser

import (
	tea "github.com/charmbracelet/bubbletea"
)

type OpenFileBrowserMsg tea.Cmd
type FileSelectedMsg tea.Msg

// Instead, create a helper function for the unique functionality
func (m *Model) handleEnterKey() (tea.Model, tea.Cmd) {
	if i, ok := m.list.SelectedItem().(fileItem); ok {
		if i.isDir {
			m.currentPath = i.path
			m.initList()
		} else {
			m.selectedFile = i.path
			return m, m.sendFileSelected()
		}
	}
	return m, nil
}

// This function can be called from the main Update method in model.go

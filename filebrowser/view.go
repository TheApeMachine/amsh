package filebrowser

import "fmt"

/*
View renders the current state of the file browser.
This method is part of the tea.Model interface and is responsible for generating
the visual representation of the file browser component.
It displays either the list of files/directories or an error message if there's an error.
The status line showing the current directory is always displayed at the bottom.
*/
func (m *Model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v", m.err)
	}
	return fmt.Sprintf("%s\n%s", m.list.View(), m.statusLine())
}

package buffer

import "strings"

func (m *Model) View() string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var views []string
	for _, view := range m.views {
		views = append(views, view)
	}

	return strings.Join(views, "\n\n")
}

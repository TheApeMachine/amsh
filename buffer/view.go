package buffer

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	builder = strings.Builder{}
)

/*
View renders the current state of the buffer.
The buffer's view is composed of any components that are currently registered, and rendering a view
with content. The buffer attempts to find the most optimal way to multiplex the views of these components.
*/
func (model *Model) View() string {
	builder.Reset()

	views := make([]string, 0, len(model.components))

	for _, component := range model.components {
		views = append(views, component.View())
	}

	builder.WriteString(lipgloss.JoinVertical(lipgloss.Top, views...))
	return builder.String()
}

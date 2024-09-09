package buffer

import (
	"strings"
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

	for _, component := range model.components {
		builder.WriteString(component.View())
	}

	return builder.String()
}

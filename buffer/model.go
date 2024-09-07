package buffer

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/theapemachine/amsh/messages"
)

var (
	width  int
	height int
)

/*
Model is the model for the buffer. It is responsible for managing the component state and views.
It acts as a central hub for all components, coordinating their interactions and rendering.
The use of a mutex ensures thread-safe access to the shared state, which is crucial for concurrent operations.
*/
type Model struct {
	components []tea.Model
	width      int
	height     int
	path       string
	err        error
}

/*
New creates a new buffer model.
It initializes the components map and sets the default active component to "filebrowser".
This factory function ensures that every new buffer instance starts with a consistent initial state.
*/
func New(path string, width, height int) *Model {
	m := &Model{
		components: make([]tea.Model, 0),
		path:       path,
		width:      width,
		height:     height,
	}

	return m
}

/*
Init initializes the buffer model. It initializes all components and returns a command to be executed.
*/
func (m *Model) Init() tea.Cmd {
	var cmds []tea.Cmd

	for _, component := range m.components {
		cmds = append(cmds, component.Init())
	}

	if m.path != "" {
		cmds = append(cmds, func() tea.Msg {
			return messages.NewMessage(
				messages.MessageOpenFile, m.path, messages.All,
			)
		})
	}

	return tea.Batch(cmds...)
}

/*
RegisterComponents registers one or more component with the buffer, which exposes the Update
method of the tea.Model interface that each component must implement.
*/
func (m *Model) RegisterComponents(name string, components ...tea.Model) {
	m.components = append(m.components, components...)
}

package buffer

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/theapemachine/amsh/messages"
	"github.com/theapemachine/amsh/ui"
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
	mode       ui.Mode
	err        error
}

/*
New creates a new buffer model.
It initializes the components map and sets the default active component to "filebrowser".
This factory function ensures that every new buffer instance starts with a consistent initial state.
*/
func New(path string, width, height int) *Model {
	return &Model{
		components: make([]tea.Model, 0),
		path:       path,
		width:      width,
		height:     height,
		mode:       ui.ModeNormal,
	}
}

/*
Init initializes the buffer model. It initializes all components and returns a command to be executed.
*/
func (model *Model) Init() tea.Cmd {
	var cmds []tea.Cmd

	for _, component := range model.components {
		cmds = append(cmds, component.Init())
	}

	if model.path != "" {
		cmds = append(cmds, func() tea.Msg {
			return messages.NewMessage(
				messages.MessageOpenFile, model.path, messages.All,
			)
		})
	}

	return tea.Batch(cmds...)
}

/*
RegisterComponents registers one or more component with the buffer, which exposes the Update
method of the tea.Model interface that each component must implement.
*/
func (model *Model) RegisterComponents(name string, components ...tea.Model) {
	model.components = append(model.components, components...)
}

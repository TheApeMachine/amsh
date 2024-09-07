package editor

import (
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/theapemachine/amsh/components"
	"github.com/theapemachine/amsh/logger"
	"github.com/theapemachine/amsh/textarea"
)

/*
Model represents the state of the editor component.
It manages the file content, cursor position, editing mode, and UI elements like viewport and textarea.
The model also supports multiple input areas, allowing for a flexible editing experience.
*/
type Model struct {
	files   []*os.File
	content []string
	width   int
	height  int
	inputs  []*textarea.Model
	focus   int
	state   components.State
	err     error
}

/*
New creates a new editor model with the given filename.
It initializes the viewport and textarea, setting up the initial state for editing.
This factory function ensures that every new editor instance starts with a consistent initial state.
*/
func New(width, height int) *Model {
	m := &Model{
		files:   make([]*os.File, 0),
		content: make([]string, 0),
		width:   width,
		height:  height,
		inputs:  make([]*textarea.Model, 0),
		focus:   0,
		state:   components.Inactive,
	}

	return m
}

/*
Init initializes the editor model.
This method is part of the tea.Model interface and is called when the editor component starts.
It sets up the initial state, such as starting the cursor blink in the focused input area.
*/
func (m *Model) Init() tea.Cmd {
	return nil
}

/*
SetSize adjusts the size of the editor components based on the given width and height.
This method is crucial for responsive design, ensuring that the editor layout adapts to window size changes.
It calculates appropriate dimensions for the viewport and textarea, accounting for borders and margins.
*/
func (m *Model) SetSize(width, height int) {
	logger.Log("editor.SetSize(%d, %d)", width, height)
	m.width = width
	m.height = height

	for _, input := range m.inputs {
		input.SetSize(width/len(m.inputs), height)
	}
}

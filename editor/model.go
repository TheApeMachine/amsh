package editor

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/theapemachine/amsh/textarea"
)

// Mode represents the current editing mode of the editor
type Mode int

const (
	NormalMode Mode = iota
	InsertMode
	initialInputs = 2
	maxInputs     = 6
	minInputs     = 1
	helpHeight    = 5
)

/*
Model represents the state of the editor component.
It manages the file content, cursor position, editing mode, and UI elements like viewport and textarea.
The model also supports multiple input areas, allowing for a flexible editing experience.
*/
type Model struct {
	filename string
	content  string
	cursor   int
	mode     Mode
	width    int
	height   int
	inputs   []*textarea.Model
	focus    int
}

/*
New creates a new editor model with the given filename.
It initializes the viewport and textarea, setting up the initial state for editing.
This factory function ensures that every new editor instance starts with a consistent initial state.
*/
func New(filename string) *Model {
	m := &Model{
		filename: filename,
		mode:     NormalMode,
		content:  "",
		cursor:   0,
		inputs:   []*textarea.Model{textarea.New()},
		focus:    0,
	}

	return m
}

/*
Init initializes the editor model.
This method is part of the tea.Model interface and is called when the editor component starts.
It sets up the initial state, such as starting the cursor blink in the focused input area.
*/
func (m *Model) Init() tea.Cmd {
	return m.inputs[m.focus].Blink()
}

/*
SetFile updates the editor with a new file.
It sets the filename and loads the file content (currently a placeholder).
This method is crucial for opening and editing different files within the same editor instance.
*/
func (m *Model) SetFile(filename string) tea.Cmd {
	m.filename = filename
	// Here you would typically load the file content
	// For now, we'll just set a placeholder content
	m.content = "Content of " + filename + "\nThis is a test content to ensure rendering works."
	m.SetContent(m.content)

	return m.sendStatusUpdate()
}

/*
SetContent updates the content of both the viewport and textarea.
This method ensures that the displayed content is synchronized across different view components.
*/
func (m *Model) SetContent(content string) {
	m.content = content
	m.inputs[m.focus].SetValue(content)
}

/*
SetSize adjusts the size of the editor components based on the given width and height.
This method is crucial for responsive design, ensuring that the editor layout adapts to window size changes.
It calculates appropriate dimensions for the viewport and textarea, accounting for borders and margins.
*/
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height

	// Adjust for borders and margins
	adjustedWidth := width - 6   // 2 for left margin, 2 for right margin, 2 for borders
	adjustedHeight := height - 5 // 2 for top/bottom margins, 2 for borders, 1 for status line

	viewportWidth := adjustedWidth / 2
	viewportHeight := adjustedHeight

	m.inputs[m.focus].SetSize(viewportWidth, viewportHeight)
	m.inputs[m.focus].SetValue(m.content)
}

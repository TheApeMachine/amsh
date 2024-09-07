package textarea

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/theapemachine/amsh/components"
	"github.com/theapemachine/amsh/logger"
	"github.com/theapemachine/amsh/ui"
)

/*
Model represents the state of the textarea component.
It wraps the bubbles textarea model, providing a consistent interface
for text input within the application.
*/
type Model struct {
	textarea.Model
	width   int
	height  int
	event   *ui.Event
	state   components.State
	enabled bool
}

/*
New creates a new textarea model with default settings.
This factory function ensures that every new textarea instance
starts with a consistent initial state and appearance.
*/
func New(width, height int) *Model {
	ta := textarea.New()
	ta.SetWidth(width)
	ta.SetHeight(height)
	ta.Prompt = ""
	ta.Placeholder = "Type something"
	ta.ShowLineNumbers = true
	ta.Cursor.Style = ui.CursorStyle
	ta.FocusedStyle.Placeholder = ui.FocussedPlaceholderStyle
	ta.BlurredStyle.Placeholder = ui.PlaceholderStyle
	ta.FocusedStyle.CursorLine = ui.CursorLineStyle
	ta.FocusedStyle.Base = ui.FocusedBorderStyle
	ta.BlurredStyle.Base = ui.BlurredBorderStyle
	ta.FocusedStyle.EndOfBuffer = ui.EndOfBufferStyle
	ta.BlurredStyle.EndOfBuffer = ui.EndOfBufferStyle
	ta.KeyMap.DeleteWordBackward.SetEnabled(false)
	ta.KeyMap.LineNext = key.NewBinding(key.WithKeys("down"))
	ta.KeyMap.LinePrevious = key.NewBinding(key.WithKeys("up"))
	ta.Focus()

	return &Model{
		Model:   ta,
		event:   ui.NewEvent(),
		state:   components.Inactive,
		enabled: true,
		width:   width,
		height:  height,
	}
}

/*
Init initializes the textarea model.
This method is part of the tea.Model interface and is called when the textarea component starts.
Currently, it doesn't perform any initialization actions, but it's included for consistency
and potential future use.
*/
func (m *Model) Init() tea.Cmd {
	logger.Log("textarea.Init()")
	return nil
}

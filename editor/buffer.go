package editor

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type keymap = struct {
	next, prev, add, remove, quit key.Binding
}

func newTextarea() textarea.Model {
	t := textarea.New()
	t.Prompt = ""
	t.Placeholder = "Type something"
	t.ShowLineNumbers = true
	t.Cursor.Style = cursorStyle
	t.FocusedStyle.Placeholder = focusedPlaceholderStyle
	t.BlurredStyle.Placeholder = placeholderStyle
	t.FocusedStyle.CursorLine = cursorLineStyle
	t.FocusedStyle.Base = focusedBorderStyle
	t.BlurredStyle.Base = blurredBorderStyle
	t.FocusedStyle.EndOfBuffer = endOfBufferStyle
	t.BlurredStyle.EndOfBuffer = endOfBufferStyle
	t.KeyMap.DeleteWordBackward.SetEnabled(false)
	t.KeyMap.LineNext = key.NewBinding(key.WithKeys("down"))
	t.KeyMap.LinePrevious = key.NewBinding(key.WithKeys("up"))
	t.Blur()
	return t
}

type Buffer struct {
	width       int
	height      int
	keyboardMgr *KeyboardManager
	help        help.Model
	components  []Component
	focus       int
}

func NewBuffer(path string) *Buffer {
	buffer := &Buffer{
		components:  make([]Component, initialInputs),
		help:        help.New(),
		keyboardMgr: NewKeyboardManager(),
	}

	for i := 0; i < initialInputs; i++ {
		textarea := newTextarea() // Assume newTextarea returns textarea.Model
		wrappedTextarea := &TextareaComponent{textarea: textarea}
		buffer.components[i] = wrappedTextarea
	}

	buffer.keyboardMgr.UpdateKeybindings(len(buffer.components))

	return buffer
}

func (buffer *Buffer) Init() tea.Cmd {
	return textarea.Blink
}

func (buffer *Buffer) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	for i, component := range buffer.components {
		newModel, cmd := component.Update(msg)
		buffer.components[i] = newModel.(Component)
		cmds = append(cmds, cmd)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, buffer.keyboardMgr.KeyMap.next):
			buffer.components[buffer.focus].Blur()
			buffer.focus = (buffer.focus + 1) % len(buffer.components)
			buffer.components[buffer.focus].Focus()
		case key.Matches(msg, buffer.keyboardMgr.KeyMap.prev):
			buffer.components[buffer.focus].Blur()
			buffer.focus--
			if buffer.focus < 0 {
				buffer.focus = len(buffer.components) - 1
			}
			buffer.components[buffer.focus].Focus()
		case key.Matches(msg, buffer.keyboardMgr.KeyMap.add):
			newComponent := &TextareaComponent{textarea: newTextarea()}
			buffer.components = append(buffer.components, newComponent)
			buffer.focus = len(buffer.components) - 1
			buffer.components[buffer.focus].Focus()
		case key.Matches(msg, buffer.keyboardMgr.KeyMap.remove):
			if len(buffer.components) > 1 {
				buffer.components = append(buffer.components[:buffer.focus], buffer.components[buffer.focus+1:]...)
				if buffer.focus >= len(buffer.components) {
					buffer.focus = len(buffer.components) - 1
				}
				buffer.components[buffer.focus].Focus()
			}
		case key.Matches(msg, buffer.keyboardMgr.KeyMap.quit):
			return buffer, tea.Quit
		}
	}

	return buffer, tea.Batch(cmds...)
}

func (buffer *Buffer) sizeInputs() {
	for i := range buffer.components {
		buffer.components[i].SetSize(
			buffer.width/len(buffer.components),
			buffer.height,
		)
	}
}

func (buffer *Buffer) View() string {
	help := buffer.help.ShortHelpView([]key.Binding{
		buffer.keyboardMgr.KeyMap.next,
		buffer.keyboardMgr.KeyMap.prev,
		buffer.keyboardMgr.KeyMap.add,
		buffer.keyboardMgr.KeyMap.remove,
		buffer.keyboardMgr.KeyMap.quit,
	})

	var views []string

	for i := range buffer.components {
		views = append(views, buffer.components[i].View())
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, views...) + "\n\n" + help
}

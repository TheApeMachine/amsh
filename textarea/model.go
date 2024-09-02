package textarea

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	textarea textarea.Model
	width    int
	height   int
}

func New() Model {
	t := textarea.New()
	t.Prompt = ""
	t.Placeholder = "Type something"
	t.ShowLineNumbers = true
	t.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
	t.FocusedStyle.Placeholder = lipgloss.NewStyle().Foreground(lipgloss.Color("99"))
	t.BlurredStyle.Placeholder = lipgloss.NewStyle().Foreground(lipgloss.Color("238"))
	t.FocusedStyle.CursorLine = lipgloss.NewStyle().Background(lipgloss.Color("57")).Foreground(lipgloss.Color("230"))
	t.FocusedStyle.Base = lipgloss.NewStyle()
	t.BlurredStyle.Base = lipgloss.NewStyle()
	t.FocusedStyle.EndOfBuffer = lipgloss.NewStyle().Foreground(lipgloss.Color("235"))
	t.BlurredStyle.EndOfBuffer = lipgloss.NewStyle().Foreground(lipgloss.Color("235"))
	t.KeyMap.DeleteWordBackward.SetEnabled(false)
	t.KeyMap.LineNext = key.NewBinding(key.WithKeys("down"))
	t.KeyMap.LinePrevious = key.NewBinding(key.WithKeys("up"))
	t.Blur()

	return Model{
		textarea: t,
		width:    80,
		height:   24,
	}
}

func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.textarea.SetWidth(width)
	m.textarea.SetHeight(height)
}

func (m *Model) Focus() {
	m.textarea.Focus()
}

func (m *Model) Blur() {
	m.textarea.Blur()
}

func (m *Model) SetValue(content string) {
	m.textarea.SetValue(content)
}

func (m *Model) Value() string {
	return m.textarea.Value()
}

// Move the style definitions from editor/style.go to textarea/style.go

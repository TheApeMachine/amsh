package features

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type keymap = struct {
	next, prev, add, remove, quit key.Binding
}

type TextArea struct {
	textarea.Model
	teleport *Teleport
}

func (t *TextArea) Init() tea.Cmd {
	return textarea.Blink
}

func (t *TextArea) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle teleport mode
		if t.teleport.IsActive() {
			// Get the first character of the key string
			keyStr := msg.String()
			if len(keyStr) > 0 {
				if cmd := t.teleport.AddInput(rune(keyStr[0])); cmd != nil {
					cmds = append(cmds, cmd)
					// Deactivate teleport after successful jump
					cmds = append(cmds, func() tea.Msg {
						t.teleport.Toggle()
						return nil
					})
				}
			}
			return t, tea.Batch(cmds...)
		}
		switch msg.String() {
		case "ctrl+e":
			return t, t.Model.Focus()
		case "ctrl+f":
			if cmd := t.teleport.Toggle(); cmd != nil {
				cmds = append(cmds, cmd)
			}
			return t, tea.Batch(cmds...)
		}
	case tea.WindowSizeMsg:
		t.Model.SetWidth(msg.Width)
		t.Model.SetHeight(msg.Height - 2)

	case TeleportMsg:
		// Handle cursor movement to the target position
		// You'll need to implement this based on your textarea's API
		// This might involve converting the line/col to the appropriate cursor position
		return t, nil
	}

	var cmd tea.Cmd
	t.Model, cmd = t.Model.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	return t, tea.Batch(cmds...)
}

func NewTextarea() *TextArea {
	ta := textarea.New()
	ta.Prompt = ""
	ta.Placeholder = "Type here..."
	ta.ShowLineNumbers = true
	ta.Cursor.Style = cursorStyle
	ta.FocusedStyle.Placeholder = focusedPlaceholderStyle
	ta.BlurredStyle.Placeholder = placeholderStyle
	ta.FocusedStyle.CursorLine = cursorLineStyle
	ta.FocusedStyle.Base = focusedBorderStyle
	ta.BlurredStyle.Base = blurredBorderStyle
	ta.FocusedStyle.EndOfBuffer = endOfBufferStyle
	ta.BlurredStyle.EndOfBuffer = endOfBufferStyle
	ta.KeyMap.DeleteWordBackward.SetEnabled(false)
	ta.KeyMap.LineNext = key.NewBinding(
		key.WithKeys("down"),
	)
	ta.KeyMap.LinePrevious = key.NewBinding(
		key.WithKeys("up"),
	)
	ta.Blur()
	return &TextArea{
		Model:    ta,
		teleport: NewTeleport(),
	}
}

var (
	cursorStyle = lipgloss.NewStyle().Foreground(
		lipgloss.Color("212"),
	)

	cursorLineStyle = lipgloss.NewStyle().Background(
		lipgloss.Color("57"),
	).Foreground(
		lipgloss.Color("230"),
	)

	placeholderStyle = lipgloss.NewStyle().Foreground(
		lipgloss.Color("238"),
	)

	endOfBufferStyle = lipgloss.NewStyle().Foreground(
		lipgloss.Color("235"),
	)

	focusedPlaceholderStyle = lipgloss.NewStyle().Foreground(
		lipgloss.Color("99"),
	)

	focusedBorderStyle = lipgloss.NewStyle().Border(
		lipgloss.RoundedBorder(),
	).BorderForeground(
		lipgloss.Color("238"),
	)

	blurredBorderStyle = lipgloss.NewStyle().Border(
		lipgloss.HiddenBorder(),
	)
)

func (t *TextArea) View() string {
	return lipgloss.JoinVertical(
		lipgloss.Left,
		t.Model.View(),
		t.teleport.View(),
	)
}

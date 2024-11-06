package features

import (
	"bufio"
	"os"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/theapemachine/amsh/errnie"
	"github.com/theapemachine/amsh/tui/components/textarea"
	"github.com/theapemachine/amsh/tui/types"
)

type TextArea struct {
	model          textarea.Model
	teleport       *Teleport
	selectionStart int
	selectionEnd   int
	selecting      bool
	mode           types.Mode
	chatWindow     *ChatWindow
	showChat       bool
}

func (ta *TextArea) Model() tea.Model {
	return ta
}

func (ta *TextArea) Name() string {
	return "textarea"
}

func (ta *TextArea) Init() tea.Cmd {
	return textarea.Blink
}

func (ta *TextArea) Size() (int, int) {
	return ta.model.Width(), ta.model.Height()
}

func (ta *TextArea) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Reserve space for status bar and borders
		reservedHeight := 4
		safeHeight := msg.Height - reservedHeight
		if safeHeight < 3 {
			safeHeight = 3
		}
		ta.model.SetWidth(msg.Width)
		ta.model.SetHeight(safeHeight)
		return ta, nil

	case LoadFileMsg:
		for scanner := bufio.NewScanner(errnie.SafeMust(func() (*os.File, error) {
			return os.Open(msg.Filepath)
		})); scanner.Scan(); {
			ta.model.InsertString(scanner.Text() + "\n")
		}
		ta.model.CursorStart()
		ta.model.Focus()
		return ta, nil

	case types.AIChunkMsg:
		ta.model.InsertString(msg.Chunk)
		return ta, nil

	case tea.KeyMsg:
		// Handle chat window escape first
		if ta.showChat && msg.Type == tea.KeyEsc {
			ta.showChat = false
			ta.chatWindow = nil
			return ta, nil
		}

		// Handle mode switching
		switch msg.String() {
		case "i":
			if ta.mode == types.ModeNormal {
				ta.mode = types.ModeInsert
				ta.model.SetBlockInput(false)
				return ta, func() tea.Msg {
					return ModeChangeMsg{Mode: types.ModeInsert}
				}
			}
		case "v":
			if ta.mode == types.ModeNormal {
				ta.mode = types.ModeVisual
				ta.model.StartSelection()
				return ta, func() tea.Msg {
					return ModeChangeMsg{Mode: types.ModeVisual}
				}
			}
		case "esc":
			switch ta.mode {
			case types.ModeInsert:
				ta.mode = types.ModeNormal
				ta.model.SetBlockInput(true)
				return ta, func() tea.Msg {
					return ModeChangeMsg{Mode: types.ModeNormal}
				}
			case types.ModeVisual:
				ta.mode = types.ModeNormal
				ta.model.EndSelection()
				return ta, func() tea.Msg {
					return ModeChangeMsg{Mode: types.ModeNormal}
				}
			}
		}

		// Handle chat window input if active
		if ta.showChat {
			var cmd tea.Cmd
			model, cmd := ta.chatWindow.Update(msg)

			// Type assert the returned model back to *ChatWindow
			if chatWindow, ok := model.(*ChatWindow); ok {
				ta.chatWindow = chatWindow
				return ta, cmd
			}

			// If type assertion fails, something went wrong
			return ta, nil
		}

		// Handle teleport mode
		if ta.teleport.IsActive() {
			var cmds []tea.Cmd
			keyStr := msg.String()
			if len(keyStr) > 0 {
				if cmd := ta.teleport.AddInput(rune(keyStr[0])); cmd != nil {
					cmds = append(cmds, cmd)
					cmds = append(cmds, func() tea.Msg {
						ta.teleport.Toggle()
						return nil
					})
				}
			}
			return ta, tea.Batch(cmds...)
		}

		// Handle navigation and update selection in visual mode
		if ta.mode == types.ModeVisual {
			var cmd tea.Cmd
			ta.model, cmd = ta.model.Update(msg)
			ta.model.UpdateSelection()
			return ta, cmd
		}

		// Handle other modes
		if ta.mode == types.ModeNormal || ta.mode == types.ModeInsert {
			var cmd tea.Cmd
			ta.model, cmd = ta.model.Update(msg)
			return ta, cmd
		}

	case types.OpenChatMsg:
		errnie.Log("OpenChatMsg", msg.Context)
		// Create and show chat window
		ta.chatWindow = NewChatWindow(msg.Context)
		ta.showChat = true
		return ta, nil
	}

	// Handle other message types
	var cmd tea.Cmd
	ta.model, cmd = ta.model.Update(msg)
	return ta, cmd
}

func NewTextarea() *TextArea {
	ta := textarea.New()
	ta.CharLimit = 1000000
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
		model:          ta,
		teleport:       NewTeleport(),
		selectionStart: 0,
		selectionEnd:   0,
		selecting:      false,
		mode:           types.ModeNormal,
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
	// If chat is active, show the chat window instead of textarea
	if t.showChat && t.chatWindow != nil {
		return t.chatWindow.View()
	}

	// Only show teleport view when it's active to save vertical space
	if t.teleport.IsActive() {
		return lipgloss.JoinVertical(
			lipgloss.Left,
			t.model.View(),
			t.teleport.View(),
		)
	}
	return t.model.View()
}

type LoadFileMsg struct {
	Filepath string
}

func (t *TextArea) GetHighlightedText() string {
	if !t.selecting || t.selectionStart == t.selectionEnd {
		return ""
	}

	start := t.selectionStart
	end := t.selectionEnd

	// Ensure start is before end
	if start > end {
		start, end = end, start
	}

	// Get the full text and extract the selection
	fullText := t.model.Value()
	if end > len(fullText) {
		end = len(fullText)
	}

	return fullText[start:end]
}

func (textarea *TextArea) StartSelection() {
	textarea.selecting = true
	textarea.selectionStart = textarea.model.LineInfo().CharOffset
	textarea.selectionEnd = textarea.selectionStart
}

func (textarea *TextArea) EndSelection() {
	textarea.selecting = false
}

func (textarea *TextArea) UpdateSelection() {
	if textarea.selecting {
		textarea.selectionEnd = textarea.model.LineInfo().CharOffset
	}
}

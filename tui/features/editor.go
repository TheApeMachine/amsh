package features

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/theapemachine/amsh/ai/provider"
	"github.com/theapemachine/amsh/ai/system"
	"github.com/theapemachine/amsh/errnie"
	"github.com/theapemachine/amsh/tui/components/overlay"
	"github.com/theapemachine/amsh/tui/components/textarea"
	"github.com/theapemachine/amsh/tui/types"
	"github.com/theapemachine/amsh/utils"
	"github.com/zyedidia/highlight"
)

type Editor struct {
	model          textarea.Model
	teleport       *Teleport
	selectionStart int
	selectionEnd   int
	selecting      bool
	mode           types.Mode
	chatWindow     *ChatWindow
	showChat       bool
}

func (editor *Editor) Model() tea.Model {
	return editor
}

func (editor *Editor) Name() string {
	return "textarea"
}

func (editor *Editor) Init() tea.Cmd {
	cmds := []tea.Cmd{textarea.Blink}
	cmds = append(cmds, func() tea.Msg {
		return ModeChangeMsg{Mode: types.ModeNormal}
	})

	editor.model.SetBlockInput(true)

	return tea.Batch(cmds...)
}

func (editor *Editor) Size() (int, int) {
	return editor.model.Width(), editor.model.Height()
}

func (editor *Editor) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	errnie.Log("editor.Update %v", msg)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Reserve space for status bar and borders
		reservedHeight := 4
		safeHeight := msg.Height - reservedHeight
		if safeHeight < 3 {
			safeHeight = 3
		}
		editor.model.SetWidth(msg.Width)
		editor.model.SetHeight(safeHeight)

	case LoadFileMsg:
		file := errnie.SafeMust(func() (*os.File, error) {
			return os.Open(msg.Filepath)
		})
		defer file.Close()

		scanner := bufio.NewScanner(file)

		// Get working directory for syntax file path
		wd, err := os.Getwd()
		if err != nil {
			errnie.Error(err)
			return editor, nil
		}

		// Load and parse syntax definition
		syntaxFile, err := os.ReadFile(filepath.Join(wd, "tui/syntax/go.yaml"))
		if err != nil {
			errnie.Error(err)
			return editor, nil
		}

		syntaxDef, err := highlight.ParseDef(syntaxFile)
		if err != nil {
			errnie.Error(err)
			return editor, nil
		}

		h := highlight.NewHighlighter(syntaxDef)
		var builder strings.Builder

		for scanner.Scan() {
			line := scanner.Text()
			matches := h.HighlightString(line)

			var styledLine string
			for colN, char := range line {
				if group, ok := matches[0][colN]; ok {
					switch group {
					case highlight.Groups["comment"]:
						styledLine += lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Render(string(char))
					case highlight.Groups["keyword"]:
						styledLine += lipgloss.NewStyle().Foreground(lipgloss.Color("5")).Render(string(char))
					case highlight.Groups["string"]:
						styledLine += lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Render(string(char))
					default:
						styledLine += string(char)
					}
				} else {
					styledLine += string(char)
				}
			}
			builder.WriteString(styledLine + "\n")
		}

		errnie.Log("styled content sample: %s", builder.String()[:min(100, builder.Len())])

		editor.model.SetValue(builder.String())
		editor.model.CursorStart()
		editor.model.Focus()
		return editor, nil

	case types.AIPromptMsg:
		editor.handleAI(msg.Prompt)

	case types.AIChunkMsg:
		editor.model.InsertString(msg.Chunk)

	case tea.KeyMsg:
		// Handle chat window escape first
		if editor.showChat && msg.Type == tea.KeyEsc {
			editor.showChat = false
			editor.chatWindow = nil
			return editor, nil
		}

		// Handle mode switching
		switch msg.String() {
		case "i":
			if editor.mode == types.ModeNormal {
				editor.mode = types.ModeInsert
				editor.model.SetBlockInput(false)
				return editor, func() tea.Msg {
					return ModeChangeMsg{Mode: types.ModeInsert}
				}
			}
		case "v":
			if editor.mode == types.ModeNormal {
				editor.mode = types.ModeVisual
				editor.model.StartSelection()
				return editor, func() tea.Msg {
					return ModeChangeMsg{Mode: types.ModeVisual}
				}
			}
		case "esc":
			switch editor.mode {
			case types.ModeInsert:
				editor.mode = types.ModeNormal
				editor.model.SetBlockInput(true)
				return editor, func() tea.Msg {
					return ModeChangeMsg{Mode: types.ModeNormal}
				}
			case types.ModeVisual:
				editor.mode = types.ModeNormal
				editor.model.EndSelection()
				return editor, func() tea.Msg {
					return ModeChangeMsg{Mode: types.ModeNormal}
				}
			}
		}

		// Handle chat window input if active
		if editor.showChat {
			var cmd tea.Cmd
			model, cmd := editor.chatWindow.Update(msg)

			// Type assert the returned model back to *ChatWindow
			if chatWindow, ok := model.(*ChatWindow); ok {
				editor.chatWindow = chatWindow
				return editor, cmd
			}

			// If type assertion fails, something went wrong
			return editor, nil
		}

		// Handle teleport mode
		if editor.teleport.IsActive() {
			var cmds []tea.Cmd
			keyStr := msg.String()
			if len(keyStr) > 0 {
				if cmd := editor.teleport.AddInput(rune(keyStr[0])); cmd != nil {
					cmds = append(cmds, cmd)
					cmds = append(cmds, func() tea.Msg {
						editor.teleport.Toggle()
						return nil
					})
				}
			}
			return editor, tea.Batch(cmds...)
		}

		// Handle navigation and update selection in visual mode
		if editor.mode == types.ModeVisual {
			var cmd tea.Cmd
			editor.model, cmd = editor.model.Update(msg)
			editor.UpdateSelection()
			return editor, cmd
		}

		// Handle other modes
		if editor.mode == types.ModeNormal || editor.mode == types.ModeInsert {
			var cmd tea.Cmd
			editor.model, cmd = editor.model.Update(msg)
			return editor, cmd
		}

	case types.OpenChatMsg:
		errnie.Log("OpenChatMsg %s", msg.Context)
		// Create and show chat window
		editor.chatWindow = NewChatWindow(msg.Context)
		editor.showChat = true
		return editor, nil
	}

	// Handle other message types
	var cmd tea.Cmd
	editor.model, cmd = editor.model.Update(msg)
	return editor, cmd
}

func NewTextarea() *Editor {
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
	return &Editor{
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

func (editor *Editor) View() string {
	// If chat is active, show the chat window instead of textarea
	if editor.showChat && editor.chatWindow != nil {
		return overlay.PlaceOverlay(
			editor.model.LineInfo().CharOffset,
			editor.model.LineInfo().RowOffset,
			editor.chatWindow.View(),
			editor.model.View(),
			true,
		)
	}

	// Only show teleport view when it's active to save vertical space
	if editor.teleport.IsActive() {
		return lipgloss.JoinVertical(
			lipgloss.Left,
			editor.model.View(),
			editor.teleport.View(),
		)
	}
	return editor.model.View()
}

type LoadFileMsg struct {
	Filepath string
}

func (editor *Editor) GetHighlightedText() string {
	if !editor.selecting || editor.selectionStart == editor.selectionEnd {
		return ""
	}

	start := editor.selectionStart
	end := editor.selectionEnd

	// Ensure start is before end
	if start > end {
		start, end = end, start
	}

	// Get the full text and extract the selection
	fullText := editor.model.Value()
	if end > len(fullText) {
		end = len(fullText)
	}

	return fullText[start:end]
}

func (editor *Editor) StartSelection() {
	editor.selecting = true
	editor.selectionStart = editor.model.LineInfo().CharOffset
	editor.selectionEnd = editor.selectionStart
}

func (editor *Editor) EndSelection() {
	editor.selecting = false
}

func (editor *Editor) UpdateSelection() {
	if editor.selecting {
		editor.selectionEnd = editor.model.LineInfo().CharOffset
	}
}

func (editor *Editor) handleAI(prompt string) {
	go func() {
		for event := range system.NewProcessManager("marvin", "editor").Execute(
			utils.JoinWith(
				"\n\n", editor.model.Value(), prompt,
			),
		) {
			if event.Type == provider.EventToken {
				editor.Update(types.AIChunkMsg{Chunk: event.Content})
			}
		}
	}()
}

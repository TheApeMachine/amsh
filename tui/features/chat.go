package features

import (
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/theapemachine/amsh/tui/types"
	"github.com/theapemachine/errnie"
)

type ChatWindow struct {
	textarea textarea.Model
	prompt   string
	context  string
	width    int
	height   int
	response string
	focused  bool
}

func NewChatWindow(highlightedText string) *ChatWindow {
	ta := textarea.New()
	ta.Placeholder = "Enter your prompt..."
	ta.ShowLineNumbers = false
	ta.Focus()

	return &ChatWindow{
		textarea: ta,
		prompt:   "",
		width:    0,
		height:   0,
		context:  highlightedText,
		response: "",
		focused:  true,
	}
}

func (chat *ChatWindow) Model() tea.Model {
	return chat
}

func (chat *ChatWindow) Name() string {
	return "chat"
}

func (chat *ChatWindow) Size() (int, int) {
	return chat.width, chat.height
}

func (chat *ChatWindow) Init() tea.Cmd {
	return textarea.Blink
}

func (chat *ChatWindow) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	errnie.Log("chat.Update %v", msg)

	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		chat.height = msg.Height
		chat.width = msg.Width
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			chat.focused = false
			cmds = append(cmds, func() tea.Msg {
				return tea.KeyMsg{Type: tea.KeyEsc}
			})
		case tea.KeyEnter:
			chat.focused = false
			cmds = append(cmds, func() tea.Msg {
				return types.AIPromptMsg{Prompt: chat.textarea.Value()}
			})

			cmds = append(cmds, func() tea.Msg {
				return tea.KeyMsg{Type: tea.KeyEsc}
			})
		}
	}

	var cmd tea.Cmd
	chat.textarea, cmd = chat.textarea.Update(msg)
	cmds = append(cmds, cmd)

	return chat, tea.Batch(cmds...)
}

func (chat *ChatWindow) View() string {
	if !chat.focused {
		return ""
	}

	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("238")).
		Padding(1)

	return style.Height(chat.height - 2).Render(chat.textarea.View())
}

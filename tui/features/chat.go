package features

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/theapemachine/amsh/ai/provider"
	"github.com/theapemachine/amsh/ai/system"
	"github.com/theapemachine/amsh/tui/components/textarea"
	"github.com/theapemachine/amsh/tui/types"
	"github.com/theapemachine/amsh/utils"
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
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		chat.height = msg.Height
		chat.width = msg.Width

	case types.AISendMsg:
		chat.handleAI()
		chat.focused = false
		return chat, func() tea.Msg {
			return tea.KeyMsg{Type: tea.KeyEsc}
		}

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			chat.focused = false
			return chat, func() tea.Msg {
				return tea.KeyMsg{Type: tea.KeyEsc}
			}
		case tea.KeyEnter:
			chat.handleAI()
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

func (chat *ChatWindow) handleAI() {
	go func() {
		for event := range system.NewProcessManager("marvin", "editor").Execute(
			utils.JoinWith(
				"\n\n", chat.context, chat.textarea.Value(),
			),
		) {
			if event.Type == provider.EventToken {
				chat.response += event.Content
				chat.Update(types.AIChunkMsg{Chunk: event.Content})
			}
		}
	}()
}

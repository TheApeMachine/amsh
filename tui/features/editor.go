package features

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	initialInputs = 2
	maxInputs     = 6
	minInputs     = 1
	helpHeight    = 5
)

var (
	cursorStyle             = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
	cursorLineStyle         = lipgloss.NewStyle().Background(lipgloss.Color("57")).Foreground(lipgloss.Color("230"))
	placeholderStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("238"))
	endOfBufferStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("235"))
	focusedPlaceholderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("99"))
	focusedBorderStyle      = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("238"))
	blurredBorderStyle      = lipgloss.NewStyle().Border(lipgloss.HiddenBorder())
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

type Editor struct {
	width  int
	height int
	keymap keymap
	help   help.Model
	inputs []textarea.Model
	focus  int
}

func NewEditor() *Editor {
	m := Editor{
		inputs: make([]textarea.Model, initialInputs),
		help:   help.New(),
		keymap: keymap{
			next: key.NewBinding(
				key.WithKeys("tab"),
				key.WithHelp("tab", "next"),
			),
			prev: key.NewBinding(
				key.WithKeys("shift+tab"),
				key.WithHelp("shift+tab", "prev"),
			),
			add: key.NewBinding(
				key.WithKeys("ctrl+n"),
				key.WithHelp("ctrl+n", "add an editor"),
			),
			remove: key.NewBinding(
				key.WithKeys("ctrl+w"),
				key.WithHelp("ctrl+w", "remove an editor"),
			),
			quit: key.NewBinding(
				key.WithKeys("esc", "ctrl+c"),
				key.WithHelp("esc", "quit"),
			),
		},
	}
	for i := 0; i < initialInputs; i++ {
		m.inputs[i] = newTextarea()
	}
	m.inputs[m.focus].Focus()
	m.updateKeybindings()
	return &m
}

func (editor *Editor) Init() tea.Cmd {
	return nil
}

func (editor *Editor) Model() tea.Model {
	return editor
}

func (editor *Editor) Name() string {
	return "editor"
}

func (editor *Editor) Size() (int, int) {
	return editor.width, editor.height
}

func (editor *Editor) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, editor.keymap.quit):
			for i := range editor.inputs {
				editor.inputs[i].Blur()
			}
			return editor, tea.Quit
		case key.Matches(msg, editor.keymap.next):
			editor.inputs[editor.focus].Blur()
			editor.focus++
			if editor.focus > len(editor.inputs)-1 {
				editor.focus = 0
			}
			cmd := editor.inputs[editor.focus].Focus()
			cmds = append(cmds, cmd)
		case key.Matches(msg, editor.keymap.prev):
			editor.inputs[editor.focus].Blur()
			editor.focus--
			if editor.focus < 0 {
				editor.focus = len(editor.inputs) - 1
			}
			cmd := editor.inputs[editor.focus].Focus()
			cmds = append(cmds, cmd)
		case key.Matches(msg, editor.keymap.add):
			editor.inputs = append(editor.inputs, newTextarea())
		case key.Matches(msg, editor.keymap.remove):
			editor.inputs = editor.inputs[:len(editor.inputs)-1]
			if editor.focus > len(editor.inputs)-1 {
				editor.focus = len(editor.inputs) - 1
			}
		}
	case tea.WindowSizeMsg:
		editor.height = msg.Height
		editor.width = msg.Width
		// case types.LoadFileMsg:
		// 	fh := errnie.SafeMust(func() (*os.File, error) {
		// 		return os.Open(msg.Filepath)
		// 	})
		// 	defer fh.Close()
	}

	editor.updateKeybindings()
	editor.sizeInputs()

	// Update all textareas
	for i := range editor.inputs {
		newModel, cmd := editor.inputs[i].Update(msg)
		editor.inputs[i] = newModel
		cmds = append(cmds, cmd)
	}

	return editor, tea.Batch(cmds...)
}

func (editor *Editor) sizeInputs() {
	for i := range editor.inputs {
		editor.inputs[i].SetWidth(editor.width / len(editor.inputs))
		editor.inputs[i].SetHeight(editor.height - helpHeight)
	}
}

func (editor *Editor) updateKeybindings() {
	editor.keymap.add.SetEnabled(len(editor.inputs) < maxInputs)
	editor.keymap.remove.SetEnabled(len(editor.inputs) > minInputs)
}

func (editor *Editor) View() string {
	help := editor.help.ShortHelpView([]key.Binding{
		editor.keymap.next,
		editor.keymap.prev,
		editor.keymap.add,
		editor.keymap.remove,
		editor.keymap.quit,
	})

	var views []string
	for i := range editor.inputs {
		views = append(views, editor.inputs[i].View())
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, views...) + "\n\n" + help
}

package tui

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/theapemachine/amsh/tui/features"
)

const (
	width         = 96
	height        = 24
	columnWidth   = 30
	initialInputs = 2
	maxInputs     = 6
	minInputs     = 1
	helpHeight    = 5
)

type keymap = struct {
	next, prev, add, remove, quit key.Binding
}

type textArea struct {
	textarea.Model
}

func (t textArea) Init() tea.Cmd {
	return textarea.Blink
}

func (t textArea) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	t.Model, cmd = t.Model.Update(msg)
	return t, cmd
}

func makeTextarea() textArea {
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
	return textArea{ta}
}

type App struct {
	width   int
	height  int
	keymap  keymap
	help    *help.Model
	screens map[string][]tea.Model
	screen  string
	inputs  []textArea
	focus   int
}

func NewApp() *App {
	return &App{
		screens: map[string][]tea.Model{
			"splash":  {features.NewSplash(width, height)},
			"editor":  {makeTextarea()},
			"browser": {features.NewBrowser()},
		},
		screen: "splash",
	}
}

func (app *App) Init() tea.Cmd {
	return textarea.Blink
}

func (app *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	for _, model := range app.screens[app.screen] {
		model, _ = model.Update(msg)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return app, tea.Quit
		case "ctrl+b":
			app.screen = "browser"
		case "ctrl+e":
			app.screen = "editor"
		}
	case tea.WindowSizeMsg:
		app.width, app.height = msg.Width, msg.Height
	}

	app.updateKeybindings()
	return app, nil
}

func (app *App) sizeInputs() {
	for i := range app.inputs {
		app.inputs[i].SetWidth(app.width / len(app.inputs))
		app.inputs[i].SetHeight(app.height / helpHeight)
	}
}

func (app *App) updateKeybindings() {
	app.keymap.add.SetEnabled(len(app.inputs) < maxInputs)
	app.keymap.remove.SetEnabled(len(app.inputs) > minInputs)
}

func (app *App) View() string {
	return app.screens[app.screen][0].View()
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

package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

var (
	red    = lipgloss.AdaptiveColor{Light: "#FE5F86", Dark: "#FE5F86"}
	indigo = lipgloss.AdaptiveColor{Light: "#5A56E0", Dark: "#7571F9"}
	green  = lipgloss.AdaptiveColor{Light: "#02BA84", Dark: "#02BF87"}
)

type Styles struct {
	Base,
	HeaderText,
	Status,
	StatusHeader,
	Highlight,
	ErrorHeaderText,
	Help lipgloss.Style
}

func NewStyles(lg *lipgloss.Renderer) *Styles {
	s := Styles{}
	s.Base = lg.NewStyle().
		Padding(1, 4, 0, 1)
	s.HeaderText = lg.NewStyle().
		Foreground(indigo).
		Bold(true).
		Padding(0, 1, 0, 2)
	s.Status = lg.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(indigo).
		PaddingLeft(1).
		MarginTop(1)
	s.StatusHeader = lg.NewStyle().
		Foreground(green).
		Bold(true)
	s.Highlight = lg.NewStyle().
		Foreground(lipgloss.Color("212"))
	s.ErrorHeaderText = s.HeaderText.Copy().
		Foreground(red)
	s.Help = lg.NewStyle().
		Foreground(lipgloss.Color("240"))
	return &s
}

var command string

var cmdline = huh.NewForm(huh.NewGroup(
	huh.NewInput().Prompt("? ").Value(&command),
))

var miniscript = huh.NewForm(huh.NewGroup(
	huh.NewText().Title("READY").Value(&command),
))

type State struct {
	command *string
	form    *huh.Form
	lg      *lipgloss.Renderer
	width   int
	height  int
	styles  *Styles
}

func NewState() *State {
	lg := lipgloss.DefaultRenderer()

	return &State{
		command: &command,
		form:    cmdline,
		lg:      lg,
		width:   640,
		height:  480,
		styles:  NewStyles(lg),
	}
}

func (state *State) Init() tea.Cmd {
	return state.form.Init()
}

func (state *State) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return state, tea.Quit
		case "tab":
			state.form = miniscript
			state.Init()
		}
	}

	var cmd tea.Cmd

	form, cmd := state.form.Update(msg)

	if f, ok := form.(*huh.Form); ok {
		state.form = f
		cmds = append(cmds, cmd)
	}

	return state, tea.Batch(cmds...)
}

func (state *State) View() string {
	s := state.styles
	v := strings.TrimSuffix(state.form.View(), "\n\n")

	form := state.lg.NewStyle().Margin(1, 0).Render(v)
	status := ""

	header := state.appBoundaryView("amsh")
	body := lipgloss.JoinHorizontal(lipgloss.Top, form, status)
	footer := state.appBoundaryView(state.form.Help().ShortHelpView(state.form.KeyBinds()))

	return s.Base.Render(header + "\n" + body + "\n\n" + footer)
}

func (state *State) appBoundaryView(text string) string {
	return lipgloss.PlaceHorizontal(
		state.width,
		lipgloss.Left,
		state.styles.HeaderText.Render(text),
		lipgloss.WithWhitespaceChars("|"),
		lipgloss.WithWhitespaceForeground(indigo),
	)
}

func (state *State) appErrorBoundaryView(text string) string {
	return lipgloss.PlaceHorizontal(
		state.width,
		lipgloss.Left,
		state.styles.ErrorHeaderText.Render(text),
		lipgloss.WithWhitespaceChars("/"),
		lipgloss.WithWhitespaceForeground(red),
	)
}

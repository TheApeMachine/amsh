package screens

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

type state int

const (
	statusNormal state = iota
	stateDone
)

type Start struct {
	state    state
	lg       *lipgloss.Renderer
	styles   *Styles
	form     *huh.Form
	width    int
	hasFocus bool
}

func NewStart() *Start {
	renderer := lipgloss.DefaultRenderer()

	return &Start{
		lg:     renderer,
		styles: NewStyles(renderer),
		form: huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Key("context").
					Options(huh.NewOptions("Tools", "Editor")...).
					Title("Start").
					Description("Select option"),
			),
		).WithWidth(45).WithShowHelp(false).WithShowErrors(false),
	}
}

func (start *Start) Init() tea.Cmd {
	return start.form.Init()
}

func (start *Start) Show() {}

func (start *Start) Focus() {
	start.hasFocus = true
}

func (start *Start) Blur() {
	start.hasFocus = false
}

func (start *Start) Update(msg tea.Msg) (*Start, tea.Cmd) {
	var cmds []tea.Cmd
	form, cmd := start.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		start.form = f
		cmds = append(cmds, cmd)
	}

	return start, tea.Batch(cmds...)
}

func (start *Start) View() string {
	if !start.hasFocus {
		return ""
	}

	s := start.styles

	switch start.form.State {
	default:
		v := strings.TrimSuffix(start.form.View(), "\n\n")
		form := start.lg.NewStyle().Margin(1, 0).Render(v)

		errors := start.form.Errors()
		header := start.appBoundaryView("amsh")
		if len(errors) > 0 {
			header = start.appErrorBoundaryView(start.errorView())
		}
		body := lipgloss.JoinHorizontal(lipgloss.Top, form)

		footer := start.appBoundaryView(start.form.Help().ShortHelpView(start.form.KeyBinds()))
		if len(errors) > 0 {
			footer = start.appErrorBoundaryView("")
		}

		return s.Base.Render(header + "\n" + body + "\n\n" + footer)
	}
}

func (start *Start) errorView() string {
	var s string
	for _, err := range start.form.Errors() {
		s += err.Error()
	}
	return s
}

func (start *Start) appBoundaryView(text string) string {
	return lipgloss.PlaceHorizontal(
		start.width,
		lipgloss.Left,
		start.styles.HeaderText.Render(text),
		lipgloss.WithWhitespaceChars("/"),
		lipgloss.WithWhitespaceForeground(indigo),
	)
}

func (start *Start) appErrorBoundaryView(text string) string {
	return lipgloss.PlaceHorizontal(
		start.width,
		lipgloss.Left,
		start.styles.ErrorHeaderText.Render(text),
		lipgloss.WithWhitespaceChars("/"),
		lipgloss.WithWhitespaceForeground(red),
	)
}

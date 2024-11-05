package features

import (
	"image/color"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lucasb-eyer/go-colorful"
	"github.com/muesli/gamut"
)

type Splash struct {
	width  int
	height int
}

func NewSplash(width, height int) Splash {
	return Splash{width, height}
}

func (s Splash) Init() tea.Cmd {
	return nil
}

func (s Splash) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	return s, nil
}

func (s Splash) View() string {
	doc := strings.Builder{}
	question := lipgloss.NewStyle().Width(50).Align(lipgloss.Center).Render(
		rainbow(lipgloss.NewStyle(), "Are you sure you want to eat marmalade?", blends),
	)
	ui := lipgloss.JoinVertical(lipgloss.Center, question)

	dialog := lipgloss.Place(s.width, 9,
		lipgloss.Center, lipgloss.Center,
		dialogBoxStyle.Render(ui),
		lipgloss.WithWhitespaceChars("猫咪"),
		lipgloss.WithWhitespaceForeground(subtle),
	)

	doc.WriteString(dialog + "\n\n")
	return doc.String()
}

func rainbow(base lipgloss.Style, s string, colors []color.Color) string {
	var str string
	for i, ss := range s {
		color, _ := colorful.MakeColor(colors[i%len(colors)])
		str = str + base.Foreground(lipgloss.Color(color.Hex())).Render(string(ss))
	}
	return str
}

var subtle = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
var blends = gamut.Blends(lipgloss.Color("#F25D94"), lipgloss.Color("#EDFF82"), 50)

var dialogBoxStyle = lipgloss.NewStyle().Border(
	lipgloss.RoundedBorder(),
).BorderForeground(
	lipgloss.Color("#874BFD"),
).Padding(1, 0).BorderTop(true).BorderLeft(true).BorderRight(true).BorderBottom(true)

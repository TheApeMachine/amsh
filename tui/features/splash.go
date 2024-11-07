package features

import (
	"image/color"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lucasb-eyer/go-colorful"
	"github.com/muesli/gamut"
)

type Splash struct {
	width  int
	height int
	logo   string
}

func NewSplash(width, height int) *Splash {
	logo, err := os.ReadFile("tui/logo.ans")
	if err != nil {
		return &Splash{width: width, height: height}
	}
	return &Splash{
		width:  width,
		height: height,
		logo:   string(logo),
	}
}

func (splash *Splash) Model() tea.Model {
	return splash
}

func (splash *Splash) Name() string {
	return "splash"
}

func (splash *Splash) Size() (int, int) {
	return splash.width, splash.height
}

func (splash *Splash) SetSize(width, height int) {
	splash.width = width
	splash.height = height
}

func (splash *Splash) Init() tea.Cmd {
	return nil
}

func (splash *Splash) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	return splash, nil
}

func (splash *Splash) View() string {
	doc := strings.Builder{}

	// Ensure we don't exceed maximum width (252 chars)
	safeWidth := min(splash.width, 182)

	// Subtract border and padding to prevent overflow
	contentWidth := safeWidth - 6 // 2 for borders, 4 for padding

	// Create the logo section
	logoSection := lipgloss.NewStyle().
		Width(contentWidth).
		Align(lipgloss.Center).
		Render(splash.logo)

	// Create the question section
	question := lipgloss.NewStyle().
		Width(contentWidth).
		Align(lipgloss.Center).
		Render(rainbow(lipgloss.NewStyle(), "Ape Machine Shell", blends))

	// Join the logo and question vertically
	ui := lipgloss.JoinVertical(lipgloss.Center,
		logoSection,
		question,
	)

	dialog := lipgloss.Place(safeWidth, splash.height,
		lipgloss.Center, lipgloss.Center,
		dialogBoxStyle.Render(ui),
		lipgloss.WithWhitespaceChars("猫咪"),
		lipgloss.WithWhitespaceForeground(subtle),
	)

	doc.WriteString(dialog)
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

var dialogBoxStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("#874BFD")).
	Padding(1, 0).
	BorderTop(true).
	BorderLeft(true).
	BorderRight(true).
	BorderBottom(true)

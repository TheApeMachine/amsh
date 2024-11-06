package features

import (
	"math/rand"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/gamut"
)

type Splash struct {
	width    int
	height   int
	logo     string
	frame    int
	cells    [][]Cell
	complete bool
	lastTick time.Time
}

// Cell represents a single character in the animation
type Cell struct {
	char     string
	revealed bool
	colorIdx int
	settled  bool
}

// tickMsg is sent when the timer ticks
type tickMsg time.Time

var rainbowColors = []lipgloss.Color{
	lipgloss.Color("#ff0000"),
	lipgloss.Color("#ff7f00"),
	lipgloss.Color("#ffff00"),
	lipgloss.Color("#00ff00"),
	lipgloss.Color("#0000ff"),
	lipgloss.Color("#4b0082"),
	lipgloss.Color("#8f00ff"),
}

func NewSplash(width, height int) *Splash {
	logo, err := os.ReadFile("tui/logo.ans")
	if err != nil {
		return &Splash{width: width, height: height}
	}

	// Initialize the grid of cells
	cells := make([][]Cell, height)
	for y := range cells {
		cells[y] = make([]Cell, width)
		for x := range cells[y] {
			cells[y][x] = Cell{
				char:     "猫咪"[rand.Intn(2) : rand.Intn(2)+1],
				revealed: false,
				colorIdx: rand.Intn(len(rainbowColors)),
				settled:  false,
			}
		}
	}

	return &Splash{
		width:    width,
		height:   height,
		logo:     string(logo),
		cells:    cells,
		complete: false,
		lastTick: time.Now(),
	}
}

func (splash *Splash) Init() tea.Cmd {
	rand.Seed(time.Now().UnixNano())
	return tea.Tick(time.Millisecond*50, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (splash *Splash) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case tickMsg:
		if splash.complete {
			return splash, nil
		}

		now := time.Now()
		if now.Sub(splash.lastTick) < time.Millisecond*50 {
			return splash, nil
		}
		splash.lastTick = now
		splash.frame++

		// Reveal cells in a wave pattern
		revealCount := 0
		settledCount := 0
		totalCells := splash.width * splash.height

		for y := 0; y < splash.height; y++ {
			for x := 0; x < splash.width; x++ {
				cell := &splash.cells[y][x]

				// Wave reveal pattern
				if !cell.revealed {
					if (x + y - splash.frame/2) < 0 {
						cell.revealed = true
					}
				} else {
					revealCount++

					// Cycle colors for revealed cells
					if !cell.settled {
						cell.colorIdx = (cell.colorIdx + 1) % len(rainbowColors)

						// Settle cells after they've been revealed for a while
						if splash.frame-((x+y)/2) > 50 {
							cell.settled = true
						}
					} else {
						settledCount++
					}
				}
			}
		}

		splash.complete = settledCount == totalCells

		return splash, tea.Tick(time.Millisecond*50, func(t time.Time) tea.Msg {
			return tickMsg(t)
		})
	}
	return splash, nil
}

func (splash *Splash) View() string {
	var sb strings.Builder

	// Center the content
	contentWidth := min(splash.width, 252)
	horizontalPadding := (splash.width - contentWidth) / 2

	for y := 0; y < splash.height; y++ {
		// Add horizontal padding
		sb.WriteString(strings.Repeat(" ", horizontalPadding))

		for x := 0; x < contentWidth; x++ {
			cell := splash.cells[y][x]
			if !cell.revealed {
				sb.WriteString(" ")
				continue
			}

			style := lipgloss.NewStyle()
			if cell.settled {
				style = style.Foreground(subtle)
			} else {
				style = style.Foreground(rainbowColors[cell.colorIdx])
			}
			sb.WriteString(style.Render(cell.char))
		}
		sb.WriteString("\n")
	}

	// If animation is complete, overlay the logo and title
	if splash.complete {
		// Create the centered content
		logoStyle := lipgloss.NewStyle().
			Width(contentWidth).
			Align(lipgloss.Center)

		content := lipgloss.JoinVertical(lipgloss.Center,
			logoStyle.Render(splash.logo),
			"",
			logoStyle.Foreground(lipgloss.Color("#874BFD")).
				Bold(true).
				Render("Ape Machine Shell"),
		)

		// Calculate vertical position for overlay
		lines := strings.Split(content, "\n")
		startY := (splash.height - len(lines)) / 2

		// Overlay the content
		contentLines := strings.Split(content, "\n")
		outputLines := strings.Split(sb.String(), "\n")
		for i, line := range contentLines {
			if startY+i < len(outputLines) {
				outputLines[startY+i] = strings.Repeat(" ", horizontalPadding) + line
			}
		}

		return strings.Join(outputLines, "\n")
	}

	return sb.String()
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

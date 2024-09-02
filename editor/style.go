package editor

import "github.com/charmbracelet/lipgloss"

const (
	initialInputs = 2
	maxInputs     = 6
	minInputs     = 1
	helpHeight    = 5
)

var (
	appStyle = lipgloss.NewStyle().Padding(1, 2, 1, 2)

	cursorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))

	cursorLineStyle = lipgloss.NewStyle().Background(
		lipgloss.Color("57"),
	).Foreground(
		lipgloss.Color("230"),
	)

	placeholderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("238"))

	endOfBufferStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("235"))

	focusedPlaceholderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("99"))

	focusedBorderStyle = lipgloss.NewStyle().Border(
		lipgloss.RoundedBorder(),
	).BorderForeground(
		lipgloss.Color("238"),
	)

	blurredBorderStyle = lipgloss.NewStyle().Border(lipgloss.HiddenBorder())

	statusBarStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")).
		Background(lipgloss.Color("57")).
		Width(80).  // This will be overridden in renderStatusBar
		Padding(0, 1).
		MarginTop(1)
)

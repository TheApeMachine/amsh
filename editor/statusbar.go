package editor

import (
	"github.com/charmbracelet/lipgloss"
)

type StatusBar struct {
	buffer *Buffer
}

func NewStatusBar(buffer *Buffer) *StatusBar {
	return &StatusBar{
		buffer: buffer,
	}
}

func (sb *StatusBar) Update() {
	// Update status bar information if needed
}

func (sb *StatusBar) Render() string {
	modeStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#FF5F87")).
		Padding(0, 1)

	fileStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#6124DF")).
		Padding(0, 1)

	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#3C3836")).
		Padding(0, 1)

	modeText := modeStyle.Render(sb.buffer.mode.String())
	fileText := fileStyle.Render(sb.buffer.filename)
	statusText := ""
	if sb.buffer.content != sb.buffer.lastSaved {
		statusText = statusStyle.Render("MODIFIED")
	}

	bar := lipgloss.JoinHorizontal(lipgloss.Left, modeText, fileText, statusText)
	return lipgloss.NewStyle().Width(sb.buffer.width).Render(bar)
}
package ui

import (
	"github.com/charmbracelet/lipgloss"
)

type ExpansionPreference int

const (
	Horizontal ExpansionPreference = iota
	Vertical
)

var (
	CursorStyle             = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
	PlaceholderStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("99"))
	BlurredPlaceholderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("238"))
	CursorLineStyle         = lipgloss.NewStyle().Background(lipgloss.Color("57")).Foreground(lipgloss.Color("230"))
	FocusedBaseStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
	BlurredBaseStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	EndOfBufferStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("235"))
)

package ui

import (
	"github.com/charmbracelet/lipgloss"
)

type LayoutPreference int

const (
	Horizontal LayoutPreference = iota
	Vertical
	Bottom
	Overlay
)

var (
	red       = lipgloss.AdaptiveColor{Light: "#FE5F86", Dark: "#FE5F86"}
	indigo    = lipgloss.AdaptiveColor{Light: "#5A56E0", Dark: "#7571F9"}
	green     = lipgloss.AdaptiveColor{Light: "#02BA84", Dark: "#02BF87"}
	Subtle    = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
	Highlight = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	Special   = lipgloss.AdaptiveColor{Light: "#43BF6D", Dark: "#73F59F"}
)

type Styles struct {
	Base,
	HeaderText,
	Status,
	StatusHeader,
	Highlight,
	ErrorHeaderText,
	DialogBox lipgloss.Style
}

func NewStyles() *Styles {
	return &Styles{
		Base:            lipgloss.NewStyle().Padding(1, 4, 0, 1),
		HeaderText:      lipgloss.NewStyle().Foreground(indigo).Bold(true).Padding(0, 1, 0, 2),
		Status:          lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(indigo).PaddingLeft(1).MarginTop(1),
		StatusHeader:    lipgloss.NewStyle().Foreground(green).Bold(true),
		Highlight:       lipgloss.NewStyle().Foreground(lipgloss.Color("212")),
		ErrorHeaderText: lipgloss.NewStyle().Foreground(indigo).Bold(true).Padding(0, 1, 0, 2).Foreground(red),
		DialogBox: lipgloss.NewStyle().Border(
			lipgloss.RoundedBorder(),
		).BorderForeground(
			lipgloss.Color("#874BFD"),
		).Padding(
			1, 0,
		).BorderTop(true).BorderLeft(true).BorderRight(true).BorderBottom(true),
	}
}

var (
	CursorStyle              = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
	CursorLineStyle          = lipgloss.NewStyle().Background(lipgloss.Color("57")).Foreground(lipgloss.Color("230"))
	PlaceholderStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("99"))
	EndOfBufferStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("235"))
	FocussedPlaceholderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("238"))
	FocusedBorderStyle       = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("238"))
	BlurredBorderStyle       = lipgloss.NewStyle().Border(lipgloss.HiddenBorder())
	KeywordStyle             = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF79C6")).Bold(true)
	VariableStyle            = lipgloss.NewStyle().Foreground(lipgloss.Color("#8BE9FD"))
	LiteralStyle             = lipgloss.NewStyle().Foreground(lipgloss.Color("#50FA7B"))
	CommentStyle             = lipgloss.NewStyle().Foreground(lipgloss.Color("#6272A4"))
	FunctionStyle            = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFB86C"))
)

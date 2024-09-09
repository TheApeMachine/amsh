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
	DialogBox,
	Cursor,
	CursorLine,
	CursorLineNumber,
	EndOfBuffer,
	LineNumber,
	Placeholder,
	Prompt,
	FocussedPlaceholder,
	FocusedBorderStyle,
	BlurredBorderStyle,
	KeywordStyle,
	VariableStyle,
	LiteralStyle,
	CommentStyle,
	Text,
	StatusNuggetStyle,
	StatusBarStyle,
	StatusStyle,
	EncodingStyle,
	StatusText,
	ModeInsertStyle,
	ModeNormalStyle,
	ModeVisualStyle,
	FunctionStyle lipgloss.Style
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
		Cursor:              lipgloss.NewStyle().Foreground(lipgloss.Color("212")),
		CursorLine:          lipgloss.NewStyle().Background(lipgloss.Color("57")).Foreground(lipgloss.Color("230")),
		CursorLineNumber:    lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
		Placeholder:         lipgloss.NewStyle().Foreground(lipgloss.Color("99")),
		EndOfBuffer:         lipgloss.NewStyle().Foreground(lipgloss.Color("235")),
		FocussedPlaceholder: lipgloss.NewStyle().Foreground(lipgloss.Color("238")),
		FocusedBorderStyle:  lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("238")),
		BlurredBorderStyle:  lipgloss.NewStyle().Border(lipgloss.HiddenBorder()),
		KeywordStyle:        lipgloss.NewStyle().Foreground(lipgloss.Color("#FF79C6")).Bold(true),
		VariableStyle:       lipgloss.NewStyle().Foreground(lipgloss.Color("#8BE9FD")),
		LiteralStyle:        lipgloss.NewStyle().Foreground(lipgloss.Color("#50FA7B")),
		CommentStyle:        lipgloss.NewStyle().Foreground(lipgloss.Color("#6272A4")),
		FunctionStyle:       lipgloss.NewStyle().Foreground(lipgloss.Color("#FFB86C")),
		Text:                lipgloss.NewStyle().Foreground(lipgloss.Color("#F8F8F2")),
		StatusNuggetStyle:   lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFDF5")).Padding(0, 1),
		StatusBarStyle: lipgloss.NewStyle().Foreground(
			lipgloss.AdaptiveColor{Light: "#343433", Dark: "#C1C6B2"},
		).Background(
			lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#353533"},
		),
		StatusStyle: lipgloss.NewStyle().Inherit(lipgloss.NewStyle().Foreground(
			lipgloss.AdaptiveColor{Light: "#343433", Dark: "#C1C6B2"},
		).Background(
			lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#353533"},
		)).Foreground(
			lipgloss.Color("#FFFDF5"),
		).Background(
			lipgloss.Color("#FF5F87"),
		).Padding(0, 1).MarginRight(1),
		EncodingStyle: lipgloss.NewStyle().Background(lipgloss.Color("#A550DF")).Align(lipgloss.Right),
		StatusText: lipgloss.NewStyle().Inherit(lipgloss.NewStyle().Foreground(
			lipgloss.AdaptiveColor{Light: "#343433", Dark: "#C1C6B2"},
		).Background(
			lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#353533"},
		)),
		ModeInsertStyle: lipgloss.NewStyle().Background(lipgloss.Color("#02BA84")).Foreground(lipgloss.Color("#FFFFFF")),
		ModeNormalStyle: lipgloss.NewStyle().Background(lipgloss.Color("#FE5F86")).Foreground(lipgloss.Color("#FFFFFF")),
		ModeVisualStyle: lipgloss.NewStyle().Background(lipgloss.Color("#5A56E0")).Foreground(lipgloss.Color("#FFFFFF")),
	}
}

func (s *Styles) ComputedCursorLine() lipgloss.Style {
	return s.CursorLine.Inherit(s.Base).Inline(true)
}

func (s *Styles) ComputedCursorLineNumber() lipgloss.Style {
	return s.CursorLineNumber.
		Inherit(s.CursorLine).
		Inherit(s.Base).
		Inline(true)
}

func (s *Styles) ComputedEndOfBuffer() lipgloss.Style {
	return s.EndOfBuffer.Inherit(s.Base).Inline(true)
}

func (s *Styles) ComputedLineNumber() lipgloss.Style {
	return s.LineNumber.Inherit(s.Base).Inline(true)
}

func (s *Styles) ComputedPlaceholder() lipgloss.Style {
	return s.Placeholder.Inherit(s.Base).Inline(true)
}

func (s *Styles) ComputedPrompt() lipgloss.Style {
	return s.Prompt.Inherit(s.Base).Inline(true)
}

func (s *Styles) ComputedText(plugin func(string) string) lipgloss.Style {
	return s.Text.Inherit(s.Base).Inline(true).Transform(plugin)
}

func (s *Styles) DefaultStyles() (*Styles, *Styles) {
	focused := &Styles{
		Base:             lipgloss.NewStyle(),
		CursorLine:       lipgloss.NewStyle().Background(lipgloss.AdaptiveColor{Light: "255", Dark: "0"}),
		CursorLineNumber: lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "240"}),
		EndOfBuffer:      lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "254", Dark: "0"}),
		LineNumber:       lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "249", Dark: "7"}),
		Placeholder:      lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
		Prompt:           lipgloss.NewStyle().Foreground(lipgloss.Color("7")),
		Text:             lipgloss.NewStyle(),
	}
	blurred := &Styles{
		Base:             lipgloss.NewStyle(),
		CursorLine:       lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "245", Dark: "7"}),
		CursorLineNumber: lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "249", Dark: "7"}),
		EndOfBuffer:      lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "254", Dark: "0"}),
		LineNumber:       lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "249", Dark: "7"}),
		Placeholder:      lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
		Prompt:           lipgloss.NewStyle().Foreground(lipgloss.Color("7")),
		Text:             lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "245", Dark: "7"}),
	}

	return focused, blurred
}

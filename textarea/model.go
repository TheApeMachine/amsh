package textarea

import (
	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/runeutil"
	"github.com/charmbracelet/bubbles/textarea/memoization"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/theapemachine/amsh/components"
	"github.com/theapemachine/amsh/ui"
)

const (
	minHeight        = 1
	defaultHeight    = 6
	defaultWidth     = 40
	defaultCharLimit = 1000000
	offset           = 4
)

// Internal messages for clipboard operations.
type (
	pasteMsg    string
	pasteErrMsg struct{ error }
)

/*
Model represents the state of the textarea component.
We now track the cursor position and the split content.
*/
type Model struct {
	Err                  error
	cache                *memoization.MemoCache[line, [][]rune]
	Prompt               string
	Placeholder          string
	ShowLineNumbers      bool
	EndOfBufferCharacter rune
	KeyMap               KeyMap
	FocusedStyle         *ui.Styles
	BlurredStyle         *ui.Styles
	styles               *ui.Styles
	Cursor               cursor.Model
	CharLimit            int
	MaxHeight            int
	MaxWidth             int
	promptFunc           func(line int) string
	promptWidth          int
	width                int
	height               int
	value                [][]rune
	focus                bool
	col                  int
	row                  int
	lastCharOffset       int
	viewport             *viewport.Model
	rsan                 runeutil.Sanitizer
	mode                 ui.Mode
	plugin               func(string) string
	state                components.State
}

/*
New creates a new textarea model with default settings.
*/
func New(width, height int) *Model {
	vp := viewport.New(0, 0)
	vp.KeyMap = viewport.KeyMap{}
	cur := cursor.New()

	styles := ui.NewStyles()

	focusedStyle, blurredStyle := styles.DefaultStyles()

	model := &Model{
		CharLimit:            defaultCharLimit,
		MaxHeight:            height - offset,
		MaxWidth:             width,
		Prompt:               lipgloss.ThickBorder().Left + " ",
		styles:               styles,
		FocusedStyle:         focusedStyle,
		BlurredStyle:         blurredStyle,
		cache:                memoization.NewMemoCache[line, [][]rune](height - offset),
		EndOfBufferCharacter: ' ',
		ShowLineNumbers:      true,
		Cursor:               cur,
		KeyMap:               DefaultKeyMap,
		value:                make([][]rune, height-offset),
		col:                  0,
		row:                  0,
		viewport:             &vp,
		mode:                 ui.ModeNormal,
		plugin:               func(s string) string { return s },
		state:                components.Inactive,
	}

	model.SetWidth(width)
	model.SetHeight(height - offset)

	return model
}

func (model *Model) Init() tea.Cmd {
	return nil
}

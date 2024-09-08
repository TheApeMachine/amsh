package textarea

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/cursor"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/rivo/uniseg"
)

var (
	s     strings.Builder
	style lipgloss.Style
)

// View renders the text area in its current state.
func (model *Model) View() string {
	if model.Value() == "" && model.row == 0 && model.col == 0 && model.Placeholder != "" {
		return model.placeholderView()
	}
	model.Cursor.TextStyle = model.styles.ComputedCursorLine()

	var (
		newLines         int
		widestLineNumber int
		lineInfo         = model.LineInfo()
	)

	s.Reset()

	displayLine := 0
	for l, line := range model.value {
		wrappedLines := model.memoizedWrap(line, model.width)

		if model.row == l {
			style = model.styles.ComputedCursorLine()
		} else {
			style = model.styles.ComputedText(model.plugin)
		}

		for wl, wrappedLine := range wrappedLines {
			prompt := model.getPromptString(displayLine)
			prompt = model.styles.ComputedPrompt().Render(prompt)
			s.WriteString(style.Render(prompt))
			displayLine++

			var ln string
			if model.ShowLineNumbers {
				lineNumber := " "
				if wl == 0 {
					lineNumber = strconv.Itoa(l + 1)
				}
				if model.row == l {
					ln = style.Render(model.styles.ComputedCursorLineNumber().Render(model.formatLineNumber(lineNumber)))
				} else {
					ln = style.Render(model.styles.ComputedLineNumber().Render(model.formatLineNumber(lineNumber)))
				}
				s.WriteString(ln)
			}

			// Note the widest line number for padding purposes later.
			lnw := lipgloss.Width(ln)
			if lnw > widestLineNumber {
				widestLineNumber = lnw
			}

			strwidth := uniseg.StringWidth(string(wrappedLine))
			padding := model.width - strwidth
			// If the trailing space causes the line to be wider than the
			// width, we should not draw it to the screen since it will result
			// in an extra space at the end of the line which can look off when
			// the cursor line is showing.
			if strwidth > model.width {
				// The character causing the line to be wider than the width is
				// guaranteed to be a space since any other character would
				// have been wrapped.
				wrappedLine = []rune(strings.TrimSuffix(string(wrappedLine), " "))
				padding -= model.width - strwidth
			}
			if model.row == l && lineInfo.RowOffset == wl {
				s.WriteString(style.Render(string(wrappedLine[:lineInfo.ColumnOffset])))
				if model.col >= len(line) && lineInfo.CharOffset >= model.width {
					model.Cursor.SetChar(" ")
					s.WriteString(model.Cursor.View())
				} else {
					model.Cursor.SetChar(string(wrappedLine[lineInfo.ColumnOffset]))
					s.WriteString(style.Render(model.Cursor.View()))
					s.WriteString(style.Render(string(wrappedLine[lineInfo.ColumnOffset+1:])))
				}
			} else {
				s.WriteString(style.Render(string(wrappedLine)))
			}
			s.WriteString(style.Render(strings.Repeat(" ", max(0, padding))))
			s.WriteRune('\n')
			newLines++
		}
	}

	// Always show at least `model.Height` lines at all times.
	// To do this we can simply pad out a few extra new lines in the view.
	for i := 0; i < model.height; i++ {
		prompt := model.getPromptString(displayLine)
		prompt = model.styles.ComputedPrompt().Render(prompt)
		s.WriteString(prompt)
		displayLine++

		// Write end of buffer content
		leftGutter := string(model.EndOfBufferCharacter)
		rightGapWidth := model.Width() - lipgloss.Width(leftGutter) + widestLineNumber
		rightGap := strings.Repeat(" ", max(0, rightGapWidth))
		s.WriteString(model.styles.ComputedEndOfBuffer().Render(leftGutter + rightGap))
		s.WriteRune('\n')
	}

	model.viewport.SetContent(modifier(s.String()))
	return model.styles.Base.Render(model.viewport.View())
}

// repositionView repositions the view of the viewport based on the defined
// scrolling behavior.
func (model *Model) repositionView() {
	min := model.viewport.YOffset
	max := min + model.viewport.Height - 1

	if row := model.cursorLineNumber(); row < min {
		model.viewport.LineUp(min - row)
	} else if row > max {
		model.viewport.LineDown(row - max)
	}
}

// Width returns the width of the textarea.
func (model *Model) Width() int {
	return model.width
}

// Height returns the current height of the textarea.
func (model *Model) Height() int {
	return model.height
}

// formatLineNumber formats the line number for display dynamically based on
// the maximum number of lines
func (model *Model) formatLineNumber(x any) string {
	// XXX: ultimately we should use a max buffer height, which has yet to be
	// implemented.
	digits := len(strconv.Itoa(model.MaxHeight))
	return fmt.Sprintf(" %*v ", digits, x)
}

func (model *Model) getPromptString(displayLine int) (prompt string) {
	prompt = model.Prompt
	if model.promptFunc == nil {
		return prompt
	}
	prompt = model.promptFunc(displayLine)
	pl := uniseg.StringWidth(prompt)
	if pl < model.promptWidth {
		prompt = fmt.Sprintf("%*s%s", model.promptWidth-pl, "", prompt)
	}
	return prompt
}

// placeholderView returns the prompt and placeholder view, if any.
func (model *Model) placeholderView() string {
	var (
		s     strings.Builder
		p     = model.Placeholder
		style = model.styles.ComputedPlaceholder()
	)

	// word wrap lines
	pwordwrap := ansi.Wordwrap(p, model.width, "")
	// wrap lines (handles lines that could not be word wrapped)
	pwrap := ansi.Hardwrap(pwordwrap, model.width, true)
	// split string by new lines
	plines := strings.Split(strings.TrimSpace(pwrap), "\n")

	for i := 0; i < model.height; i++ {
		lineStyle := model.styles.ComputedPlaceholder()
		lineNumberStyle := model.styles.ComputedLineNumber()
		if len(plines) > i {
			lineStyle = model.styles.ComputedCursorLine()
			lineNumberStyle = model.styles.ComputedCursorLineNumber()
		}

		// render prompt
		prompt := model.getPromptString(i)
		prompt = model.styles.ComputedPrompt().Render(prompt)
		s.WriteString(lineStyle.Render(prompt))

		// when show line numbers enabled:
		// - render line number for only the cursor line
		// - indent other placeholder lines
		// this is consistent with vim with line numbers enabled
		if model.ShowLineNumbers {
			var ln string

			switch {
			case i == 0:
				ln = strconv.Itoa(i + 1)
				fallthrough
			case len(plines) > i:
				s.WriteString(lineStyle.Render(lineNumberStyle.Render(model.formatLineNumber(ln))))
			default:
			}
		}

		switch {
		// first line
		case i == 0:
			// first character of first line as cursor with character
			model.Cursor.TextStyle = model.styles.ComputedPlaceholder()
			model.Cursor.SetChar(string(plines[0][0]))
			s.WriteString(lineStyle.Render(model.Cursor.View()))

			// the rest of the first line
			s.WriteString(lineStyle.Render(style.Render(plines[0][1:] + strings.Repeat(" ", max(0, model.width-uniseg.StringWidth(plines[0]))))))
		// remaining lines
		case len(plines) > i:
			// current line placeholder text
			if len(plines) > i {
				s.WriteString(lineStyle.Render(style.Render(plines[i] + strings.Repeat(" ", max(0, model.width-uniseg.StringWidth(plines[i]))))))
			}
		default:
			// end of line buffer character
			eob := model.styles.ComputedEndOfBuffer().Render(string(model.EndOfBufferCharacter))
			s.WriteString(eob)
		}

		// terminate with new line
		s.WriteRune('\n')
	}

	model.viewport.SetContent(s.String())
	return model.styles.Base.Render(model.viewport.View())
}

// Blink returns the blink command for the cursor.
func Blink() tea.Msg {
	return cursor.Blink()
}

func (model *Model) memoizedWrap(runes []rune, width int) [][]rune {
	input := line{runes: runes, width: width}
	if v, ok := model.cache.Get(input); ok {
		return v
	}
	v := wrap(runes, width)
	model.cache.Set(input, v)
	return v
}

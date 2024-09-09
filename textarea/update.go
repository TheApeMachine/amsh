package textarea

import (
	"strings"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/runeutil"
	"github.com/charmbracelet/bubbles/textarea/memoization"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/rivo/uniseg"
	"github.com/theapemachine/amsh/components"
	"github.com/theapemachine/amsh/logger"
	"github.com/theapemachine/amsh/messages"
	"github.com/theapemachine/amsh/ui"
)

func (model *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	EndSection := logger.StartSection("textarea.Update", "update")
	defer EndSection()

	logger.Debug("<- <%T> %v", msg, msg)

	oldRow, oldCol := model.cursorLineNumber(), model.col
	var cmds []tea.Cmd

	if model.value[model.row] == nil {
		model.value[model.row] = make([]rune, 0)
	}

	if model.MaxHeight > 0 && model.MaxHeight != model.cache.Capacity() {
		model.cache = memoization.NewMemoCache[line, [][]rune](model.MaxHeight)
	}

	switch msg := msg.(type) {
	case messages.Message[ui.Mode]:
		logger.Debug("Setting texarea mode to: %v", msg.Data)
		model.mode = msg.Data

		if !messages.ShouldProcessMessage(model.state, msg.Context) {
			return model, nil
		}
	case tea.KeyMsg:
		switch model.mode {
		case ui.ModeNormal:
			logger.Debug("<- NORMAL <tea.KeyMsg> %s", msg.String())
			model.handleNormalMode(msg)
		case ui.ModeInsert:
			logger.Debug("<- INSERT <tea.KeyMsg> %s", msg.String())
			model.handleInsertMode(msg)
		default:
			logger.Debug("<- UNKNOWN <tea.KeyMsg> %s", msg.String())
		}
	case pasteMsg:
		model.insertRunesFromUserInput([]rune(msg))
	case pasteErrMsg:
		model.Err = msg
	}

	vp, cmd := model.viewport.Update(msg)
	model.viewport = &vp
	cmds = append(cmds, cmd)

	newRow, newCol := model.cursorLineNumber(), model.col
	model.Cursor, cmd = model.Cursor.Update(msg)
	if (newRow != oldRow || newCol != oldCol) && model.Cursor.Mode() == cursor.CursorBlink {
		model.Cursor.Blink = false
		cmd = model.Cursor.BlinkCmd()
	}
	cmds = append(cmds, cmd)

	model.repositionView()

	return model, tea.Batch(cmds...)
}

// SetValue sets the value of the text input.
func (model *Model) SetValue(s string) {
	model.Reset()
	model.InsertString(s)
}

// InsertString inserts a string at the cursor position.
func (model *Model) InsertString(s string) {
	model.insertRunesFromUserInput([]rune(s))
}

// InsertRune inserts a rune at the cursor position.
func (model *Model) InsertRune(r rune) {
	model.insertRunesFromUserInput([]rune{r})
}

// insertRunesFromUserInput inserts runes at the current cursor position.
func (model *Model) insertRunesFromUserInput(runes []rune) {
	// Clean up any special characters in the input provided by the
	// clipboard. This avoids bugs due to e.g. tab characters and
	// whatnot.
	runes = model.san().Sanitize(runes)

	var availSpace int
	if model.CharLimit > 0 {
		availSpace = model.CharLimit - model.Length()
		// If the char limit's been reached, cancel.
		if availSpace <= 0 {
			return
		}
		// If there's not enough space to paste the whole thing cut the pasted
		// runes down so they'll fit.
		if availSpace < len(runes) {
			runes = runes[:availSpace]
		}
	}

	// Split the input into lines.
	var lines [][]rune
	lstart := 0
	for i := 0; i < len(runes); i++ {
		if runes[i] == '\n' {
			// Queue a line to become a new row in the text area below.
			// Beware to clamp the max capacity of the slice, to ensure no
			// data from different rows get overwritten when later edits
			// will modify this line.
			lines = append(lines, runes[lstart:i:i])
			lstart = i + 1
		}
	}
	if lstart <= len(runes) {
		// The last line did not end with a newline character.
		// Take it now.
		lines = append(lines, runes[lstart:])
	}

	// Obey the maximum height limit.
	if model.MaxHeight > 0 && len(model.value)+len(lines)-1 > model.MaxHeight {
		allowedHeight := max(0, model.MaxHeight-len(model.value)+1)
		lines = lines[:allowedHeight]
	}

	if len(lines) == 0 {
		// Nothing left to insert.
		return
	}

	// Save the remainder of the original line at the current
	// cursor position.
	tail := make([]rune, len(model.value[model.row][model.col:]))
	copy(tail, model.value[model.row][model.col:])

	// Paste the first line at the current cursor position.
	model.value[model.row] = append(model.value[model.row][:model.col], lines[0]...)
	model.col += len(lines[0])

	if numExtraLines := len(lines) - 1; numExtraLines > 0 {
		// Add the new lines.
		// We try to reuse the slice if there's already space.
		var newGrid [][]rune
		if cap(model.value) >= len(model.value)+numExtraLines {
			// Can reuse the extra space.
			newGrid = model.value[:len(model.value)+numExtraLines]
		} else {
			// No space left; need a new slice.
			newGrid = make([][]rune, len(model.value)+numExtraLines)
			copy(newGrid, model.value[:model.row+1])
		}
		// Add all the rows that were after the cursor in the original
		// grid at the end of the new grid.
		copy(newGrid[model.row+1+numExtraLines:], model.value[model.row+1:])
		model.value = newGrid
		// Insert all the new lines in the middle.
		for _, l := range lines[1:] {
			model.row++
			model.value[model.row] = l
			model.col = len(l)
		}
	}

	// Finally add the tail at the end of the last line inserted.
	model.value[model.row] = append(model.value[model.row], tail...)

	model.SetCursor(model.col)
}

// Value returns the value of the text input.
func (model *Model) Value() string {
	if model.value == nil {
		return ""
	}

	var v strings.Builder
	for _, l := range model.value {
		v.WriteString(string(l))
		v.WriteByte('\n')
	}

	return strings.TrimSuffix(v.String(), "\n")
}

// Length returns the number of characters currently in the text input.
func (model *Model) Length() int {
	var l int
	for _, row := range model.value {
		l += uniseg.StringWidth(string(row))
	}
	// We add len(model.value) to include the newline characters.
	return l + len(model.value) - 1
}

// Focused returns the focus state on the model.
func (model *Model) Focused() bool {
	return model.focus
}

// Focus sets the focus state on the model. When the model is in focus it can
// receive keyboard input and the cursor will be hidden.
func (model *Model) Focus() tea.Cmd {
	model.state = components.Focused
	model.styles = model.FocusedStyle
	return model.Cursor.Focus()
}

// Blur removes the focus state on the model. When the model is blurred it can
// not receive keyboard input and the cursor will be hidden.
func (model *Model) Blur() {
	model.state = components.Active
	model.styles = model.BlurredStyle
	model.Cursor.Blur()
}

// Reset sets the input to its default state with no input.
func (model *Model) Reset() {
	startCap := model.MaxHeight
	if startCap <= 0 {
		startCap = model.height
	}
	model.value = make([][]rune, minHeight, startCap)
	model.col = 0
	model.row = 0
	model.viewport.GotoTop()
	model.SetCursor(0)
}

// san initializes or retrieves the rune sanitizer.
func (model *Model) san() runeutil.Sanitizer {
	if model.rsan == nil {
		// Textinput has all its input on a single line so collapse
		// newlines/tabs to single spaces.
		model.rsan = runeutil.NewSanitizer()
	}
	return model.rsan
}

// SetWidth sets the width of the textarea to fit exactly within the given width.
// This means that the textarea will account for the width of the prompt and
// whether or not line numbers are being shown.
//
// Ensure that SetWidth is called after setting the Prompt and ShowLineNumbers,
// It is important that the width of the textarea be exactly the given width
// and no more.
func (model *Model) SetWidth(w int) {
	// Update prompt width only if there is no prompt function as SetPromptFunc
	// updates the prompt width when it is called.
	if model.promptFunc == nil {
		model.promptWidth = uniseg.StringWidth(model.Prompt)
	}

	// Add base style borders and padding to reserved outer width.
	reservedOuter := model.styles.Base.GetHorizontalFrameSize()

	// Add prompt width to reserved inner width.
	reservedInner := model.promptWidth

	// Add line number width to reserved inner width.
	if model.ShowLineNumbers {
		const lnWidth = 4 // Up to 3 digits for line number plus 1 margin.
		reservedInner += lnWidth
	}

	// Input width must be at least one more than the reserved inner and outer
	// width. This gives us a minimum input width of 1.
	minWidth := reservedInner + reservedOuter + 1
	inputWidth := max(w, minWidth)

	// Input width must be no more than maximum width.
	if model.MaxWidth > 0 {
		inputWidth = min(inputWidth, model.MaxWidth)
	}

	// Since the width of the viewport and input area is dependent on the width of
	// borders, prompt and line numbers, we need to calculate it by subtracting
	// the reserved width from themodel.

	model.viewport.Width = inputWidth - reservedOuter
	model.width = inputWidth - reservedOuter - reservedInner
}

// SetHeight sets the height of the textarea.
func (model *Model) SetHeight(h int) {
	if model.MaxHeight > 0 {
		model.height = clamp(h-offset, minHeight, model.MaxHeight)
		model.viewport.Height = clamp(h-offset, minHeight, model.MaxHeight)
	} else {
		model.height = max(h-offset, minHeight)
		model.viewport.Height = max(h-offset, minHeight)
	}
}

// SetPromptFunc supersedes the Prompt field and sets a dynamic prompt
// instead.
// If the function returns a prompt that is shorter than the
// specified promptWidth, it will be padded to the left.
// If it returns a prompt that is longer, display artifacts
// may occur; the caller is responsible for computing an adequate
// promptWidth.
func (model *Model) SetPromptFunc(promptWidth int, fn func(lineIdx int) string) {
	model.promptFunc = fn
	model.promptWidth = promptWidth
}

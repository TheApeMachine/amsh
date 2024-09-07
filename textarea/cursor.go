package textarea

import (
	"strings"
	"unicode"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	rw "github.com/mattn/go-runewidth"
	"github.com/rivo/uniseg"
)

// CursorDown moves the cursor down by one line, ensuring it stays within bounds.
func (model *Model) CursorDown() {
	// Move down if there is a next line available
	if model.row < len(model.value)-1 {
		model.row++
	}

	// Ensure the column stays within the bounds of the new line's length
	model.col = min(model.col, len(model.value[model.row]))

	// Reset the lastCharOffset to 0 after moving vertically
	model.lastCharOffset = 0
}

// CursorUp moves the cursor up by one line.
func (model *Model) CursorUp() {
	li := model.LineInfo()
	charOffset := max(model.lastCharOffset, li.CharOffset)
	model.lastCharOffset = charOffset

	if li.RowOffset <= 0 && model.row > 0 {
		model.row--
		model.col = len(model.value[model.row])
	} else {
		// Move the cursor to the end of the previous line.
		// This can be done by moving the cursor to the start of the line and
		// then subtracting 2 to account for the trailing space we keep on
		// soft-wrapped lines.
		model.col = li.StartColumn - 2
	}

	nli := model.LineInfo()
	model.col = nli.StartColumn

	if nli.Width <= 0 {
		return
	}

	offset := 0
	for offset < charOffset {
		if model.col >= len(model.value[model.row]) || offset >= nli.CharWidth-1 {
			break
		}
		offset += rw.RuneWidth(model.value[model.row][model.col])
		model.col++
	}
}

// SetCursor moves the cursor to the given position. If the position is
// out of bounds the cursor will be moved to the start or end accordingly.
func (model *Model) SetCursor(col int) {
	model.col = clamp(col, 0, len(model.value[model.row]))
	// Any time that we move the cursor horizontally we need to reset the last
	// offset so that the horizontal position when navigating is adjusted.
	model.lastCharOffset = 0
}

// CursorStart moves the cursor to the start of the input field.
func (model *Model) CursorStart() {
	model.SetCursor(0)
}

// CursorEnd moves the cursor to the end of the input field.
func (model *Model) CursorEnd() {
	model.SetCursor(len(model.value[model.row]))
}

// deleteBeforeCursor deletes all text before the cursor. Returns whether or
// not the cursor blink should be reset.
func (model *Model) deleteBeforeCursor() {
	model.value[model.row] = model.value[model.row][model.col:]
	model.SetCursor(0)
}

// deleteAfterCursor deletes all text after the cursor. Returns whether or not
// the cursor blink should be reset. If input is masked delete everything after
// the cursor so as not to reveal word breaks in the masked input.
func (model *Model) deleteAfterCursor() {
	model.value[model.row] = model.value[model.row][:model.col]
	model.SetCursor(len(model.value[model.row]))
}

// transposeLeft exchanges the runes at the cursor and immediately
// before. No-op if the cursor is at the beginning of the line.  If
// the cursor is not at the end of the line yet, moves the cursor to
// the right.
func (model *Model) transposeLeft() {
	if model.col == 0 || len(model.value[model.row]) < 2 {
		return
	}
	if model.col >= len(model.value[model.row]) {
		model.SetCursor(model.col - 1)
	}
	model.value[model.row][model.col-1], model.value[model.row][model.col] = model.value[model.row][model.col], model.value[model.row][model.col-1]
	if model.col < len(model.value[model.row]) {
		model.SetCursor(model.col + 1)
	}
}

// deleteWordLeft deletes the word left to the cursor. Returns whether or not
// the cursor blink should be reset.
func (model *Model) deleteWordLeft() {
	if model.col == 0 || len(model.value[model.row]) == 0 {
		return
	}

	// Linter note: it's critical that we acquire the initial cursor position
	// here prior to altering it via SetCursor() below. As such, moving this
	// call into the corresponding if clause does not apply here.
	oldCol := model.col //nolint:ifshort

	model.SetCursor(model.col - 1)
	for unicode.IsSpace(model.value[model.row][model.col]) {
		if model.col <= 0 {
			break
		}
		// ignore series of whitespace before cursor
		model.SetCursor(model.col - 1)
	}

	for model.col > 0 {
		if !unicode.IsSpace(model.value[model.row][model.col]) {
			model.SetCursor(model.col - 1)
		} else {
			if model.col > 0 {
				// keep the previous space
				model.SetCursor(model.col + 1)
			}
			break
		}
	}

	if oldCol > len(model.value[model.row]) {
		model.value[model.row] = model.value[model.row][:model.col]
	} else {
		model.value[model.row] = append(model.value[model.row][:model.col], model.value[model.row][oldCol:]...)
	}
}

// deleteWordRight deletes the word right to the cursor.
func (model *Model) deleteWordRight() {
	if model.col >= len(model.value[model.row]) || len(model.value[model.row]) == 0 {
		return
	}

	oldCol := model.col

	for model.col < len(model.value[model.row]) && unicode.IsSpace(model.value[model.row][model.col]) {
		// ignore series of whitespace after cursor
		model.SetCursor(model.col + 1)
	}

	for model.col < len(model.value[model.row]) {
		if !unicode.IsSpace(model.value[model.row][model.col]) {
			model.SetCursor(model.col + 1)
		} else {
			break
		}
	}

	if model.col > len(model.value[model.row]) {
		model.value[model.row] = model.value[model.row][:oldCol]
	} else {
		model.value[model.row] = append(model.value[model.row][:oldCol], model.value[model.row][model.col:]...)
	}

	model.SetCursor(oldCol)
}

// characterRight moves the cursor one character to the right.
func (model *Model) characterRight() {
	if model.col < len(model.value[model.row]) {
		model.SetCursor(model.col + 1)
	} else {
		if model.row < len(model.value)-1 {
			model.row++
			model.CursorStart()
		}
	}
}

// characterLeft moves the cursor one character to the left.
// If insideLine is set, the cursor is moved to the last
// character in the previous line, instead of one past that.
func (model *Model) characterLeft(insideLine bool) {
	if model.col == 0 && model.row != 0 {
		model.row--
		model.CursorEnd()
		if !insideLine {
			return
		}
	}
	if model.col > 0 {
		model.SetCursor(model.col - 1)
	}
}

// wordLeft moves the cursor one word to the left. Returns whether or not the
// cursor blink should be reset. If input is masked, move input to the start
// so as not to reveal word breaks in the masked input.
func (model *Model) wordLeft() {
	for {
		model.characterLeft(true /* insideLine */)
		if model.col < len(model.value[model.row]) && !unicode.IsSpace(model.value[model.row][model.col]) {
			break
		}
	}

	for model.col > 0 {
		if unicode.IsSpace(model.value[model.row][model.col-1]) {
			break
		}
		model.SetCursor(model.col - 1)
	}
}

// wordRight moves the cursor one word to the right. Returns whether or not the
// cursor blink should be reset. If the input is masked, move input to the end
// so as not to reveal word breaks in the masked input.
func (model *Model) wordRight() {
	model.doWordRight(func(int, int) { /* nothing */ })
}

func (model *Model) doWordRight(fn func(charIdx int, pos int)) {
	// Skip spaces forward.
	for model.col >= len(model.value[model.row]) || unicode.IsSpace(model.value[model.row][model.col]) {
		if model.row == len(model.value)-1 && model.col == len(model.value[model.row]) {
			// End of text.
			break
		}
		model.characterRight()
	}

	charIdx := 0
	for model.col < len(model.value[model.row]) {
		if unicode.IsSpace(model.value[model.row][model.col]) {
			break
		}
		fn(charIdx, model.col)
		model.SetCursor(model.col + 1)
		charIdx++
	}
}

// uppercaseRight changes the word to the right to uppercase.
func (model *Model) uppercaseRight() {
	model.doWordRight(func(_ int, i int) {
		model.value[model.row][i] = unicode.ToUpper(model.value[model.row][i])
	})
}

// lowercaseRight changes the word to the right to lowercase.
func (model *Model) lowercaseRight() {
	model.doWordRight(func(_ int, i int) {
		model.value[model.row][i] = unicode.ToLower(model.value[model.row][i])
	})
}

// capitalizeRight changes the word to the right to title case.
func (model *Model) capitalizeRight() {
	model.doWordRight(func(charIdx int, i int) {
		if charIdx == 0 {
			model.value[model.row][i] = unicode.ToTitle(model.value[model.row][i])
		}
	})
}

// moveToBegin moves the cursor to the beginning of the input.
func (model *Model) moveToBegin() {
	model.row = 0
	model.SetCursor(0)
}

// moveToEnd moves the cursor to the end of the input.
func (model *Model) moveToEnd() {
	model.row = len(model.value) - 1
	model.SetCursor(len(model.value[model.row]))
}

// cursorLineNumber returns the line number that the cursor is on.
// This accounts for soft wrapped lines.
func (model *Model) cursorLineNumber() int {
	line := 0
	for i := 0; i < model.row; i++ {
		// Calculate the number of lines that the current line will be split
		// into.
		line += len(model.memoizedWrap(model.value[i], model.width))
	}
	line += model.LineInfo().RowOffset
	return line
}

// mergeLineBelow merges the current line the cursor is on with the line below.
func (model *Model) mergeLineBelow(row int) {
	if row >= len(model.value)-1 {
		return
	}

	// To perform a merge, we will need to combine the two lines and then
	model.value[row] = append(model.value[row], model.value[row+1]...)

	// Shift all lines up by one
	for i := row + 1; i < len(model.value)-1; i++ {
		model.value[i] = model.value[i+1]
	}

	// And, remove the last line
	if len(model.value) > 0 {
		model.value = model.value[:len(model.value)-1]
	}
}

// mergeLineAbove merges the current line the cursor is on with the line above.
func (model *Model) mergeLineAbove(row int) {
	if row <= 0 {
		return
	}

	model.col = len(model.value[row-1])
	model.row = model.row - 1

	// To perform a merge, we will need to combine the two lines and then
	model.value[row-1] = append(model.value[row-1], model.value[row]...)

	// Shift all lines up by one
	for i := row; i < len(model.value)-1; i++ {
		model.value[i] = model.value[i+1]
	}

	// And, remove the last line
	if len(model.value) > 0 {
		model.value = model.value[:len(model.value)-1]
	}
}

func (model *Model) splitLine(row, col int) {
	// To perform a split, take the current line and keep the content before
	// the cursor, take the content after the cursor and make it the content of
	// the line underneath, and shift the remaining lines down by one
	head, tailSrc := model.value[row][:col], model.value[row][col:]
	tail := make([]rune, len(tailSrc))
	copy(tail, tailSrc)

	model.value = append(model.value[:row+1], model.value[row:]...)

	model.value[row] = head
	model.value[row+1] = tail

	model.col = 0
	model.row++
}

// Paste is a command for pasting from the clipboard into the text input.
func Paste() tea.Msg {
	str, err := clipboard.ReadAll()
	if err != nil {
		return pasteErrMsg{err}
	}
	return pasteMsg(str)
}

func wrap(runes []rune, width int) [][]rune {
	var (
		lines  = [][]rune{{}}
		word   = []rune{}
		row    int
		spaces int
	)

	// Word wrap the runes
	for _, r := range runes {
		if unicode.IsSpace(r) {
			spaces++
		} else {
			word = append(word, r)
		}

		if spaces > 0 {
			if uniseg.StringWidth(string(lines[row]))+uniseg.StringWidth(string(word))+spaces > width {
				row++
				lines = append(lines, []rune{})
				lines[row] = append(lines[row], word...)
				lines[row] = append(lines[row], repeatSpaces(spaces)...)
				spaces = 0
				word = nil
			} else {
				lines[row] = append(lines[row], word...)
				lines[row] = append(lines[row], repeatSpaces(spaces)...)
				spaces = 0
				word = nil
			}
		} else {
			// If the last character is a double-width rune, then we may not be able to add it to this line
			// as it might cause us to go past the width.
			lastCharLen := rw.RuneWidth(word[len(word)-1])
			if uniseg.StringWidth(string(word))+lastCharLen > width {
				// If the current line has any content, let's move to the next
				// line because the current word fills up the entire line.
				if len(lines[row]) > 0 {
					row++
					lines = append(lines, []rune{})
				}
				lines[row] = append(lines[row], word...)
				word = nil
			}
		}
	}

	if uniseg.StringWidth(string(lines[row]))+uniseg.StringWidth(string(word))+spaces >= width {
		lines = append(lines, []rune{})
		lines[row+1] = append(lines[row+1], word...)
		// We add an extra space at the end of the line to account for the
		// trailing space at the end of the previous soft-wrapped lines so that
		// behaviour when navigating is consistent and so that we don't need to
		// continually add edges to handle the last line of the wrapped input.
		spaces++
		lines[row+1] = append(lines[row+1], repeatSpaces(spaces)...)
	} else {
		lines[row] = append(lines[row], word...)
		spaces++
		lines[row] = append(lines[row], repeatSpaces(spaces)...)
	}

	return lines
}

func repeatSpaces(n int) []rune {
	return []rune(strings.Repeat(string(' '), n))
}

func clamp(v, low, high int) int {
	if high < low {
		low, high = high, low
	}
	return min(high, max(low, v))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

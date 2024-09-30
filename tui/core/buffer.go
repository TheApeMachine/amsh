package core

import (
	"fmt"
	"os"
)

type Mode uint

const (
	NormalMode Mode = iota
	InsertMode
	CommandMode
)

type Buffer struct {
	mode      Mode
	Data      [][]rune
	Cursor    *Cursor
	height    int
	Filename  string
	StatusMsg string
	queue     *Queue
}

func NewBuffer(height int, cursor *Cursor, queue *Queue) *Buffer {
	return &Buffer{
		Data:   [][]rune{{}}, // Initialize with one empty line
		Cursor: cursor,
		height: height,
		queue:  queue,
	}
}

func (buffer *Buffer) Render() {
	fmt.Print("\033[H\033[J") // Clear screen and move to top-left

	for i, line := range buffer.Data {
		if i >= buffer.height-1 {
			break
		}
		fmt.Printf("\033[%d;1H", i+1)
		fmt.Print(string(line))
	}

	// Move cursor back to current position
	buffer.Cursor.Move(buffer.Cursor.X, buffer.Cursor.Y)

	// Redraw status line
	buffer.ShowStatus(buffer.getStatusMessage())
	buffer.flushStdout()
}

func (buffer *Buffer) renderLineFromCol(lineIdx int, colIdx int) {
	if lineIdx >= len(buffer.Data) {
		return
	}
	line := buffer.Data[lineIdx]
	fmt.Printf("\033[%d;%dH", lineIdx+1, colIdx+1)
	fmt.Print("\033[K") // Clear from cursor to end of line
	if colIdx < len(line) {
		fmt.Print(string(line[colIdx:]))
	}
	// Move cursor back to position after the change
	buffer.Cursor.Move(colIdx+1, lineIdx+1)
	buffer.flushStdout()
}

func (buffer *Buffer) renderFromLine(startLine int) {
	for i := startLine; i < len(buffer.Data) && i < buffer.height-1; i++ {
		fmt.Printf("\033[%d;1H", i+1)
		fmt.Print("\033[K") // Clear the line
		fmt.Print(string(buffer.Data[i]))
	}
	// Move cursor back to current position
	buffer.Cursor.Move(buffer.Cursor.X, buffer.Cursor.Y)
	buffer.flushStdout()
}

func (buffer *Buffer) ShowStatus(status string) {
	fmt.Printf("\033[%d;1H", buffer.height)
	fmt.Print("\033[K") // Clear the status line
	fmt.Print(status)
	buffer.Cursor.Move(buffer.Cursor.X, buffer.Cursor.Y)
	buffer.flushStdout()
}

func (buffer *Buffer) getStatusMessage() string {
	switch buffer.mode {
	case NormalMode:
		if buffer.Filename != "" {
			return fmt.Sprintf("NORMAL | %s", buffer.Filename)
		}
		return "NORMAL"
	case InsertMode:
		return "-- INSERT --"
	case CommandMode:
		return ":" + buffer.StatusMsg
	default:
		return ""
	}
}

func (buffer *Buffer) flushStdout() {
	os.Stdout.Sync()
}

func (buffer *Buffer) InsertChar(ch rune) {
	lineIdx := buffer.Cursor.Y - 1
	colIdx := buffer.Cursor.X - 1

	// Ensure the buffer has enough lines
	for len(buffer.Data) <= lineIdx {
		buffer.Data = append(buffer.Data, []rune{})
	}

	line := buffer.Data[lineIdx]

	// Insert the character into the line
	if colIdx > len(line) {
		colIdx = len(line)
	}
	line = append(line[:colIdx], append([]rune{ch}, line[colIdx:]...)...)
	buffer.Data[lineIdx] = line

	// Re-render the current line from the cursor position
	buffer.renderLineFromCol(lineIdx, colIdx)

	// Move cursor to the position after the inserted character
	buffer.Cursor.MoveForward(1, len(line)+1)
}

func (buffer *Buffer) DeleteCharUnderCursor() {
	lineIdx := buffer.Cursor.Y - 1
	colIdx := buffer.Cursor.X - 1

	if lineIdx >= len(buffer.Data) || colIdx >= len(buffer.Data[lineIdx]) {
		// Nothing to delete
		return
	}

	// Remove the character
	line := buffer.Data[lineIdx]
	line = append(line[:colIdx], line[colIdx+1:]...)
	buffer.Data[lineIdx] = line

	// Re-render the current line from the cursor position
	buffer.renderLineFromCol(lineIdx, colIdx)
}

func (buffer *Buffer) HandleBackspace() {
	if buffer.Cursor.X > 1 {
		lineIdx := buffer.Cursor.Y - 1
		colIdx := buffer.Cursor.X - 2 // Character before the cursor

		line := buffer.Data[lineIdx]

		if colIdx >= len(line) {
			colIdx = len(line) - 1
		}

		if colIdx >= 0 {
			// Remove the character from the line
			line = append(line[:colIdx], line[colIdx+1:]...)
			buffer.Data[lineIdx] = line

			// Move cursor back
			buffer.Cursor.MoveBackward(1)

			// Re-render the line from the deletion point
			buffer.renderLineFromCol(lineIdx, colIdx)
		}
	} else if buffer.Cursor.Y > 1 {
		// At the start of the line, merge with the previous line
		currentLineIdx := buffer.Cursor.Y - 1
		previousLineIdx := buffer.Cursor.Y - 2

		currentLine := buffer.Data[currentLineIdx]
		previousLine := buffer.Data[previousLineIdx]

		// Move cursor to the end of the previous line
		buffer.Cursor.Move(len(previousLine)+1, previousLineIdx+1)

		// Merge lines
		previousLine = append(previousLine, currentLine...)
		buffer.Data[previousLineIdx] = previousLine

		// Remove the current line
		buffer.Data = append(buffer.Data[:currentLineIdx], buffer.Data[currentLineIdx+1:]...)

		// Re-render from the previous line
		buffer.renderFromLine(previousLineIdx)
	}
}

func (buffer *Buffer) HandleEnter() {
	lineIdx := buffer.Cursor.Y - 1
	colIdx := buffer.Cursor.X - 1

	line := buffer.Data[lineIdx]

	// Split the line at the cursor position
	newLine := append([]rune{}, line[colIdx:]...)
	buffer.Data[lineIdx] = line[:colIdx]
	buffer.Data = append(buffer.Data[:lineIdx+1], append([][]rune{newLine}, buffer.Data[lineIdx+1:]...)...)

	// Move cursor to the beginning of the new line
	buffer.Cursor.Move(1, buffer.Cursor.Y+1)

	// Re-render from the current line
	buffer.renderFromLine(lineIdx)

	buffer.flushStdout()
}

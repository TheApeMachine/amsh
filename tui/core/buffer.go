package core

import (
	"fmt"
	"os"

	"github.com/theapemachine/amsh/data"
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
	fmt.Print("\\033[H\\033[J") // Clear screen and move to top-left

	for i, line := range buffer.Data {
		if i >= buffer.height-1 {
			break
		}
		fmt.Printf("\\033[%d;1H", i+1)
		fmt.Print(string(line))
	}

	// Move cursor back to current position
	buffer.Cursor.Move(buffer.Cursor.X, buffer.Cursor.Y)

	// Redraw status line
	buffer.ShowStatus(buffer.getStatusMessage())
	buffer.flushStdout()
}

// Method to publish buffer events using Artifact
func (buffer *Buffer) publishBufferEvent(eventType string, content string) {
	artifact := data.New("buffer", "event", eventType, []byte(content))
	buffer.queue.Publish("buffer_event", artifact)
}

func (buffer *Buffer) renderLineFromCol(lineIdx int, colIdx int) {
	if lineIdx >= len(buffer.Data) {
		return
	}
	line := buffer.Data[lineIdx]
	fmt.Printf("\\033[%d;%dH", lineIdx+1, colIdx+1)
	fmt.Print("\\033[K") // Clear from cursor to end of line
	if colIdx < len(line) {
		fmt.Print(string(line[colIdx:]))
	}
	// Move cursor back to position after the change
	buffer.Cursor.Move(colIdx+1, lineIdx+1)
	buffer.flushStdout()

	// Publish an event for rendering a line
	buffer.publishBufferEvent("render_line", fmt.Sprintf("Rendered line %d from column %d", lineIdx, colIdx))
}

func (buffer *Buffer) renderFromLine(startLine int) {
	for i := startLine; i < len(buffer.Data) && i < buffer.height-1; i++ {
		fmt.Printf("\\033[%d;1H", i+1)
		fmt.Print("\\033[K") // Clear the line
		fmt.Print(string(buffer.Data[i]))
	}
	// Move cursor back to current position
	buffer.Cursor.Move(buffer.Cursor.X, buffer.Cursor.Y)
	buffer.flushStdout()

	// Publish an event for rendering from a specific line
	buffer.publishBufferEvent("render_from_line", fmt.Sprintf("Rendered from line %d", startLine))
}

func (buffer *Buffer) ShowStatus(status string) {
	fmt.Printf("\\033[%d;1H", buffer.height)
	fmt.Print("\\033[K") // Clear the status line
	fmt.Print(status)
	buffer.Cursor.Move(buffer.Cursor.X, buffer.Cursor.Y)
	buffer.flushStdout()

	// Publish an event for showing status
	buffer.publishBufferEvent("show_status", status)
}

func (buffer *Buffer) InsertRune(r rune) {
	x, y := buffer.Cursor.X, buffer.Cursor.Y
	if y < 0 || y >= len(buffer.Data) {
		return
	}
	if x < 0 || x > len(buffer.Data[y]) {
		return
	}

	// Insert the rune at the current cursor position
	buffer.Data[y] = append(buffer.Data[y][:x], append([]rune{r}, buffer.Data[y][x:]...)...)
	buffer.Cursor.Move(x+1, y)

	// Publish an event for inserting a rune
	buffer.publishBufferEvent("insert_rune", fmt.Sprintf("Inserted rune '%c' at (%d, %d)", r, x, y))

	// Re-render the line from the current cursor position
	buffer.renderLineFromCol(y, x)
	buffer.flushStdout()
}

func (buffer *Buffer) getStatusMessage() string {
	return fmt.Sprintf("Buffer: %s", buffer.Filename)
}

func (buffer *Buffer) flushStdout() {
	os.Stdout.Sync()
}

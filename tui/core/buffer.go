package core

import (
	"fmt"
	"os"
	"sync"

	"github.com/theapemachine/amsh/data"
	"github.com/theapemachine/amsh/errnie"
)

type Buffer struct {
	mode      Mode
	Data      [][]rune
	Cursor    *Cursor
	height    int
	Filename  string
	StatusMsg string
	queue     *Queue
	mutex     sync.RWMutex // Mutex to synchronize buffer access
}

func NewBuffer(height int, cursor *Cursor, queue *Queue) *Buffer {
	buffer := &Buffer{
		Data:   [][]rune{{}}, // Initialize with one empty line
		Cursor: cursor,
		height: height,
		queue:  queue,
		mutex:  sync.RWMutex{},
	}

	// Start event handling
	go buffer.run()

	return buffer
}

// run listens to 'buffer' topic and handles events
func (buffer *Buffer) run() {
	bufferSub := buffer.queue.Subscribe("buffer")
	for artifact := range bufferSub {
		role, err := artifact.Role()
		if err != nil {
			errnie.Error(err)
			continue
		}
		switch role {
		case "InsertChar":
			payload, _ := artifact.Payload()
			if len(payload) > 0 {
				buffer.InsertRune(rune(payload[0]))
			}
		case "DeleteChar":
			buffer.DeleteRune()
		case "Backspace":
			buffer.DeleteRune()
		case "Enter":
			buffer.InsertRune('\n')
		default:
			errnie.Warn("Unknown buffer event role: %s", role)
		}
	}
}

// InsertRune inserts a rune at the current cursor position
func (buffer *Buffer) InsertRune(r rune) {
	buffer.mutex.Lock()
	defer buffer.mutex.Unlock()

	x, y := buffer.Cursor.X, buffer.Cursor.Y
	// Adjust indices for 0-based slices
	y = y - 1
	x = x - 1

	// Ensure y is within bounds
	if y < 0 {
		y = 0
	} else if y >= len(buffer.Data) {
		y = len(buffer.Data) - 1
	}

	// Ensure x is within bounds
	if x < 0 {
		x = 0
	} else if x > len(buffer.Data[y]) {
		x = len(buffer.Data[y])
	}

	line := buffer.Data[y]

	if r == '\n' {
		// Split the current line at the cursor position
		newLine := append([]rune{}, line[:x]...)
		nextLine := append([]rune{}, line[x:]...)
		buffer.Data[y] = newLine
		buffer.Data = append(buffer.Data[:y+1], append([][]rune{nextLine}, buffer.Data[y+1:]...)...)
		// Move cursor to the beginning of the next line
		buffer.Cursor.Move(1, buffer.Cursor.Y+1)
	} else {
		// Insert the rune at the current position
		newLine := append(line[:x], append([]rune{r}, line[x:]...)...)
		buffer.Data[y] = newLine
		// Move cursor forward
		buffer.Cursor.Move(buffer.Cursor.X+1, buffer.Cursor.Y)
	}

	// Re-render the buffer
	buffer.Render()
}

// DeleteRune deletes a rune before the current cursor position
func (buffer *Buffer) DeleteRune() {
	buffer.mutex.Lock()
	defer buffer.mutex.Unlock()

	x, y := buffer.Cursor.X, buffer.Cursor.Y
	// Adjust indices for 0-based slices
	y = y - 1
	x = x - 1

	if y < 0 || y >= len(buffer.Data) {
		// Nothing to delete
		return
	}

	line := buffer.Data[y]

	if x > len(line) {
		x = len(line)
	}

	if x > 0 {
		// Delete the rune before position x
		newLine := append(line[:x-1], line[x:]...)
		buffer.Data[y] = newLine
		// Move cursor backward
		buffer.Cursor.Move(buffer.Cursor.X-1, buffer.Cursor.Y)
	} else if y > 0 {
		// Join with previous line
		prevLine := buffer.Data[y-1]
		buffer.Data[y-1] = append(prevLine, line...)
		// Remove current line
		buffer.Data = append(buffer.Data[:y], buffer.Data[y+1:]...)
		// Move cursor to end of previous line
		buffer.Cursor.Move(len(prevLine)+1, buffer.Cursor.Y-1)
	}

	// Re-render the buffer
	buffer.Render()
}

func (buffer *Buffer) Render() {
	buffer.mutex.RLock()
	defer buffer.mutex.RUnlock()

	// Render buffer lines
	for i, line := range buffer.Data {
		if i >= buffer.height-1 { // Reserve the last line for the command/status
			break
		}
		fmt.Printf("\033[%d;1H\033[K%s", i+1, string(line)) // Move to line, clear it, and print
	}

	// Clear any lines below if buffer shrank
	for i := len(buffer.Data); i < buffer.height-1; i++ {
		fmt.Printf("\033[%d;1H\033[K", i+1) // Clear line
	}

	// Redraw status line (commands or status messages)
	buffer.ShowStatus(buffer.getStatusMessage())

	// Move cursor to current position
	buffer.Cursor.Move(buffer.Cursor.X, buffer.Cursor.Y)
	buffer.flushStdout()
}

// Method to publish buffer events using Artifact
func (buffer *Buffer) publishBufferEvent(eventType string, content string) {
	artifact := data.New("buffer_event", eventType, content, nil)
	buffer.queue.Publish("buffer_event", artifact)
}

func (buffer *Buffer) renderLineFromCol(lineIdx int, colIdx int) {
	buffer.mutex.RLock()
	defer buffer.mutex.RUnlock()

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

	// Publish an event for rendering a line
	buffer.publishBufferEvent("render_line", fmt.Sprintf("Rendered line %d from column %d", lineIdx+1, colIdx+1))
}

func (buffer *Buffer) renderFromLine(startLine int) {
	buffer.mutex.RLock()
	defer buffer.mutex.RUnlock()

	for i := startLine; i < len(buffer.Data) && i < buffer.height-1; i++ {
		fmt.Printf("\033[%d;1H", i+1)
		fmt.Print("\033[K") // Clear the line
		fmt.Print(string(buffer.Data[i]))
	}
	// Move cursor back to current position
	buffer.Cursor.Move(buffer.Cursor.X, buffer.Cursor.Y)
	buffer.flushStdout()

	// Publish an event for rendering from a specific line
	buffer.publishBufferEvent("render_from_line", fmt.Sprintf("Rendered from line %d", startLine+1))
}

func (buffer *Buffer) ShowStatus(message string) {
	// Save current cursor position
	fmt.Print("\033[s")
	// Move cursor to status line and display message
	fmt.Printf("\033[%d;1H\033[K%s", buffer.height, message)
	// Restore cursor position
	fmt.Print("\033[u")
	buffer.flushStdout()
}

func (buffer *Buffer) getStatusMessage() string {
	return fmt.Sprintf("Buffer: %s", buffer.Filename)
}

func (buffer *Buffer) flushStdout() {
	os.Stdout.Sync()
}

// Update updates the content of the buffer at a specific line and column.
func (buffer *Buffer) Update(line, col int, content []rune) {
	buffer.mutex.Lock()
	defer buffer.mutex.Unlock()

	if line < 0 || line >= len(buffer.Data) {
		fmt.Println("Line index out of range")
		return
	}

	if col < 0 || col > len(buffer.Data[line]) {
		fmt.Println("Column index out of range")
		return
	}

	// Insert the content at the specified position.
	lineData := buffer.Data[line]
	newLine := append(lineData[:col], append(content, lineData[col:]...)...)
	buffer.Data[line] = newLine

	buffer.renderFromLine(line) // Re-render the buffer from the modified line

	// Publish an event to the queue indicating that the buffer has been updated.
	artifact := data.New("buffer_event", "buffer_update", fmt.Sprintf("Buffer updated at line %d, col %d", line+1, col+1), nil)
	buffer.queue.Publish("buffer_event", artifact)
}

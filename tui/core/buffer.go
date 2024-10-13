// File: core/buffer.go

package core

import (
	"fmt"
	"os"
	"sync"

	"github.com/theapemachine/amsh/errnie"
)

type Buffer struct {
	Data     [][]rune
	Cursor   *Cursor
	height   int
	Filename string
	queue    *Queue
	mutex    sync.RWMutex // Mutex to synchronize buffer access
}

func NewBuffer(height int, cursor *Cursor, queue *Queue) *Buffer {
	errnie.Trace()
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

// run listens to 'buffer_event' topic and handles events
func (buffer *Buffer) run() {
	errnie.Trace()
	bufferSub := buffer.queue.Subscribe("buffer_event")
	for artifact := range bufferSub {
		scope, err := artifact.Scope()
		if err != nil {
			errnie.Error(err)
			continue
		}
		switch scope {
		case "InsertChar":
			payload, _ := artifact.Payload()
			if len(payload) > 0 {
				buffer.InsertRune(rune(payload[0]))
			}
		case "DeleteChar":
			buffer.DeleteRune()
		case "Enter":
			buffer.InsertRune('\n')
		default:
			errnie.Warn("Unknown buffer event scope: %s", scope)
		}
	}
}

// InsertRune inserts a rune at the current cursor position
func (buffer *Buffer) InsertRune(r rune) {
	errnie.Trace()
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
	errnie.Trace()
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
	errnie.Trace()
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

// ShowStatus displays a status message at the bottom of the buffer
func (buffer *Buffer) ShowStatus(message string) {
	errnie.Trace()
	// Save current cursor position
	fmt.Print("\033[s")
	// Move cursor to status line and display message
	fmt.Printf("\033[%d;1H\033[K%s", buffer.height, message)
	// Restore cursor position
	fmt.Print("\033[u")
	buffer.flushStdout()
}

func (buffer *Buffer) getStatusMessage() string {
	errnie.Trace()
	return fmt.Sprintf("Buffer: %s", buffer.Filename)
}

func (buffer *Buffer) flushStdout() {
	errnie.Trace()
	os.Stdout.Sync()
}

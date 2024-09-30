package chat

import (
	"fmt"
	"os"
	"strings"

	"github.com/theapemachine/amsh/tui/core"
)

// Define ChatWindow struct:
type ChatWindow struct {
	x                 int // Top-left X position
	y                 int // Top-left Y position
	width             int
	height            int
	Active            bool
	buffer            *core.Buffer
	underlyingContent []string
}

func NewChatWindow(width, height int, buffer *core.Buffer) *ChatWindow {
	return &ChatWindow{
		width:  int(float64(width) * 0.8),
		height: int(float64(height) * 0.3),
		x:      (width - int(float64(width)*0.8)) / 2,
		y:      (height - int(float64(height)*0.3)) / 2,
		Active: false,
		buffer: buffer,
	}
}

func (cw *ChatWindow) SaveUnderlyingContent() {
	// Save the content under the chat window area
	cw.underlyingContent = make([]string, cw.height)
	for i := 0; i < cw.height; i++ {
		// Read the line from the buffer
		lineIdx := cw.y - 1 + i
		if lineIdx < len(cw.buffer.Data) {
			line := cw.buffer.Data[lineIdx]
			cw.underlyingContent[i] = string(line)
		} else {
			cw.underlyingContent[i] = ""
		}
	}
}

func (cw *ChatWindow) RestoreUnderlyingContent() {
	// Restore the saved content
	for i, content := range cw.underlyingContent {
		lineIdx := cw.y - 1 + i
		if lineIdx < len(cw.buffer.Data) {
			cw.buffer.Data[lineIdx] = []rune(content)
		}
		fmt.Printf("\033[%d;1H", lineIdx+1)
		fmt.Print("\033[K") // Clear the line
		fmt.Print(content)
	}
	// Move cursor back to current position
	cw.buffer.Cursor.Move(cw.buffer.Cursor.X, cw.buffer.Cursor.Y)
	cw.flushStdout()
}

func (cw *ChatWindow) Draw() {
	// Draw a box representing the chat window
	// Use ANSI escape codes to draw lines and corners
	// Position: cw.x, cw.y
	// Size: cw.width, cw.height
	// Example:
	fmt.Printf("\033[%d;%dH", cw.y, cw.x)
	// Draw top border
	fmt.Print("+" + strings.Repeat("-", cw.width-2) + "+")
	// Draw sides
	for i := 1; i < cw.height-1; i++ {
		fmt.Printf("\033[%d;%dH", cw.y+i, cw.x)
		fmt.Print("|" + strings.Repeat(" ", cw.width-2) + "|")
	}
	// Draw bottom border
	fmt.Printf("\033[%d;%dH", cw.y+cw.height-1, cw.x)
	fmt.Print("+" + strings.Repeat("-", cw.width-2) + "+")
	cw.flushStdout()
}

func (cw *ChatWindow) flushStdout() {
	os.Stdout.Sync()
}

func (cw *ChatWindow) AddMessage(msg string) {
	// Add message to buffer
	cw.buffer.Data = append(cw.buffer.Data, []rune(msg))
	// If buffer exceeds the window height, remove oldest messages
	if len(cw.buffer.Data) > cw.height-2 {
		cw.buffer.Data = cw.buffer.Data[1:]
	}
	// Re-draw the messages inside the chat window
	cw.DrawMessages()
}

func (cw *ChatWindow) DrawMessages() {
	for i, msg := range cw.buffer.Data {
		fmt.Printf("\033[%d;%dH", cw.y+1+i, cw.x+1)
		fmt.Printf("%-*s", cw.width-2, string(msg))
	}
	cw.flushStdout()
}

func (cw *ChatWindow) UpdateInputDisplay(input string) {
	// Display the current input at the bottom of the chat window
	fmt.Printf("\033[%d;%dH", cw.y+cw.height-2, cw.x+1)
	fmt.Print(strings.Repeat(" ", cw.width-2)) // Clear the line
	fmt.Printf("\033[%d;%dH", cw.y+cw.height-2, cw.x+1)
	fmt.Printf("Input: %s", input)
	cw.flushStdout()
}

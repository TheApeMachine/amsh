// File: core/keyboard.go

package core

import (
	"fmt"
	"os"

	"github.com/theapemachine/amsh/data"
	"github.com/theapemachine/amsh/errnie"
)

type Keyboard struct {
	queue              *Queue
	mode               Mode
	chatWindowActive   bool
	chatInputBuffer    string
	commandInputBuffer string   // Field for Command Mode input
	ctx                *Context // Reference to Context
}

// NewKeyboard initializes a new Keyboard.
func NewKeyboard(queue *Queue) *Keyboard {
	errnie.Trace()
	keyboard := &Keyboard{
		queue: queue,
		mode:  &Normal{},
	}
	return keyboard
}

// SetContext assigns the Context to the Keyboard and subscribes to mode changes.
func (keyboard *Keyboard) SetContext(ctx *Context) {
	errnie.Trace()
	keyboard.ctx = ctx

	// Subscribe to mode_change events to update the current mode.
	modeSub := keyboard.queue.Subscribe("mode_change")
	go keyboard.handleModeChange(modeSub)
}

func (keyboard *Keyboard) handleKeys(b byte) {
	errnie.Trace()
	switch keyboard.mode.(type) {
	case *Normal:
		keyboard.handleNormalMode(b)
	case *Insert:
		keyboard.handleInsertMode(b)
	case *Command:
		keyboard.handleCommandModeKey(b)
	}
}

// HandleInput processes keyboard inputs and publishes relevant events based on the current mode.
func (keyboard *Keyboard) HandleInput(b byte) {
	errnie.Trace()
	if b == 27 { // ESC key
		// Handle escape sequences (e.g., arrow keys)
		buf := make([]byte, 2)
		n, err := os.Stdin.Read(buf)
		if err == nil && n == 2 && buf[0] == '[' {
			switch buf[1] {
			case 'A':
				// Up arrow
				artifact := data.New("Keyboard", "cursor_move", "up", nil)
				keyboard.queue.Publish("cursor_event", artifact)
			case 'B':
				// Down arrow
				artifact := data.New("Keyboard", "cursor_move", "down", nil)
				keyboard.queue.Publish("cursor_event", artifact)
			case 'C':
				// Right arrow
				artifact := data.New("Keyboard", "cursor_move", "right", nil)
				keyboard.queue.Publish("cursor_event", artifact)
			case 'D':
				// Left arrow
				artifact := data.New("Keyboard", "cursor_move", "left", nil)
				keyboard.queue.Publish("cursor_event", artifact)
			default:
				// Unknown sequence, ignore
			}
		} else {
			// It's an ESC key press
			keyboard.handleKeys(b)
		}
		return
	}

	keyboard.handleKeys(b)
}

// handleModeChange processes mode_change events to switch modes.
func (keyboard *Keyboard) handleModeChange(modeSub <-chan *data.Artifact) {
	errnie.Trace()
	for artifact := range modeSub {
		var newMode Mode
		switch artifact.Peek("scope") {
		case "NormalMode":
			newMode = &Normal{}
		case "InsertMode":
			newMode = &Insert{}
		case "CommandMode":
			newMode = &Command{}
		default:
			errnie.Warn("Unknown mode: %s", artifact.Peek("scope"))
			continue
		}

		// Exit the current mode and enter the new mode.
		if keyboard.mode != nil {
			keyboard.mode.Exit()
		}
		keyboard.mode = newMode
		if keyboard.mode != nil {
			keyboard.mode.Enter(keyboard.ctx)
		}
	}
}

// handleNormalMode processes input in Normal Mode.
func (keyboard *Keyboard) handleNormalMode(b byte) {
	errnie.Trace()
	switch b {
	case 'i':
		// Switch to Insert Mode
		artifact := data.New("Keyboard", "mode_change", "InsertMode", nil)
		keyboard.queue.Publish("mode_change", artifact)
		fmt.Print("\033[?25h") // Show cursor
	case 'h':
		// Move left
		artifact := data.New("Keyboard", "cursor_move", "left", nil)
		keyboard.queue.Publish("cursor_event", artifact)
	case 'l':
		// Move right
		artifact := data.New("Keyboard", "cursor_move", "right", nil)
		keyboard.queue.Publish("cursor_event", artifact)
	case 'j':
		// Move down
		artifact := data.New("Keyboard", "cursor_move", "down", nil)
		keyboard.queue.Publish("cursor_event", artifact)
	case 'k':
		// Move up
		artifact := data.New("Keyboard", "cursor_move", "up", nil)
		keyboard.queue.Publish("cursor_event", artifact)
	case ':':
		// Switch to Command Mode
		artifact := data.New("Keyboard", "mode_change", "CommandMode", nil)
		keyboard.queue.Publish("mode_change", artifact)
		fmt.Print("\033[?25l") // Hide cursor
	case 'q':
		// Publish a quit event
		artifact := data.New("Keyboard", "app_event", "Quit", nil)
		keyboard.queue.Publish("app_event", artifact)
	}
}

// handleInsertMode processes input in Insert Mode.
func (keyboard *Keyboard) handleInsertMode(b byte) {
	errnie.Trace()
	switch b {
	case 127: // Backspace key
		artifact := data.New("Keyboard", "buffer_modify", "DeleteChar", nil)
		keyboard.queue.Publish("buffer_event", artifact)
	case 13: // Enter key
		artifact := data.New("Keyboard", "buffer_modify", "Enter", nil)
		keyboard.queue.Publish("buffer_event", artifact)
	case 27: // Escape key to return to Normal Mode
		artifact := data.New("Keyboard", "mode_change", "NormalMode", nil)
		keyboard.queue.Publish("mode_change", artifact)
		fmt.Print("\033[?25h") // Show cursor
	default:
		if b >= 32 && b <= 126 {
			// Publish an insert character event
			artifact := data.New("Keyboard", "buffer_modify", "InsertChar", []byte{b})
			keyboard.queue.Publish("buffer_event", artifact)
		}
	}
}

// handleCommandModeKey handles input in Command Mode.
func (keyboard *Keyboard) handleCommandModeKey(b byte) {
	errnie.Trace()
	switch b {
	case '\r', '\n': // Enter key
		artifact := data.New("Keyboard", "command_input", "SubmitCommandInput", nil)
		keyboard.queue.Publish("command_input", artifact)
	case 127, '\b': // Backspace key
		artifact := data.New("Keyboard", "command_input", "BackspaceCommandInput", nil)
		keyboard.queue.Publish("command_input", artifact)
	default:
		// Handle other printable characters
		if b >= 32 && b <= 126 {
			artifact := data.New("Keyboard", "command_input", "UpdateCommandInput", []byte{b})
			keyboard.queue.Publish("command_input", artifact)
		}
	}
}

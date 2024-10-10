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
	ctx                *Context // New field to reference Context
}

// NewKeyboard initializes a new Keyboard without a Context.
func NewKeyboard(queue *Queue) *Keyboard {
	keyboard := &Keyboard{
		queue: queue,
		mode:  &Normal{},
	}

	return keyboard
}

func (keyboard *Keyboard) SetContext(ctx *Context) {
	keyboard.ctx = ctx

	// Subscribe to mode_change events to update the current mode.
	modeSub := keyboard.queue.Subscribe("mode_change")
	go keyboard.handleModeChange(modeSub)
}

// getContext safely retrieves the Context.
func (keyboard *Keyboard) getContext() *Context {
	return keyboard.ctx
}

func (keyboard *Keyboard) handleKeys(b byte) {
	switch keyboard.mode.(type) {
	case *Normal:
		keyboard.handleNormalMode(b)
	case *Insert:
		keyboard.handleInsertMode(b)
	case *Command:
		keyboard.handleCommandModeKey(b)
	default:
		errnie.Warn("Unhandled mode type")
	}
}

func (keyboard *Keyboard) handleCommandModeKey(b byte) {
	switch b {
	case '\r', '\n': // Enter key
		artifact := data.New("keyboard", "SubmitCommandInput", "", nil)
		keyboard.queue.Publish("command_input", artifact)
	case 127, '\b': // Backspace key
		artifact := data.New("keyboard", "BackspaceCommandInput", "", nil)
		keyboard.queue.Publish("command_input", artifact)
	default:
		// Handle other printable characters
		if b >= 32 && b <= 126 {
			artifact := data.New("keyboard", "UpdateCommandInput", "", []byte{b})
			keyboard.queue.Publish("command_input", artifact)
		}
	}
}

func (keyboard *Keyboard) HandleInput(b byte) {
	keyboard.handleKeys(b)
}

func (keyboard *Keyboard) handleModeChange(modeSub <-chan *data.Artifact) {
	for artifact := range modeSub {
		role, err := artifact.Role()
		if err != nil {
			errnie.Error(err)
			continue
		}

		var newMode Mode
		switch role {
		case "NormalMode":
			newMode = &Normal{}
		case "InsertMode":
			newMode = &Insert{}
		case "CommandMode":
			newMode = &Command{}
		default:
			errnie.Warn("Unknown mode: %s", role)
			continue
		}

		// Exit the current mode and enter the new mode.
		if keyboard.mode != nil {
			keyboard.mode.Exit()
		}
		keyboard.mode = newMode
		if keyboard.mode != nil {
			keyboard.mode.Enter(keyboard.ctx) // Pass the context here
		}
	}
}

func (keyboard *Keyboard) handleNormalMode(b byte) {
	switch b {
	case 'i':
		// Switch to Insert Mode
		artifact := data.New("Keyboard", "InsertMode", "", nil)
		keyboard.queue.Publish("mode_change", artifact)
		fmt.Print("\033[?25h") // Show cursor
	case ':':
		// Switch to Command Mode
		artifact := data.New("Keyboard", "CommandMode", "", nil)
		keyboard.queue.Publish("mode_change", artifact)
		fmt.Print("\033[?25l") // Hide cursor
	case 'q':
		// Publish a quit event
		artifact := data.New("app_event", "quit", "", nil)
		keyboard.queue.Publish("app_event", artifact)
	case 'h':
		// Move left in normal mode
		artifact := data.New("Keyboard", "MoveLeft", "", nil)
		keyboard.queue.Publish("buffer", artifact)
	case 'l':
		// Move right in normal mode
		artifact := data.New("Keyboard", "MoveRight", "", nil)
		keyboard.queue.Publish("buffer", artifact)
	case 'j':
		// Move down in normal mode
		artifact := data.New("Keyboard", "MoveDown", "", nil)
		keyboard.queue.Publish("buffer", artifact)
	case 'k':
		// Move up in normal mode
		artifact := data.New("Keyboard", "MoveUp", "", nil)
		keyboard.queue.Publish("buffer", artifact)
	case 'x':
		// Publish a delete character event
		artifact := data.New("Keyboard", "DeleteChar", "", nil)
		keyboard.queue.Publish("buffer", artifact)
	}
}

func (keyboard *Keyboard) handleInsertMode(b byte) {
	switch b {
	case 127: // Backspace key
		artifact := data.New("Keyboard", "DeleteChar", "", nil)
		keyboard.queue.Publish("buffer", artifact)
	case 13: // Enter key
		artifact := data.New("Keyboard", "Enter", "", nil)
		keyboard.queue.Publish("buffer", artifact)
	case 27: // Escape key to return to Normal Mode
		artifact := data.New("Keyboard", "NormalMode", "", nil)
		keyboard.queue.Publish("mode_change", artifact)
		fmt.Print("\033[?25h") // Show cursor
	default:
		if b >= 32 && b <= 126 {
			// Publish an insert character event
			artifact := data.New("Keyboard", "InsertChar", "", []byte{b})
			keyboard.queue.Publish("buffer", artifact)
		}
		// Ignore other keys (e.g., arrow keys)
	}
}

// Example: handleCommandMode publishing to "command_event" topic
func (keyboard *Keyboard) handleCommandMode(b byte) {
	if b == 13 { // Enter key
		command := keyboard.commandInputBuffer
		keyboard.commandInputBuffer = "" // Reset the buffer

		switch command {
		case "q", "quit":
			// Publish a quit event
			artifact := data.New("app_event", "Quit", "quit", nil)
			keyboard.queue.Publish("app_event", artifact)
		// Add more commands here as needed
		default:
			errnie.Warn("Unknown command: %s", command)
		}
	} else if b == 27 { // Escape key to exit Command Mode
		artifact := data.New("mode_event", "NormalMode", "exit_command_mode", nil)
		keyboard.queue.Publish("mode_event", artifact)
		fmt.Print("\033[?25h") // Show cursor
	} else {
		// Append character to command buffer
		keyboard.commandInputBuffer += string(b)
		// Optionally, update the command input display
		artifact := data.New("Keyboard", "UpdateCommandInput", "", []byte(keyboard.commandInputBuffer))
		keyboard.queue.Publish("command_input", artifact)
	}
}

func (keyboard *Keyboard) ReadInput() {
	b := make([]byte, 1)
	n, err := os.Stdin.Read(b)

	if err != nil || n == 0 {
		return
	}

	if keyboard.chatWindowActive {
		keyboard.handleChatInput(b[0])
		return
	}

	if b[0] == 7 {
		artifact := data.New("Keyboard", "ToggleChat", "", nil)
		keyboard.queue.Publish("chat", artifact)
		return
	}

	if b[0] == 27 { // Escape character
		if _, ok := keyboard.mode.(*Normal); ok {
			// Potential escape sequence
			b2 := make([]byte, 2)
			n2, err := os.Stdin.Read(b2)
			if err != nil || n2 != 2 {
				return
			}

			if b2[0] == 91 {
				switch b2[1] {
				case 65:
					// Arrow up
					artifact := data.New("Keyboard", "MoveUp", "", nil)
					keyboard.queue.Publish("buffer", artifact)
				case 66:
					// Arrow down
					artifact := data.New("Keyboard", "MoveDown", "", nil)
					keyboard.queue.Publish("buffer", artifact)
				case 67:
					// Arrow right
					artifact := data.New("Keyboard", "MoveRight", "", nil)
					keyboard.queue.Publish("buffer", artifact)
				case 68:
					// Arrow left
					artifact := data.New("Keyboard", "MoveLeft", "", nil)
					keyboard.queue.Publish("buffer", artifact)
				}
			}
		}
		return
	}

	switch mode := keyboard.mode.(type) {
	case *Normal:
		_ = mode
		keyboard.handleNormalMode(b[0])
	case *Insert:
		_ = mode
		keyboard.handleInsertMode(b[0])
	case *Command:
		_ = mode
		// Add Command mode handling if needed
	}
}

func (keyboard *Keyboard) handleChatInput(b byte) {
	// Handle input for the chat window
	switch b {
	case 13: // Enter key
		// Send message to AI system
		message := keyboard.chatInputBuffer
		keyboard.chatInputBuffer = ""
		artifact := data.New("Keyboard", "SendMessage", "", []byte(message))
		keyboard.queue.Publish("chat", artifact)
	case 127: // Backspace
		if len(keyboard.chatInputBuffer) > 0 {
			keyboard.chatInputBuffer = keyboard.chatInputBuffer[:len(keyboard.chatInputBuffer)-1]
			// Update chat input display
			artifact := data.New("Keyboard", "UpdateChatInput", "", []byte(keyboard.chatInputBuffer))
			keyboard.queue.Publish("chat", artifact)
		}
	default:
		// Append character to input buffer
		keyboard.chatInputBuffer += string(b)
		// Update chat input display
		artifact := data.New("Keyboard", "UpdateChatInput", "", []byte(keyboard.chatInputBuffer))
		keyboard.queue.Publish("chat", artifact)
	}
}

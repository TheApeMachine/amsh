package core

import (
	"os"

	"github.com/theapemachine/amsh/data"
	"github.com/theapemachine/amsh/errnie"
)

type Keyboard struct {
	queue            *Queue
	mode             Mode
	chatWindowActive bool
	chatInputBuffer  string
}

func NewKeyboard(queue *Queue) *Keyboard {
	var (
		role string
		err  error
	)

	keyboard := &Keyboard{
		queue: queue,
		mode:  &Normal{}, // Initialize with Normal mode
	}

	// Subscribe to mode_change to update the current mode.
	modeSub := queue.Subscribe("mode_change")
	go func() {
		for artifact := range modeSub {
			if role, err = artifact.Role(); err != nil {
				errnie.Error(err)
				continue
			}

			switch role {
			case "NormalMode":
				keyboard.mode = &Normal{}
			case "InsertMode":
				keyboard.mode = &Insert{}
			case "CommandMode":
				keyboard.mode = &Command{}
			}
		}
	}()

	// Subscribe to chat_state to update chatWindowActive flag
	chatStateSub := queue.Subscribe("chat_state")
	go func() {
		for artifact := range chatStateSub {
			payload, err := artifact.Payload()
			if err != nil {
				errnie.Error(err)
				continue
			}
			state := string(payload)
			if state == "active" {
				keyboard.chatWindowActive = true
			} else if state == "inactive" {
				keyboard.chatWindowActive = false
			}
		}
	}()

	return keyboard
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

func (keyboard *Keyboard) handleNormalMode(b byte) {
	switch b {
	case 'i':
		// Publish an event to switch to Insert Mode
		artifact := data.New("Keyboard", "InsertMode", "", nil)
		keyboard.queue.Publish("mode_change", artifact)
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
	case ':':
		// Publish an event to switch to Command Mode
		artifact := data.New("Keyboard", "CommandMode", "", nil)
		keyboard.queue.Publish("mode_change", artifact)
	case 'q':
		// Publish a quit event
		artifact := data.New("Keyboard", "Quit", "", nil)
		keyboard.queue.Publish("app", artifact)
	case 'x':
		// Publish a delete character event
		artifact := data.New("Keyboard", "DeleteChar", "", nil)
		keyboard.queue.Publish("buffer", artifact)
	}
}

// handleInsertMode processes input in Insert Mode.
func (keyboard *Keyboard) handleInsertMode(b byte) {
	switch b {
	case 127: // Backspace key
		artifact := data.New("Keyboard", "Backspace", "", nil)
		keyboard.queue.Publish("buffer", artifact)
	case 13: // Enter key
		artifact := data.New("Keyboard", "Enter", "", nil)
		keyboard.queue.Publish("buffer", artifact)
	case 27: // Escape key to return to Normal Mode
		artifact := data.New("Keyboard", "NormalMode", "", nil)
		keyboard.queue.Publish("mode_change", artifact)
	default:
		if b >= 32 && b <= 126 {
			// Publish an insert character event
			artifact := data.New("Keyboard", "InsertChar", "", []byte{b})
			keyboard.queue.Publish("buffer", artifact)
		}
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

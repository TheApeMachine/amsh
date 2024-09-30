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
		mode:  NormalMode,
	}

	// Subscribe to mode_change to update the current mode.
	modeSub := queue.Subscribe("mode_change")
	go func() {
		for artifact := range modeSub {
			if role, err = artifact.Role(); err != nil {
				errnie.Error(err)
			}

			switch role {
			case "NormalMode":
				keyboard.mode = NormalMode
			case "InsertMode":
				keyboard.mode = InsertMode
			case "CommandMode":
				keyboard.mode = CommandMode
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
		keyboard.queue.Publish("chat", data.New("Keyboard", "ToggleChat", "", nil))
		return
	}

	if b[0] == 27 { // Escape character
		if keyboard.mode == NormalMode {
			// Potential escape sequence
			b2 := make([]byte, 2)
			n2, err := os.Stdin.Read(b2)
			if err != nil || n2 != 2 {
				return
			}
			if b2[0] == 91 { // '[' character in escape sequence
				switch b2[1] {
				case 'A':
					keyboard.queue.Publish("cursor", data.New("Keyboard", "MoveUp", "", nil))
				case 'B':
					keyboard.queue.Publish("cursor", data.New("Keyboard", "MoveDown", "", nil))
				case 'C':
					keyboard.queue.Publish("cursor", data.New("Keyboard", "MoveForward", "", nil))
				case 'D':
					keyboard.queue.Publish("cursor", data.New("Keyboard", "MoveBackward", "", nil))
				}
			}
		} else {
			// In Insert or Command Mode, ESC switches to Normal Mode
			if keyboard.mode == InsertMode || keyboard.mode == CommandMode {
				keyboard.queue.Publish("mode_change", data.New("Keyboard", "NormalMode", "", nil))
			}
		}
		return
	}

	// Process based on current mode
	switch keyboard.mode {
	case NormalMode:
		keyboard.handleNormalMode(b[0])
	case InsertMode:
		keyboard.handleInsertMode(b[0])
	case CommandMode:
		keyboard.queue.Publish("command_input", data.New("Keyboard", "CommandInput", "", []byte{b[0]}))
	}
}

// handleNormalMode processes input in Normal Mode.
func (keyboard *Keyboard) handleNormalMode(b byte) {
	switch b {
	case 'i':
		// Publish an event to switch to Insert Mode
		keyboard.queue.Publish("mode_change", data.New("Keyboard", "InsertMode", "", nil))
	case 'h':
		// Publish a cursor move event
		keyboard.queue.Publish("cursor", data.New("Keyboard", "MoveBackward", "", nil))
	case 'j':
		keyboard.queue.Publish("cursor", data.New("Keyboard", "MoveDown", "", nil))
	case 'k':
		keyboard.queue.Publish("cursor", data.New("Keyboard", "MoveUp", "", nil))
	case 'l':
		keyboard.queue.Publish("cursor", data.New("Keyboard", "MoveForward", "", nil))
	case ':':
		// Publish an event to switch to Command Mode
		keyboard.queue.Publish("mode_change", data.New("Keyboard", "CommandMode", "", nil))
	case 'q':
		// Publish a quit event
		keyboard.queue.Publish("app", data.New("Keyboard", "Quit", "", nil))
	case 'x':
		// Publish a delete character event
		keyboard.queue.Publish("buffer", data.New("Keyboard", "DeleteChar", "", nil))
	}
}

// handleInsertMode processes input in Insert Mode.
func (keyboard *Keyboard) handleInsertMode(b byte) {
	switch b {
	case 127: // Backspace key
		keyboard.queue.Publish("buffer", data.New("Keyboard", "Backspace", "", nil))
	case 13: // Enter key
		keyboard.queue.Publish("buffer", data.New("Keyboard", "Enter", "", nil))
	case 27: // Escape key to return to Normal Mode
		keyboard.queue.Publish("mode_change", data.New("Keyboard", "NormalMode", "", nil))
	default:
		if b >= 32 && b <= 126 {
			// Publish an insert character event
			keyboard.queue.Publish("buffer", data.New("Keyboard", "InsertChar", "", []byte{b}))
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
		keyboard.queue.Publish("chat", data.New("Keyboard", "SendMessage", "", []byte(message)))
	case 127: // Backspace
		if len(keyboard.chatInputBuffer) > 0 {
			keyboard.chatInputBuffer = keyboard.chatInputBuffer[:len(keyboard.chatInputBuffer)-1]
			// Update chat input display
			keyboard.queue.Publish("chat", data.New("Keyboard", "UpdateChatInput", "", []byte(keyboard.chatInputBuffer)))
		}
	default:
		// Append character to input buffer
		keyboard.chatInputBuffer += string(b)
		// Update chat input display
		keyboard.queue.Publish("chat", data.New("Keyboard", "UpdateChatInput", "", []byte(keyboard.chatInputBuffer)))
	}
}
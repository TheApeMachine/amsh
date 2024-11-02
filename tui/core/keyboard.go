package core

import (
	"errors"
	"io"
	"log"
	"os"

	"github.com/muesli/cancelreader"
	"github.com/theapemachine/amsh/errnie"
)

type CmdType uint

const (
	CmdTypeNone CmdType = iota
	CmdTypeQuit
	CmdTypeModeNormal
	CmdTypeModeInsert
	CmdTypeModeVisual
	CmdMoveLeft
	CmdMoveRight
	CmdMoveUp
	CmdMoveDown
	CmdDeleteChar
	CmdBackspace
	CmdNewLineBelow
	CmdNewLineAbove
	CmdTeleportMode
	CmdTeleportInput
	CmdEnter
	CmdBrowserToggle
	CmdBrowserUp
	CmdBrowserDown
	CmdBrowserEnter
	CmdBrowserBack
)

type KeyMsg struct {
	Key rune
	Cmd CmdType
}

type Keyboard struct {
	reader cancelreader.CancelReader
	mode   Mode
}

func NewKeyboard() *Keyboard {
	return &Keyboard{
		reader: errnie.SafeMust(func() (cancelreader.CancelReader, error) {
			return cancelreader.NewReader(os.Stdin)
		}),
		mode: ModeNormal,
	}
}

func (keyboard *Keyboard) Pipe() chan KeyMsg {
	msgChan := make(chan KeyMsg)
	buffer := make([]byte, 3)

	go func() {
		defer close(msgChan)

		for {
			n := errnie.SafeMust(func() (int, error) {
				return keyboard.Read(buffer)
			})

			if n > 0 {
				// Handle escape sequences (arrow keys)
				if buffer[0] == 27 && n >= 3 && buffer[1] == 91 {
					key := buffer[2]
					if msg := keyboard.handleArrowKey(rune(key)); msg != nil {
						msgChan <- *msg
					}
				} else {
					// Regular key handling
					if msg := keyboard.matchCmd(rune(buffer[0])); msg != nil {
						msgChan <- *msg
					}
				}
			}
		}
	}()

	return msgChan
}

func (keyboard *Keyboard) handleArrowKey(key rune) *KeyMsg {
	var cmd CmdType
	switch key {
	case 65: // Up arrow
		cmd = CmdMoveUp
	case 66: // Down arrow
		cmd = CmdMoveDown
	case 67: // Right arrow
		cmd = CmdMoveRight
	case 68: // Left arrow
		cmd = CmdMoveLeft
	default:
		return nil
	}
	log.Printf("Arrow key pressed: %d, sending command: %v", key, cmd)
	return &KeyMsg{Key: key, Cmd: cmd}
}

func (keyboard *Keyboard) matchCmd(key rune) *KeyMsg {
	// Handle special keys
	switch key {
	case 27: // ESC key
		keyboard.mode = ModeNormal
		return &KeyMsg{Key: key, Cmd: CmdTypeModeNormal}
	case 13: // Enter key
		return &KeyMsg{Key: key, Cmd: CmdEnter}
	case 127, 8: // Backspace
		return &KeyMsg{Key: key, Cmd: CmdBackspace}
	case 32: // Space key
		if keyboard.mode == ModeInsert {
			return &KeyMsg{Key: key, Cmd: CmdTypeNone} // Allow space in insert mode
		}
	case 't':
		if keyboard.mode == ModeNormal {
			return &KeyMsg{Key: key, Cmd: CmdTeleportMode}
		}
	}

	// For insert mode, send all regular characters directly
	if keyboard.mode == ModeInsert && key >= 32 && key < 127 {
		return &KeyMsg{Key: key, Cmd: CmdTypeNone}
	}

	// Handle normal mode commands
	if keyboard.mode == ModeNormal {
		switch key {
		case 'q':
			return &KeyMsg{Key: key, Cmd: CmdTypeQuit}
		case 'i':
			keyboard.mode = ModeInsert
			return &KeyMsg{Key: key, Cmd: CmdTypeModeInsert}
		case 'x':
			return &KeyMsg{Key: key, Cmd: CmdDeleteChar}
		case 'b':
			return &KeyMsg{Key: key, Cmd: CmdBrowserToggle}
		}
	}

	return nil
}

func (keyboard *Keyboard) Read(p []byte) (n int, err error) {
	_, err = keyboard.reader.Read(p[:])

	if errors.Is(err, cancelreader.ErrCanceled) {
		return 0, io.EOF
	}

	return len(p), err
}

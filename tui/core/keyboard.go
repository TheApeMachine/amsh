package core

import (
	"errors"
	"io"
	"os"
	"strings"

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
)

var cmdMap = map[Mode]map[string]CmdType{
	ModeNormal: {
		"q": CmdTypeQuit,
		"i": CmdTypeModeInsert,
		"v": CmdTypeModeVisual,
	},
	ModeInsert: {
		"esc": CmdTypeModeNormal,
	},
	ModeVisual: {
		"esc": CmdTypeModeNormal,
	},
}

type KeyMsg struct {
	Key rune
	Cmd CmdType
}

type Keyboard struct {
	reader cancelreader.CancelReader
	mode   Mode
	buf    string
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
	buffer := make([]byte, 1)

	go func() {
		defer close(msgChan)

		// Read from the keyboard
		for {
			n := errnie.SafeMust(func() (int, error) {
				return keyboard.Read(buffer)
			})

			if n > 0 {
				key := rune(buffer[0])
				if msg := keyboard.matchCmd(key); msg != nil {
					msgChan <- *msg
					keyboard.buf = ""
				}
			}
		}
	}()

	return msgChan
}

/*
matchCmd will be called any time a key is pressed, and one of the
following conditions are met:

1. The current key, combined with the previous keys, has not eliminated all potential matches.
2. The current key is pressed within the timeout period after the previous key.
3. The current key is the start of a new sequence, after the previous sequence resolved.
*/
func (keyboard *Keyboard) matchCmd(key rune) *KeyMsg {
	keyboard.buf += string(key)

	// First we check if we have a partial match in the cmdMap.
	for key := range cmdMap[keyboard.mode] {
		// If we have a partial match, we need to wait for a potential full match.
		if strings.HasPrefix(keyboard.buf, key) {
			return nil
		}
	}

	// If we have a full match, we can return the command.
	if cmd, exists := cmdMap[keyboard.mode][keyboard.buf]; exists {
		return &KeyMsg{Key: key, Cmd: cmd}
	}

	// If we don't have a full match, and we don't have a partial match,
	// we return the key as a normal key.
	return &KeyMsg{Key: key, Cmd: CmdTypeNone}
}

func (keyboard *Keyboard) Read(p []byte) (n int, err error) {
	_, err = keyboard.reader.Read(p[:])

	if errors.Is(err, cancelreader.ErrCanceled) {
		return 0, io.EOF
	}

	return len(p), err
}

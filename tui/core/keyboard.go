package core

import "time"

/*
Keyboard is a wrapper around the keyboard file descriptor.
It is designed to perform keyboard-specific operations and
prioritize raw performance.
*/
type Keyboard struct {
	sequence  []byte
	index     int
	lastPress time.Time
	timeout   time.Duration
	key       byte
}

/*
NewKeyboard creates a new Keyboard.
*/
func NewKeyboard() *Keyboard {
	return &Keyboard{
		sequence:  []byte{},
		index:     0,
		lastPress: time.Now(),
		timeout:   0,
		key:       0,
	}
}

func (keyboard *Keyboard) Read(p []byte) (n int, err error) {
	if time.Since(keyboard.lastPress) > keyboard.timeout {
		keyboard.index = 0
	}

	keyboard.lastPress = time.Now()

	if keyboard.key == keyboard.sequence[keyboard.index] {
		keyboard.index++
		if keyboard.index == len(keyboard.sequence) {
			keyboard.index = 0
			return 1, nil
		}
	} else {
		keyboard.index = 0
	}
	return 0, nil
}

func (keyboard *Keyboard) Write(p []byte) (n int, err error) {
	keyboard.sequence = append(keyboard.sequence, p...)
	return len(p), nil
}

func (keyboard *Keyboard) Close() error {
	return nil
}

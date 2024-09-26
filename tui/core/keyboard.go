package core

import (
    "io"
    "os"
    "time"
)

// Keyboard wraps an io.ReadWriteCloser (os.Stdin) and handles key mappings.
type Keyboard struct {
    input   io.ReadWriteCloser
    Keymap  map[byte]func()
    timeout time.Duration
}

// NewKeyboard creates a new Keyboard instance with os.Stdin as the input source.
func NewKeyboard() *Keyboard {
    return &Keyboard{
        input:   os.Stdin,
        Keymap:  make(map[byte]func()),
        timeout: 500 * time.Millisecond, // Example timeout
    }
}

// Read reads from the input, handles key mappings, and triggers actions.
func (k *Keyboard) Read(p []byte) (n int, err error) {
    n, err = k.input.Read(p)
    if n > 0 {
        key := p[0]
        if action, exists := k.Keymap[key]; exists {
            action()
        }
    }
    return n, err
}

// Write is implemented to satisfy the io.Writer interface but can be customized.
func (k *Keyboard) Write(p []byte) (n int, err error) {
    return os.Stdout.Write(p)
}

// Close closes the input source.
func (k *Keyboard) Close() error {
    return k.input.Close()
}

package core

import (
	"bytes"
	"os"
	"syscall"
)

/*
Terminal is a wrapper around the terminal file descriptor.
It is designed to perform terminal-specific operations and
prioritize raw performance.
*/
type Terminal struct {
	buf *bytes.Buffer
	out int
}

/*
NewTerminal creates a new Terminal and initializes the file descriptor.
*/
func NewTerminal() *Terminal {
	return &Terminal{
		out: int(os.Stdout.Fd()),
	}
}

func (terminal *Terminal) Read(p []byte) (n int, err error) {
	return syscall.Write(terminal.out, p)
}

func (terminal *Terminal) Write(p []byte) (n int, err error) {
	return terminal.buf.Write(p)
}

func (terminal *Terminal) Close() error {
	return nil
}

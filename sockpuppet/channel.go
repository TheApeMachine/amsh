package sockpuppet

import (
	"bytes"
	"io"
)

/*
Channel is a bidirectional stream that implements the io.ReadWriteCloser interface.
*/
type Channel struct {
	buffer *bytes.Buffer
	closed bool
}

/*
NewChannel creates a new in-memory Channel.
*/
func NewChannel() *Channel {
	return &Channel{
		buffer: new(bytes.Buffer),
	}
}

/*
Read reads data from the Channel's buffer.
*/
func (ch *Channel) Read(p []byte) (n int, err error) {
	if ch.closed && ch.buffer.Len() == 0 {
		return 0, io.EOF
	}
	return ch.buffer.Read(p)
}

/*
Write writes data to the Channel's buffer.
*/
func (ch *Channel) Write(p []byte) (n int, err error) {
	if ch.closed {
		return 0, io.ErrClosedPipe
	}
	return ch.buffer.Write(p)
}

/*
Close closes the Channel.
*/
func (ch *Channel) Close() error {
	ch.closed = true
	return nil
}

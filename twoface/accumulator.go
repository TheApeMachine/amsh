package twoface

import (
	"bytes"
	"io"

	"github.com/theapemachine/amsh/data"
	"github.com/theapemachine/errnie"
)

/*
Accumulator is a wrapper around a stream of events that implements io.ReadWriteCloser.
It provides a simple way to pipe data between components.
*/
type Accumulator struct {
	pr *io.PipeReader
	pw *io.PipeWriter

	buffer *bytes.Buffer
}

func NewAccumulator() *Accumulator {
	errnie.Trace("%s", "Accumulator.NewAccumulator", "new")

	pr, pw := io.Pipe()
	return &Accumulator{
		pr:     pr,
		pw:     pw,
		buffer: bytes.NewBuffer(nil),
	}
}

/*
Read called the first time will first passthrough all the data in the pipe, and return EOF.
This allows io.Copy to break off the passthrough data first. A secondary call will return
the accumulated data. After that, it will return EOF.
Any further calls will return EOF.
*/
func (accumulator *Accumulator) Read(p []byte) (n int, err error) {
	// Only try to unmarshal if we have data
	if len(p) > 0 {
		artifact := data.Empty()
		artifact.Unmarshal(p)
		errnie.Trace("%s", "artifact.Payload", artifact.Peek("payload"))
	}

	if accumulator.pr != nil {
		n = errnie.SafeMust(func() (int, error) {
			return accumulator.pr.Read(p)
		})

		accumulator.pr = nil
		return n, io.EOF
	}

	// Return EOF if buffer is empty
	if accumulator.buffer.Len() == 0 {
		return 0, io.EOF
	}

	// Read only up to len(p) bytes
	n = copy(p, accumulator.buffer.Bytes())

	// If we couldn't read everything, only advance the buffer by what we did read
	if n < accumulator.buffer.Len() {
		accumulator.buffer.Next(n)
		return n, nil
	}

	// If we read everything, return EOF
	return n, io.EOF
}

func (accumulator *Accumulator) Write(p []byte) (n int, err error) {
	// Only try to unmarshal if we have data
	if len(p) > 0 {
		artifact := data.Empty()
		artifact.Unmarshal(p)
		errnie.Trace("%s", "artifact.Payload", artifact.Peek("payload"))
	}

	// Write in a goroutine to prevent blocking
	go func() {
		// Write to both the pipe (passthrough) and the buffer (accumulator).
		sink := io.MultiWriter(accumulator.pw, accumulator.buffer)
		sink.Write(p)
	}()
	return len(p), nil
}

func (accumulator *Accumulator) Close() error {
	errnie.Trace("%s", "Accumulator.Close", "close")

	if err := accumulator.pw.Close(); err != nil {
		return err
	}
	return accumulator.pr.Close()
}

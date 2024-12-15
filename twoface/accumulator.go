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
    // First try to read from pipe
    if accumulator.pr != nil {
        n, err = accumulator.pr.Read(p)
        if err != nil && err != io.EOF {
            return n, err
        }

        // Only try to unmarshal if we have valid data
        if n > 0 {
            artifact := data.Empty()
            if err := artifact.Unmarshal(p[:n]); err != nil {
                errnie.Error(err)
                // Continue even if unmarshal fails - the raw data will still be returned
            }
        }

        return n, err
    }

    // If pipe is closed, try to read from buffer
    if accumulator.buffer != nil && accumulator.buffer.Len() > 0 {
        return accumulator.buffer.Read(p)
    }

    return 0, io.EOF
}

func (accumulator *Accumulator) Write(p []byte) (n int, err error) {
	// Only try to unmarshal if we have data
	if len(p) > 0 {
		artifact := data.Empty()
		artifact.Unmarshal(p)
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
	if err := accumulator.pw.Close(); err != nil {
		return err
	}
	return accumulator.pr.Close()
}

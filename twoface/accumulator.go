package twoface

import (
	"io"

	"github.com/theapemachine/amsh/data"
	"github.com/theapemachine/errnie"
)

type Buffer struct {
	pr *io.PipeReader
	pw *io.PipeWriter
}

func NewBuffer() *Buffer {
	pr, pw := io.Pipe()

	return &Buffer{
		pr: pr,
		pw: pw,
	}
}

func (buffer *Buffer) Read(p []byte) (n int, err error) {
	if buffer.pr == nil {
		return 0, io.EOF
	}

	buf := make([]byte, 1024)
	n, err = buffer.pr.Read(buf)

	errnie.Debug("buffer.Read", "buf", string(buf[:n]))

	artifact := data.Empty()
	errnie.Error(artifact.Unmarshal(buf[:n]))

	return copy(p, []byte(artifact.Peek("payload"))), nil
}

func (buffer *Buffer) Write(p []byte) (n int, err error) {
	return buffer.pw.Write(p)
}

/*
Accumulator is a wrapper around a stream of events that implements io.ReadWriteCloser.
It provides a simple way to pipe data between components.
*/
type Accumulator struct {
	pr *io.PipeReader
	pw *io.PipeWriter

	buffer *Buffer
}

func NewAccumulator() *Accumulator {
	pr, pw := io.Pipe()

	return &Accumulator{
		pr:     pr,
		pw:     pw,
		buffer: NewBuffer(),
	}
}

func (accumulator *Accumulator) Buffer() *Buffer {
	return accumulator.buffer
}

/*
Read reads from the accumulator.
*/
func (accumulator *Accumulator) Read(p []byte) (n int, err error) {
	if accumulator.pr != nil {
		if n = errnie.SafeMust(func() (int, error) {
			return accumulator.pr.Read(p)
		}); n == 0 {
			accumulator.pr.Close()
			return 0, io.EOF
		}

		errnie.Debug("accumulator.Read", "p", string(p[:n]))

		// Only try to unmarshal if we have valid data
		if n > 0 {
			artifact := data.Empty()
			errnie.Error(artifact.Unmarshal(p[:n]))
		}

		return
	}

	return
}

func (accumulator *Accumulator) Write(p []byte) (n int, err error) {
	go func() {
		sink := io.MultiWriter(accumulator.pw, accumulator.buffer)
		n, err = sink.Write(p)

		errnie.Debug("accumulator.Write", "p", string(p[:n]))
	}()

	return len(p), nil
}

func (accumulator *Accumulator) Close() (err error) {
	if err = errnie.Error(accumulator.pw.Close()); err != nil {
		return
	}

	if err = errnie.Error(accumulator.pr.Close()); err != nil {
		return
	}

	return
}

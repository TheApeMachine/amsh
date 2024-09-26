package ai

import (
	"bufio"
	"io"
)

type IO struct {
	reader *bufio.Reader
	writer *bufio.Writer
}

func NewIO(r io.Reader, w io.Writer) *IO {
	return &IO{
		reader: bufio.NewReader(r),
		writer: bufio.NewWriter(w),
	}
}

/*
Read reads the input from the user.
*/
func (io *IO) Read(p []byte) (n int, err error) {
	return io.reader.Read(p)
}

/*
Write writes the output to the user.
*/
func (io *IO) Write(p []byte) (n int, err error) {
	if n, err = io.writer.Write(p); err != nil {
		return n, err
	}

	return n, io.writer.Flush()
}

/*
Close closes the IO.
*/
func (io *IO) Close() error {
	if err := io.writer.Flush(); err != nil {
		return err
	}
	return nil
}

package data

import (
	"sync"

	"github.com/theapemachine/amsh/errnie"
)

/*
bufpool is a buffer pool for the Artifact.
*/
var bufpool = sync.Pool{
	New: func() any {
		return make([]byte, 0, 1024)
	},
}

/*
Read implements the io.Reader interface for the Artifact.
It marshals the entire artifact into the provided byte slice.
*/
func (artifact *Artifact) Read(p []byte) (n int, err error) {
	var buf []byte

	// Marshal the artifact into bytes.
	if buf, err = artifact.Message().Marshal(); err != nil {
		return 0, err
	}

	// If the provided byte slice is too small, grow it.
	if len(p) < len(buf) {
		// Grow the slice to fit the marshaled data.
		p = make([]byte, len(buf))
	}

	// Copy the marshaled bytes into the provided byte slice.
	return copy(p, buf), nil
}

/*
Write implements the io.Writer interface for the Artifact.
It writes the entire artifact to the provided stream.
*/
func (artifact *Artifact) Write(p []byte) (n int, err error) {
	// Get a buffer from the pool.
	buf := bufpool.Get().([]byte)
	defer bufpool.Put(&buf)

	if buf, err = artifact.Payload(); err != nil {
		err = errnie.Error(err)
	}

	// Copy the provided byte slice into the buffer.
	artifact.SetPayload(append(buf, p...))
	return len(p), nil
}

/*
Close implements the io.Closer interface for the Artifact.
*/
func (artifact *Artifact) Close() error {
	// No-op for this example, but could be extended to manage resources.
	return nil
}

package data

import (
	"errors"
	"io"
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
	// Marshal the artifact.
	buf := artifact.Marshal()
	if buf == nil {
		return 0, errnie.Error(errors.New("failed to marshal artifact"))
	}

	copy(p, buf)

	return len(p), io.EOF
}

/*
Write implements the io.Writer interface for the Artifact.
It writes the entire artifact to the provided stream.
*/
func (artifact *Artifact) Write(p []byte) (n int, err error) {
	payload, err := artifact.Payload()

	if err != nil {
		return 0, errnie.Error(err)
	}

	payload = append(payload, p...)
	err = artifact.SetPayload(payload)

	if err != nil {
		return 0, errnie.Error(err)
	}

	return len(p), nil
}

/*
Close implements the io.Closer interface for the Artifact.
*/
func (artifact Artifact) Close() error {
	// No-op for this example, but could be extended to manage resources.
	return nil
}

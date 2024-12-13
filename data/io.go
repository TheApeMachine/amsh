package data

import (
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
	artifact.Marshal(p)
	return len(p), io.EOF
}

/*
Write implements the io.Writer interface for the Artifact.
It unmarshals the provided bytes into the current artifact.
*/
func (artifact *Artifact) Write(p []byte) (n int, err error) {
	artifact.Unmarshal(p)
	return len(p), nil
}

func (artifact *Artifact) Append(str string) error {
	payload, err := artifact.Payload()
	if err != nil {
		return errnie.Error(err)
	}

	buf := bufpool.Get().([]byte)
	defer bufpool.Put(buf)

	buf = append(buf[:0], payload...)
	buf = append(buf, str...)

	return errnie.Error(artifact.SetPayload(buf))
}

/*
Close implements the io.Closer interface for the Artifact.
*/
func (artifact Artifact) Close() error {
	// No-op for this example, but could be extended to manage resources.
	return nil
}

package data

import (
	"capnproto.org/go/capnp/v3"
	"github.com/theapemachine/amsh/errnie"
)

func (artifact *Artifact) Marshal(p []byte) {
	errnie.Trace("%s", "Artifact.Marshal", "marshal")

	var (
		buf []byte
		err error
	)

	if buf, err = artifact.Message().Marshal(); err != nil {
		// Log the error, we don't need to return here, because we
		// rely on the caller handling things if the buffer does not
		// contain the expected data.
		errnie.Error(err)
	}

	// Return the buffer in whatever state it may be.
	copy(p, buf)
}

func (artifact *Artifact) Unmarshal(buf []byte) {
	// Check if buffer is empty or too small
	if len(buf) == 0 {
		errnie.Trace("empty buffer, skipping unmarshal")
		return
	}

	var (
		msg    *capnp.Message
		artfct Artifact
		err    error
	)

	// Unmarshal is a bit of a misnomer in the world of Cap 'n Proto,
	// but they went with it anyway.
	if msg, err = capnp.Unmarshal(buf); err != nil {
		errnie.Error(err)
		return
	}

	// Read a Datagram instance from the message.
	if artfct, err = ReadRootArtifact(msg); err != nil {
		errnie.Error(err)
		return
	}

	// Overwrite the pointer to our empty instance with the one
	// pointing to our root Datagram.
	artifact = &artfct

	if payload := artifact.Peek("payload"); payload != "" {
		errnie.Trace("%s", "payload", payload)
	}
}

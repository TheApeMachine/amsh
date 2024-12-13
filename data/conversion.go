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
	var (
		msg    *capnp.Message
		artfct Artifact
		err    error
	)

	// Unmarshal is a bit of a misnomer in the world of Cap 'n Proto,
	// but they went with it anyway.
	if msg, err = capnp.Unmarshal(buf); errnie.Error(err) != nil {
		return
	}

	errnie.Raw(msg)

	// Read a Datagram instance from the message.
	if artfct, err = ReadRootArtifact(msg); errnie.Error(err) != nil {
		return
	}

	// Overwrite the pointer to our empty instance with the one
	// pointing to our root Datagram.
	artifact = &artfct

	errnie.Trace("%s", "payload", artifact.Peek("payload"))
}

package data

import (
	"capnproto.org/go/capnp/v3"
	"fmt"
	"github.com/theapemachine/errnie"
)

func (artifact *Artifact) Marshal(p []byte) {
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

func (artifact *Artifact) Unmarshal(buf []byte) error {
	var (
		msg    *capnp.Message
		artfct Artifact
		err    error
	)

	if len(buf) == 0 {
		return fmt.Errorf("empty buffer")
	}

	if msg, err = capnp.Unmarshal(buf); err != nil {
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}

	if msg == nil {
		return fmt.Errorf("nil message after unmarshal")
	}

	if artfct, err = ReadRootArtifact(msg); err != nil {
		return fmt.Errorf("failed to read root artifact: %w", err)
	}

	*artifact = artfct
	return nil
}

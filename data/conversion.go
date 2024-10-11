package data

import (
	"capnproto.org/go/capnp/v3"
	"github.com/theapemachine/amsh/errnie"
)

func (artifact Artifact) Marshal() []byte {
	var (
		buf []byte
		err error
	)

	if buf, err = artifact.Message().Marshal(); err != nil {
		return nil
	}

	return buf
}

func (artifact Artifact) Unmarshal(buf []byte) Artifact {
	var (
		msg *capnp.Message
		err error
	)

	if msg, err = capnp.Unmarshal(buf); errnie.Error(err) != nil {
		return Empty
	}

	if artifact, err = ReadRootArtifact(msg); errnie.Error(err) != nil {
		return Empty
	}

	return artifact
}

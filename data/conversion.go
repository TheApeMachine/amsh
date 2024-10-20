package data

import (
	"fmt"

	"capnproto.org/go/capnp/v3"
	"github.com/theapemachine/amsh/errnie"
)

func (artifact *Artifact) Marshal() []byte {
	var (
		buf []byte
		err error
	)

	if buf, err = artifact.Message().Marshal(); err != nil {
		errnie.Error(fmt.Errorf("marshal error: %w", err))
		return nil
	}

	return buf
}

func Unmarshal(buf []byte) *Artifact {
	var (
		msg *capnp.Message
		af  Artifact
		err error
	)

	if msg, err = capnp.Unmarshal(buf); errnie.Error(err) != nil {
		errnie.Error(fmt.Errorf("unmarshal error: %w", err))
		return nil
	}

	if errnie.Error(err) != nil {
		return nil
	}

	if af, err = ReadRootArtifact(msg); errnie.Error(err) != nil {
		return nil
	}

	return &af
}

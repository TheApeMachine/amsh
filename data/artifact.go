package data

import (
	"time"

	"capnproto.org/go/capnp/v3"
	"github.com/theapemachine/amsh/errnie"
)

const version = "0.0.1"

/*
Empty is an empty artifact.
*/
var Empty = &Artifact{}

/*
New creates a new artifact with the given origin, role, scope, and data.
*/
func New(origin, role, scope string, data []byte) *Artifact {
	var (
		seg      *capnp.Segment
		err      error
		artifact Artifact
	)

	if _, seg, err = capnp.NewMessage(capnp.SingleSegment(nil)); err != nil {
		return nil
	}

	if artifact, err = NewArtifact(seg); err != nil {
		return nil
	}

	errnie.Error(artifact.SetOrigin(origin))
	errnie.Error(artifact.SetRole(role))
	errnie.Error(artifact.SetScope(scope))
	errnie.Error(artifact.SetPayload(data))

	artifact.SetTimestamp(uint64(time.Now().UnixNano()))
	artifact.SetVersion(version)

	return &artifact
}

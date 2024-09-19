package data

import (
	"time"

	"capnproto.org/go/capnp/v3"
	"github.com/theapemachine/amsh/logger"
)

const version = "0.0.1"

/*
Empty is an empty artifact.
*/
var Empty = Artifact{}

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

	logger.Error("Error setting origin: %v", artifact.SetOrigin(origin))
	logger.Error("Error setting role: %v", artifact.SetRole(role))
	logger.Error("Error setting scope: %v", artifact.SetScope(scope))
	logger.Error("Error setting payload: %v", artifact.SetPayload(data))

	artifact.SetTimestamp(uint64(time.Now().UnixNano()))
	artifact.SetVersion(version)

	return &artifact
}

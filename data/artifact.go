package data

import (
	"capnproto.org/go/capnp/v3"
	"github.com/theapemachine/amsh/logger"
)

var Empty = Artifact{}

func New(origin, role, scope string, data []byte) Artifact {
	_, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		return Artifact{}
	}

	artifact, err := NewArtifact(seg)
	if err != nil {
		return Artifact{}
	}

	logger.Error("Error setting origin: %v", artifact.SetOrigin(origin))
	logger.Error("Error setting role: %v", artifact.SetRole(role))
	logger.Error("Error setting scope: %v", artifact.SetScope(scope))
	logger.Error("Error setting payload: %v", artifact.SetPayload(data))

	return artifact
}

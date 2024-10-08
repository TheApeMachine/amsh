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

func (artifact *Artifact) SetAttrs(attrs map[string]string) {
	// Create a new list of attributes
	attrList, err := NewAttribute_List(artifact.Segment(), int32(len(attrs)))
	if err != nil {
		errnie.Error(err)
		return
	}

	// Populate the attribute list
	i := 0
	for key, value := range attrs {
		attr := attrList.At(i)
		errnie.Error(attr.SetKey(key))
		errnie.Error(attr.SetValue(value))

		i++
	}

	// Set the attributes on the artifact
	errnie.Error(artifact.SetAttributes(attrList))
}
